package consumer

import (
	"elastictrail/common"
	"elastictrail/logger"
	"errors"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	//"gopkg.in/resty.v0"
)

type GaugeType int

const (
	GaugeTypeCountMatchesInTimespan GaugeType = iota
	GaugeTypeExportMatchAsFloat
)

// format can be kubernetes of simple
type PrometheusConsumer struct {
	Format           string
	RegexStr         string
	GaugeType        GaugeType
	GaugeName        string
	GaugeDescription string
	regex            *regexp.Regexp
	gauge            prometheus.Gauge
	value            float64
}

func (consumer *PrometheusConsumer) BatchDone() {

	if consumer.GaugeType == GaugeTypeCountMatchesInTimespan {
		if consumer.gauge != nil {
			logger.Debugf("Updating gauge %s value: %f", consumer.GaugeName, consumer.value)
			consumer.gauge.Set(consumer.value)
			consumer.value = 0
		}
	}
}

func (consumer *PrometheusConsumer) Consume(line common.LogLine) error {
	var str string

	//initialize gauge
	if consumer.gauge == nil {
		if consumer.GaugeName == "" {
			logger.Error("PrometheusConsumer.Consume: Gauge set without a name, skipping!")
			return errors.New("Gauge set without a name, skipping!")
		}

		if consumer.GaugeDescription == "" {
			logger.Warn("Gauge set without a description, will use GaugeName instead: " + consumer.GaugeName)
			consumer.GaugeDescription = consumer.GaugeName
		}
		consumer.gauge = prometheus.NewGauge(prometheus.GaugeOpts{Name: consumer.GaugeName, Help: consumer.GaugeDescription})
		prometheus.MustRegister(consumer.gauge)
	}

	switch consumer.Format {
	case "kubernetes":
		str = line.GetField("@timestamp") + " " + line.GetField("kubernetes.container_name") + " NS: " + line.GetField("kubernetes.namespace_name") + " [" + line.GetField("level") + "] " + line.Message()
	case "simple":
		str = line.GetField("message")
	}

	//logger.Debug("prom consumer consuming line! " + str)
	//initialize stuff on the first line recieved
	if consumer.regex == nil {
		consumer.regex = regexp.MustCompile(consumer.RegexStr)
	}

	// if retStr != "" {
	// 	logger.Debug("Regex search Found: " + retStr + " in line: " + str)
	// }

	switch consumer.GaugeType {
	case GaugeTypeCountMatchesInTimespan:
		if consumer.regex.MatchString(str) {
			consumer.value++
		}
	case GaugeTypeExportMatchAsFloat:
		retMatch := consumer.regex.FindStringSubmatch(str)
		if len(retMatch) > 0 {
			val, err := strconv.ParseFloat(retMatch[1], 64)

			if err != nil {
				logger.Error("Parse float error while parsing: " + retMatch[1] + " in line: " + str)
				return err
			}
			logger.Debugf("Updating gauge %s value: %f", consumer.GaugeName, val)
			consumer.gauge.Set(val)
		}
	}
	return nil
}

func (consumer *PrometheusConsumer) Name() string {
	return "prometheus-consumer"
}
