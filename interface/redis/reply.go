package redis

type Reply interface {
	ToBytes() []byte
}

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}
