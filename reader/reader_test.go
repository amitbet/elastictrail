package reader

import (
	"elastictrail/consumer"
	"encoding/json"
	"io/ioutil"
	"testing"
	"vncproxy/logger"
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

func TestPromExporterLineCounter(t *testing.T) {
	reader := EsPoller{
		//eshost: "elasticsearch-dev-1991374869.us-west-2.elb.amazonaws.com",
		ESHost:      "localhost",
		ESPort:      "9200",
		IndexPrefix: "logstash-",
		UseSSL:      false,
		fieldNames: []string{"message",
			"log",
			"@timestamp",
			"timestamp",
			"level",
			"kubernetes.container_name",
			"kubernetes.namespace_name",
		},
		PollintIntervalSecs: 2,
		QueryString:         "level: (debug OR info)",
		TimeWindowSecs:      5,
	}
	//reader.RegisterConsumer(&consumer.ConsoleConsumer{Format: "kubernetes"})
	reader.RegisterConsumer(&consumer.PrometheusConsumer{Format: "kubernetes", GaugeType: consumer.GaugeTypeCountMatchesInTimespan, RegexStr: ".", GaugeName: "testLogCounterGauge"})
	reader.Start()
}

func TestPromExporterCapturingRegex(t *testing.T) {
	reader := EsPoller{
		//eshost: "elasticsearch-dev-1991374869.us-west-2.elb.amazonaws.com",
		ESHost:      "localhost",
		ESPort:      "9200",
		IndexPrefix: "logstash-",
		UseSSL:      false,
		fieldNames: []string{"message",
			"log",
			"@timestamp",
			"timestamp",
			"level",
			"kubernetes.container_name",
			"kubernetes.namespace_name",
		},
		PollintIntervalSecs: 5,
		QueryString:         "level: debug AND Redistimer AND kubernetes.container_name: \"browserlab-container\" AND kubernetes.namespace_name: \"srf-integ-ga\"",
		TimeWindowSecs:      5,
	}
	//reader.RegisterConsumer(&consumer.ConsoleConsumer{Format: "kubernetes"})
	reader.RegisterConsumer(&consumer.PrometheusConsumer{Format: "kubernetes", GaugeType: consumer.GaugeTypeExportMatchAsFloat, RegexStr: "_onInterval: completed at time: (\\d*)", GaugeName: "SRF_LastRedisTimerLogMonitor"})
	reader.Start()
}

func TestPollerJsonConfig(t *testing.T) {
	// conf1 := LogExporterConfig{
	// 	PollingConfigs: []PollingConfig{
	// 		PollingConfig{
	// 			PollintIntervalSecs: 5,
	// 			TimeWindowSecs:      5,
	// 			Query:               "lable:debug",
	// 			GaugeConfigs: []GaugeConfig{
	// 				GaugeConfig{
	// 					RegexStr:  "lable:debug",
	// 					GaugeType: 0,
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// b, _ := json.Marshal(conf1)
	// fmt.Println(string(b))
	jsonFile := "/Users/amitbet/Dropbox/go/src/elastictrail/reader/myConfig.json"
	buf, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		logger.Error("error reading file: ", jsonFile, err)
		return
	}
	json2 := string(buf)

	//json1 := "{\"pollingConfigurations\":[{\"interval\":15,\"timeWindow\":15,\"query\":\"level:(debug)\",\"gaugeConfigurations\":[{\"regex\":\"jobmanager\",\"type\":0, \"name\":\"srf_debug_counter_jobmanager\"},{\"regex\":\"testmanager\",\"type\":0, \"name\":\"srf_debug_counter_testmanager\"}]}]}"
	m := LogExporterConfig{}

	json.Unmarshal([]byte(json2), &m)
	//serialized, _ := json.Marshal(m)
	//fmt.Printf("here! %s", serialized)

	m.Run()
}
