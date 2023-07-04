package pubsub

import Dict "hangdis/datastruct/dict"

type Hub struct {
	subs Dict.Dict
}

func MakeHub() *Hub {
	return &Hub{
		subs: Dict.MakeConcurrent(3),
	}
}
