package reader

import (
	"elastictrail/consumer"
	"testing"
)

//import "elastictrail/consumer"

// func TestEsFeeder(t *testing.T) {
// 	reader := EsReader{eshost: "elasticsearch-dev-1991374869.us-west-2.elb.amazonaws.com",
// 		esport:      "9200",
// 		indexPrefix: "logstash-",
// 		useSSL:      false,
// 		fieldNames: []string{"message",
// 			"log",
// 			"@timestamp",
// 			"timestamp",
// 			"level",
// 			"kubernetes.container_name",
// 			"kubernetes.namespace_name",
// 		},
// 		filters: map[string]string{
// 			"kubernetes.namespace_name": "srf-green-public",
// 		},
// 	}
// 	reader.RegisterConsumer(&consumer.ConsoleConsumer{})
// 	reader.Start()
// }

func TestAutoGrouper(t *testing.T) {
	reader := EsReader{
		//eshost: "elasticsearch-dev-1991374869.us-west-2.elb.amazonaws.com",
		eshost:      "34.209.71.120",
		esport:      "9200",
		indexPrefix: "logstash-",
		useSSL:      false,
		fieldNames: []string{"message",
			"log",
			"@timestamp",
			"timestamp",
			"level",
			"kubernetes.container_name",
			"kubernetes.namespace_name",
		},
		filters: map[string]string{
			"kubernetes.namespace_name": "srf-integ-ga",
		},
	}
	reader.RegisterConsumer(consumer.NewGrouperConsumer())
	reader.Start()
}
