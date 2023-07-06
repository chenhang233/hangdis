package database

import (
	"hangdis/pubsub"
)

func init() {
	RegisterPubSubCommand("SUBSCRIBE", pubsub.Subscribe, flagWrite)
	RegisterPubSubCommand("UNSUBSCRIBE", pubsub.UnSubscribe, flagWrite)
	RegisterPubSubCommand("PUBLISH", pubsub.Publish, flagWrite)
}
