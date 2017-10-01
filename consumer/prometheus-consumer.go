package consumer

import (
	"elastictrail/common"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	//"gopkg.in/resty.v0"
)

type GaugeType int

const (
	GaugeTypeMatchCounter GaugeType = iota
	GaugeTypeExportMatchAsFloat
)

// format can be kubernetes of simple
type PrometheusConsumer struct {
	Format    string
	RegexStr  string
	GaugeType GaugeType
}

func (consumer *PrometheusConsumer) Consume(line common.LogLine) error {
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

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: "alive", Help: "Indicates the SRF farm liveliness 0 means dead!"})
	prometheus.MustRegister(gauge)
	re := regexp.MustCompile(consumer.RegexStr)
	switch consumer.GaugeType {

	case GaugeTypeMatchCounter:
		retStr := re.FindString(str)
		if retStr != "" {
			gauge.Inc()
		}
	case GaugeTypeExportMatchAsFloat:
		retStr := re.FindString(str)
		intVal, err := strconv.ParseFloat(retStr, 64)

		if err != nil {
			return err
		}
		gauge.Set(intVal)
	}

	// if checkIsAlive(*svcAddr) {
	// 	gauge.Set(1)
	// } else {
	// 	gauge.Set(0)
	// }

	// time.Sleep(10 * time.Second)

	// fmt.Printf("##line: %s\n", str)

	return nil
}

func (consumer *PrometheusConsumer) Name() string {
	return "prometheus-consumer"
}
