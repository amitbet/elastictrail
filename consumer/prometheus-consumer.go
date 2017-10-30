package consumer

import (
	"elastictrail/common"
	"elastictrail/logger"
	"errors"
	"regexp"
	"strconv"
	"strings"

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
	Gauge            prometheus.Gauge
	Value            float64
}

func (consumer *PrometheusConsumer) BatchDone() {

	if consumer.GaugeType == GaugeTypeCountMatchesInTimespan {
		if consumer.Gauge != nil {
			logger.Debugf("Updating gauge %s value: %f", consumer.GaugeName, consumer.Value)
			consumer.Gauge.Set(consumer.Value)
			consumer.Value = 0
		}
	}
}
func fillFormatString(fmtStr string, line common.LogLine) string {
	working := fmtStr
	ret := fmtStr
	for {
		start := strings.Index(working, "{{")
		end := strings.Index(working, "}}")
		if start == -1 || end == -1 {
			break
		}

		tok := fmtStr[start+2 : end]
		working = string(fmtStr[end+2:])
		ret = strings.Replace(ret, "{{"+tok+"}}", line.GetField(tok), 1)
	}
	//logger.Debug("ret: " + ret)
	return ret
}

func (consumer *PrometheusConsumer) Consume(line common.LogLine) error {
	var str string

	//initialize gauge
	if consumer.Gauge == nil {
		if consumer.GaugeName == "" {
			logger.Error("PrometheusConsumer.Consume: Gauge set without a name, skipping!")
			return errors.New("Gauge set without a name, skipping!")
		}

		if consumer.GaugeDescription == "" {
			logger.Warn("Gauge set without a description, will use GaugeName instead: " + consumer.GaugeName)
			consumer.GaugeDescription = consumer.GaugeName
		}
		consumer.Gauge = prometheus.NewGauge(prometheus.GaugeOpts{Name: consumer.GaugeName, Help: consumer.GaugeDescription})
		prometheus.MustRegister(consumer.Gauge)
	}

	switch consumer.Format {
	case "kubernetes":
		str = line.GetField("@timestamp") + " " + line.GetField("kubernetes.container_name") + " NS: " + line.GetField("kubernetes.namespace_name") + " [" + line.GetField("level") + "] " + line.Message()
	case "simple":
		str = line.Message()
	default:
		str = fillFormatString(consumer.Format, line)
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
			consumer.Value++
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
			consumer.Gauge.Set(val)
		}
	}
	return nil
}

func (consumer *PrometheusConsumer) Name() string {
	return "prometheus-consumer"
}
