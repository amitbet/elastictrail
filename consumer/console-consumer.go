package consumer

import (
	"elastictrail/common"
	"fmt"
)

// format can be kubernetes of simple
type ConsoleConsumer struct {
	Format string
}

func (consumer *ConsoleConsumer) Consume(line common.LogLine) error {
	//fmt.Printf("%v\n", line)
	//j, _ := json.Marshal(line.Content)
	//fmt.Printf(">>line:%v\n", string(j))
	var str string
	switch consumer.Format {
	case "kubernetes":
		str = line.GetField("@timestamp") + " " + line.GetField("kubernetes.container_name") + " NS: " + line.GetField("kubernetes.namespace_name") + " [" + line.GetField("level") + "] " + line.Message()
	case "simple":
		str = line.GetField("message")
	}
	fmt.Printf("##line: %s\n", str)

	return nil
}

func (consumer *ConsoleConsumer) Name() string {
	return "console-consumer"
}
