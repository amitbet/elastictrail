package common

// logConsumer is an interface for consumers of the lines stream any log reader produces
type LogConsumer interface {
	Consume(line LogLine) error
	Name() string
	BatchDone()
}
