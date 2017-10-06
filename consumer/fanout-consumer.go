package consumer

import (
	"elastictrail/common"
	"fmt"
)

//FanOutConsumer takes a logField and splits the line stream according to the values of this field
type FanOutConsumer struct {
	LogField        string
	ConsumerByValue map[string]*consumerInfo
	autoCreate      *AutoCreateConfig // whether to autocreate AutoGroupperConsumers for each of the field values
}

// AutoCreateConfig allows to auto-create consumers, the type of the sub-consumers can be: FanOut / Grouper / Console
type AutoCreateConfig struct {
	consumerType     string            // which consumer to create for each value
	fanOutLogField   string            // what log field to fan out next (if fanOut is the consumerType selected)
	fanOutAutoCreate *AutoCreateConfig // optional recursive field for nested fanOut configurations
}
type consumerInfo struct {
	common.LogConsumer
	lineChannel chan common.LogLine
}

func (fc *FanOutConsumer) Name() string {
	return "fanout-consumer"
}

func (fc *FanOutConsumer) RegisterConsumer(value string, consumer common.LogConsumer) {
	if fc.ConsumerByValue == nil {
		fc.ConsumerByValue = map[string]*consumerInfo{}
	}

	ch := make(chan common.LogLine)
	consumerInf := consumerInfo{LogConsumer: consumer, lineChannel: ch}
	go func() {
		for line := range ch {
			err := consumerInf.Consume(line)
			if err != nil {
				fmt.Printf("error while sending line to: %s, err: %s\n", consumerInf.Name(), err)
			}
		}
	}()
	fc.ConsumerByValue[value] = &consumerInf
}

func (fc *FanOutConsumer) BatchDone() {
	for _, child := range fc.ConsumerByValue {
		child.BatchDone()
	}
}

func (fc *FanOutConsumer) Consume(line common.LogLine) error {
	val := line.GetField(fc.LogField)
	consumer := fc.ConsumerByValue[val]

	//handle autocreate case
	if consumer == nil && fc.autoCreate != nil {
		var c common.LogConsumer
		switch fc.autoCreate.consumerType {
		case "Grouper":
			c = NewGrouperConsumer()
		case "FanOut":
			c = &FanOutConsumer{LogField: fc.autoCreate.fanOutLogField, autoCreate: fc.autoCreate.fanOutAutoCreate}
		case "Console":
			c = &ConsoleConsumer{Format: "simple"}
		}
		fc.RegisterConsumer(val, c)
		consumer = fc.ConsumerByValue[val]
	}

	//send the log line to be processed
	if consumer != nil {
		consumer.lineChannel <- line
	}
	return nil
}
