package consumer

import (
	"elastictrail/common"
	"fmt"
	"strconv"
)

type GrouperConsumer struct {
	groupper *AutoGroupper
	common.ConsumerDictionary
}

func NewGrouperConsumer() *GrouperConsumer {
	groupper := NewAutoGroupper()

	return &GrouperConsumer{groupper: groupper}
}

var lineCount int

func (consumer *GrouperConsumer) BatchDone() {
	// for _, group := range consumer.groupper.groups {
	// 	consumer.ConsumerDictionary.Distribute(group)
	// }
	consumer.ConsumerDictionary.DistributeBatchDone()
}

func (consumer *GrouperConsumer) Consume(line common.LogLine) error {
	//fmt.Printf("%v\n", line)
	//j, _ := json.Marshal(line.Content)
	//fmt.Printf(">>line:%v\n", string(j))

	//str := line.Time() + " " + line.GetField("kubernetes.container_name") + " NS: " + line.GetField("kubernetes.namespace_name") + " [" + line.Level() + "] " + line.Message()
	//fmt.Printf("##line:" + str)
	lineCount++
	message := line.GetField("message")
	if message == "" {
		message = line.GetField("log")
	}

	consumer.groupper.FindGroup(line)

	if lineCount%1000 == 0 {
		//some console output for visibility
		for _, gr := range consumer.groupper.groups {
			fmt.Printf("%v\n", gr.String())
		}
		fmt.Println("----------" + strconv.Itoa(lineCount))
	}
	return nil
}

func (consumer *GrouperConsumer) Name() string {
	return "grouper-consumer"
}
