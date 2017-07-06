package reader

// Feeder is an interface for objects capable of returning a reader to a log stream
// log items should be json type objects (map[string]interface{})
type Feeder interface {
	Start()
	RegisterConsumer()
}
