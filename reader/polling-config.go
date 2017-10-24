package reader

import (
	"elastictrail/consumer"
	"net/http"
	"vncproxy/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type GaugeConfig struct {
	RegexStr  string             `json:"regex"`
	GaugeType consumer.GaugeType `json:"type"`
	GaugeName string             `json:"name"`
	GaugeDesc string             `json:"desc"`
}

type PollingConfig struct {
	PollintIntervalSecs int           `json:"interval"`
	TimeWindowSecs      int           `json:"timeWindow"`
	Query               string        `json:"query"`
	GaugeConfigs        []GaugeConfig `json:"gaugeConfigurations"`
	PrintLines          bool          `json:"printLines"`
}

type LogExporterConfig struct {
	PollingConfigs []PollingConfig `json:"pollingConfigurations"`
}

func (conf *PollingConfig) Run() {
	reader := EsPoller{
		//eshost: "elasticsearch-dev-1991374869.us-west-2.elb.amazonaws.com",
		ESHost:      "localhost",
		ESPort:      "9200",
		IndexPrefix: "logstash-",
		UseSSL:      false,
		PrintLines:  conf.PrintLines,
		FieldNames: []string{"message",
			"log",
			"@timestamp",
			"timestamp",
			"level",
			"kubernetes.container_name",
			"kubernetes.namespace_name",
		},
		PollintIntervalSecs: conf.PollintIntervalSecs,
		QueryString:         conf.Query, // "level: debug AND Redistimer AND kubernetes.container_name: \"browserlab-container\" AND kubernetes.namespace_name: \"srf-integ-ga\"",
		TimeWindowSecs:      conf.TimeWindowSecs,
	}
	//reader.RegisterConsumer(&consumer.ConsoleConsumer{Format: "kubernetes"})
	for _, regexConf := range conf.GaugeConfigs {
		reader.RegisterConsumer(&consumer.PrometheusConsumer{Format: "kubernetes", GaugeType: regexConf.GaugeType, RegexStr: regexConf.RegexStr, GaugeName: regexConf.GaugeName, GaugeDescription: regexConf.GaugeDesc})
	}

	//var addr = flag.String("listen-address", ":8181", "The address to listen on for HTTP requests.")

	go reader.Start()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":8181", nil)
	if err != nil {
		logger.Fatalf("error in exporter http server: %v", err)
	}
}

func (configs *LogExporterConfig) Run() {
	for _, conf := range configs.PollingConfigs {
		conf.Run()
	}
}
