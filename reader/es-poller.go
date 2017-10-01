package reader

import (
	"bytes"
	"elastictrail/common"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

// func fatalf(msg string, args ...interface{}) {
// 	fmt.Printf(msg, args...)
// 	os.Exit(2)
// }

// StringArray implements flag.Value interface
// type StringArray []string

// type consumerInfo struct {
// 	common.LogConsumer
// 	lineChannel chan common.LogLine
// }

// EsPoller is a log reader for the elasticsearch datasource
type EsPoller struct {
	//lines               []common.LogLine
	ESHost              string
	ESPort              string
	IndexPrefix         string
	UseSSL              bool
	QueryString         string
	PollintIntervalSecs int
	TimeWindowSecs      int

	fieldNames []string
	consumers  []consumerInfo
}

func NewEsPoller(queryStr string, elasticHost string, elasticPort string) *EsPoller {
	poller := &EsPoller{QueryString: queryStr, ESHost: elasticHost, ESPort: elasticPort}
	return poller
}

// RegisterConsumer creates a channel for log lines and runs the consumer.Consume function in go routine(s)
func (reader *EsPoller) RegisterConsumer(consumer common.LogConsumer) {
	ch := make(chan common.LogLine)
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
	reader.consumers = append(reader.consumers, consumerInf)
}

// Start runs the line query loop which starts gathering lines and distributing it to consumers
func (reader *EsPoller) Start() { // chan map[string]interface{}
	if reader.PollintIntervalSecs == 0 {
		reader.PollintIntervalSecs = 15
	}
	if reader.TimeWindowSecs == 0 {
		reader.PollintIntervalSecs = 60
	}
	if reader.ESPort == "" {
		reader.ESPort = "9200"
	}

	if reader.ESHost == "" {
		reader.ESHost = "localhost:9200"
	}
	if reader.IndexPrefix == "" {
		reader.IndexPrefix = "logstash-"
	}

	host := reader.ESHost + ":" + reader.ESPort
	msgFields := StringArray{}
	timeField := "@timestamp"
	include := "kubernetes.namespace_name:srf" //"kubernetes.namespace_name:srf-green-public"
	exclude := ""
	size := 1000

	//reader.UseSSL = false
	useSource := false

	// flag.StringVar(&host, "host", host, "host and port of elasticsearch")
	// flag.StringVar(&indexPrefix, "prefix", indexPrefix, "prefix of log indexes")
	// flag.Var(&msgFields, "message", "message fields to display")
	// flag.StringVar(&timeField, "timestamp", timeField, "timestap field to sort by")
	// flag.StringVar(&include, "include", include, "comma separated list of field:value pairs to include")
	// flag.StringVar(&exclude, "exclude", exclude, "comma separated list of field:value pairs to exclude")
	// flag.IntVar(&size, "size", size, "number of docs to return per polling interval")
	// flag.IntVar(&poll, "poll", poll, "time in seconds to poll for new data from ES")
	// flag.BoolVar(&useSSL, "ssl", useSSL, "use https for URI scheme")
	// flag.BoolVar(&useSource, "source", useSource, "use _source field to output result")
	// flag.BoolVar(&showID, "id", showID, "show _id field")

	// flag.Parse()

	// If no message field is explicitly requested we will follow @message
	if len(reader.fieldNames) == 0 {
		//msgFields = append(msgFields, "@message")
		reader.fieldNames = append(msgFields, "log")
		//msgFields = append(reader.fieldNames, "@timestamp")
	}

	exFilter := map[string]interface{}{}
	if len(include) == 0 && len(exclude) == 0 {
		exFilter["match_all"] = map[string]interface{}{}
	} else {
		filter := map[string]interface{}{}
		if len(include) > 0 {
			filter["must"] = getTerms(include)
		}
		if len(exclude) > 0 {
			filter["must_not"] = getTerms(exclude)
		}
		exFilter["bool"] = filter
	}

	startFromTime := time.Now().Add(time.Duration(reader.TimeWindowSecs))

	var scheme string
	if reader.UseSSL {
		scheme = "https"
	} else {
		scheme = "http"
	}
	rootURL := fmt.Sprintf("%s://%s", scheme, host)

	for {
		url1 := fmt.Sprintf("%s/*/_stats", rootURL)

		resp, err := http.Get(url1)
		if err != nil {
			fatalf("Error contacting Elasticsearch %s: %v", host, err)
		}
		status := map[string]interface{}{}
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			fatalf("Error decoding _status response from Elasticsearch: %v", err)
		}
		resp.Body.Close()

		indices := []string{}
		for k := range status["indices"].(map[string]interface{}) {
			if strings.HasPrefix(k, reader.IndexPrefix) {
				indices = append(indices, k)
			}
		}
		if len(indices) == 0 {
			fatalf("No indexes found with the prefix '%s'", reader.IndexPrefix)
		}
		sort.Strings(indices)
		index := indices[len(indices)-1]

		queryObj := map[string]interface{}{
			//"index": reader.indexPrefix,
			"size":   size,
			"fields": append(reader.fieldNames, timeField),
			"sort": map[string]interface{}{
				"@timestamp": map[string]interface{}{
					"order":         "asc",
					"unmapped_type": "boolean",
				},
			},
			"query": map[string]interface{}{
				"filtered": map[string]interface{}{
					"query_string": map[string]interface{}{
						"analyze_wildcard": true,
						"default_field":    "log",
						"query":            reader.QueryString,
					},
					"filter": map[string]interface{}{
						"bool": map[string]interface{}{
							"must": []interface{}{
								map[string]interface{}{
									"range": map[string]interface{}{
										timeField: map[string]interface{}{
											"gt": startFromTime.Format(time.RFC3339Nano),
										},
									},
								},
							},
							"must_not": []interface{}{},
						},
					},
				},
			},
		}
		// query := queryObj["query"].(map[string]interface{})
		// filter := query["filtered"].(map[string]interface{})

		// var sb bytes.Buffer
		// for key, value := range reader.filters {
		// 	sb.WriteString(key + ":" + value + " AND ")
		// }
		// qstr := sb.String()
		// qstr = qstr[:len(qstr)-5]

		// filter["query"] = map[string]interface{}{
		// 	"query_string": map[string]interface{}{
		// 		"analyze_wildcard": true,
		// 		"default_field":    "log",
		// 		"query":            qstr,
		// 	},
		// }

		req2, err := json.Marshal(queryObj)

		url := fmt.Sprintf("%s/%s/_search", rootURL, index)

		fmt.Println("***** " + string(req2))

		if err != nil {
			fatalf("Error creating search body: %v", err)
		}
		resp, err = http.Post(url, "application/json", bytes.NewReader(req2))
		if err != nil {
			fatalf("Error searching Elasticsearch: %v", err)
		}
		if resp.StatusCode != 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			fatalf("Elasticsearch failed: %s\n%s", resp.Status, string(body))
		}
		results := map[string]interface{}{}
		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			fatalf("Error reading search results: %v", err)
		}
		resp.Body.Close()

		lines := reader.handleQueryResults(results, useSource)
		for _, line := range lines {
			for _, consumerInf := range reader.consumers {
				consumerInf.lineChannel <- line
			}
		}

		time.Sleep(time.Duration(reader.PollintIntervalSecs) * time.Second)
	}
}

func (reader *EsPoller) handleQueryResults(results map[string]interface{}, useSource bool) []common.LogLine {

	lines := []common.LogLine{}
	hits := results["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, hit := range hits {
		h, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		var target map[string]interface{}
		if useSource {
			target = h["_source"].(map[string]interface{})
		} else {
			target = h["fields"].(map[string]interface{})
		}

		line := common.NewESLogLine(target, []string{"message", "log"}, []string{"timestamp", "@timestamp"}, []string{"level"})

		lines = append(lines, line)
	}
	return lines
}
