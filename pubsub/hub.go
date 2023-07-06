package pubsub

import (
	Dict "hangdis/datastruct/dict"
	"sync"
)

type Hub struct {
	subs Dict.Dict
	mu   sync.RWMutex
}

func MakeHub() *Hub {
	return &Hub{
		subs: Dict.MakeConcurrent(3),
	}
}
