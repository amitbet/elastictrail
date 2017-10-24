package common

import (
	"fmt"
)

// logConsumer is an interface for consumers of the lines stream any log reader produces
type LogConsumer interface {
	Consume(line LogLine) error
	Name() string
	BatchDone()
}

type consumerInfo struct {
	LogConsumer
	lineChannel chan LogLine
}

type ConsumerDictionary struct {
	consumers []consumerInfo
}

func (dict *ConsumerDictionary) Distribute(line LogLine) {
	for _, consumerInf := range dict.consumers {
		consumerInf.lineChannel <- line
	}
}
func (dict *ConsumerDictionary) DistributeBatchDone() {
	for _, consumerInf := range dict.consumers {
		consumerInf.BatchDone()
	}
}

// RegisterConsumer creates a channel for log lines and runs the consumer.Consume function in go routine(s)
func (dict *ConsumerDictionary) RegisterConsumer(consumer LogConsumer) {
	ch := make(chan LogLine)
	//qch := make(chan bool)
	consumerInf := consumerInfo{LogConsumer: consumer, lineChannel: ch}
	go func() {
		for line := range ch {
			err := consumerInf.Consume(line)
			if err != nil {
				fmt.Printf("error while sending line to: %s, err: %s\n", consumerInf.Name(), err)
			}
		}
	}()
	dict.consumers = append(dict.consumers, consumerInf)
}
