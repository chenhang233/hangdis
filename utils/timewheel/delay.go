package timewheel

import (
	"time"
)

var tw = New(time.Second, 360)

func init() {
	tw.Start()
}

func At(at time.Time, key string, job func()) {
	tw.AddJob(at.Sub(time.Now()), key, job)
}

func Cancel(key string) {
	tw.RemoveJob(key)
}
