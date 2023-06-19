package timewheel

import (
	"container/list"
	"hangdis/utils/logs"
	"time"
)

type Task struct {
	delay  time.Duration
	circle int
	key    string
	job    func()
}

type Location struct {
	slot    int
	element *list.Element
}

type TimeWheel struct {
	ticker         *time.Ticker
	interval       time.Duration
	tasks          map[string]*Location
	slots          []*list.List
	slotNum        int
	currentPos     int
	removeTaskChan chan string
	addTaskChan    chan Task
	stopChan       chan bool
}

func New(interval time.Duration, slotNum int) *TimeWheel {
	if interval <= 0 || slotNum <= 0 {
		return nil
	}
	t := &TimeWheel{
		interval:       interval,
		tasks:          make(map[string]*Location),
		slots:          make([]*list.List, slotNum),
		slotNum:        slotNum,
		currentPos:     0,
		removeTaskChan: make(chan string),
		addTaskChan:    make(chan Task),
		stopChan:       make(chan bool),
	}
	t.initSlots()
	return t
}

func (t *TimeWheel) initSlots() {
	for i := 0; i < t.slotNum; i++ {
		t.slots[i] = list.New()
	}
}

func (t *TimeWheel) Start() {
	t.ticker = time.NewTicker(t.interval)
	go t.start()
}

func (t *TimeWheel) start() {
	for {
		select {
		case <-t.ticker.C:
			t.tickHandler()
		case task := <-t.addTaskChan:
			t.addTask(&task)
		case key := <-t.removeTaskChan:
			t.removeTask(key)
		case <-t.stopChan:
			t.ticker.Stop()
			return
		}
	}
}

func (t *TimeWheel) tickHandler() {
	if t.currentPos == t.slotNum-1 {
		t.currentPos = 0
	} else {
		t.currentPos++
	}
	for i := 0; i < t.slotNum; i++ {
		go t.scanAndRunTask(t.slots[i])
	}
}

func (t *TimeWheel) scanAndRunTask(l *list.List) {
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logs.LOG.Error.Println(err)
				}
			}()
			job := task.job
			job()
		}()
		next := e.Next()
		l.Remove(e)
		delete(t.tasks, task.key)
		e = next
	}
}

func (t *TimeWheel) addTask(task *Task) {
	if task.key == "" {
		return
	}
	pos, c := t.getPositionAndCircle(task.delay)
	task.circle = c
	e := t.slots[pos].PushBack(task)
	location := &Location{
		slot:    pos,
		element: e,
	}
	if _, ok := t.tasks[task.key]; ok {
		t.removeTask(task.key)
	}
	t.tasks[task.key] = location
}

func (t *TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	delaySeconds := int(d.Seconds())
	intervalSeconds := int(t.interval.Seconds())
	circle = delaySeconds / intervalSeconds
	pos = (t.currentPos + delaySeconds/intervalSeconds) % t.slotNum
	return
}
func (t *TimeWheel) removeTask(key string) {
	location, ok := t.tasks[key]
	if !ok {
		return
	}
	t.slots[location.slot].Remove(location.element)
	delete(t.tasks, key)
}

func (t *TimeWheel) Stop() {
	t.stopChan <- true
}

func (t *TimeWheel) AddJob(delay time.Duration, key string, job func()) {
	if delay < 0 {
		return
	}
	t.addTaskChan <- Task{
		delay: delay,
		job:   job,
		key:   key,
	}
}

func (t *TimeWheel) RemoveJob(key string) {
	t.removeTaskChan <- key
}
