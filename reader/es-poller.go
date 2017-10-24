package reader

import (
	"bytes"
	"elastictrail/common"
	"elastictrail/consumer"
	"os"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
	"vncproxy/logger"
)

func fatalf(msg string, args ...interface{}) {
	logger.Fatalf(msg, args...)
	//fmt.Printf(msg, args...)
	os.Exit(2)
}

// split string and parse to terms for query filter
func getTerms(args string) []map[string]interface{} {
	terms := []map[string]interface{}{}
	for k, v := range parsePairs(args) {
		terms = append(terms, map[string]interface{}{"term": map[string]interface{}{k: v}})
	}
	return terms
}

// split string and parse to key-value pairs
func parsePairs(args string) map[string]string {
	exkv := map[string]string{}
	for _, pair := range strings.Split(args, ",") {
		kv := strings.Split(pair, ":")
		if _, ok := exkv[kv[0]]; ok {
			exkv[kv[0]] = exkv[kv[0]]
		} else {
			exkv[kv[0]] = string(kv[1])
		}
	}
	return exkv
}

//StringArray implements flag.Value interface
type StringArray []string

type consumerInfo struct {
	common.LogConsumer
	lineChannel chan common.LogLine
}

// EsPoller is a log reader for the elasticsearch datasource
type EsPoller struct {
	ESHost                    string //elasticsearch host
	ESPort                    string //elasticsearch port
	IndexPrefix               string //elasticsearch index prefix
	UseSSL                    bool   //run query using https or http
	QueryString               string //text search query in elastic search syntax
	PollintIntervalSecs       int    //polling interval
	TimeWindowSecs            int    //time window for the query, if not set will default to continuous - using the last result timestamp as the line-timestamp to start from
	PrintLines                bool
	FieldNames                []string //fields to include in the query results
	common.ConsumerDictionary          // the consumer registry
}

func NewEsPoller(queryStr string, elasticHost string, elasticPort string) *EsPoller {
	poller := &EsPoller{QueryString: queryStr, ESHost: elasticHost, ESPort: elasticPort}
	return poller
}

// Start runs the line query loop which starts gathering lines and distributing it to consumers
func (reader *EsPoller) Start() { // chan map[string]interface{}
	if reader.PollintIntervalSecs == 0 {
		reader.PollintIntervalSecs = 15
	}
	// if reader.TimeWindowSecs == 0 {
	// 	reader.PollintIntervalSecs = 60
	// }
	if reader.ESPort == "" {
		reader.ESPort = "9200"
	}
	// if reader.Format == "" {
	// 	reader.Format = "kubernetes"
	// }
	if reader.ESHost == "" {
		reader.ESHost = "localhost"
	}
	if reader.IndexPrefix == "" {
		reader.IndexPrefix = "logstash-"
	}
	lastTime := time.Now().Add(time.Duration(-reader.PollintIntervalSecs) * time.Second)

	host := reader.ESHost + ":" + reader.ESPort
	msgFields := StringArray{}
	timeField := "@timestamp"
	size := 1000

	useSource := false

	// If no message field is explicitly requested we will follow @message
	if len(reader.FieldNames) == 0 {
		//msgFields = append(msgFields, "@message")
		reader.FieldNames = append(msgFields, "log")
		//msgFields = append(reader.FieldNames, "@timestamp")
	}

	var scheme string
	if reader.UseSSL {
		scheme = "https"
	} else {
		scheme = "http"
	}
	rootURL := fmt.Sprintf("%s://%s", scheme, host)

	if reader.PrintLines {
		reader.RegisterConsumer(&consumer.ConsoleConsumer{Format: "kubernetes"})
	}

	for {
		url1 := fmt.Sprintf("%s/*/_stats", rootURL)

		var startFromTime time.Time
		if reader.TimeWindowSecs > 0 {
			startFromTime = time.Now().Add(time.Duration(-reader.TimeWindowSecs) * time.Second)
		} else {
			startFromTime = lastTime
		}

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
			"fields": append(reader.FieldNames, timeField),
			"sort": map[string]interface{}{
				"@timestamp": map[string]interface{}{
					"order":         "asc",
					"unmapped_type": "boolean",
				},
			},
			"query": map[string]interface{}{
				"filtered": map[string]interface{}{

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
		query := queryObj["query"].(map[string]interface{})
		filter := query["filtered"].(map[string]interface{})

		if reader.QueryString != "" {

			filter["query"] = map[string]interface{}{
				"query_string": map[string]interface{}{
					"analyze_wildcard": true,
					"default_field":    "log",
					"query":            reader.QueryString,
				},
			}
		}

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
		logger.Debugf("***** Got %d lines from query\n", len(lines))
		for _, line := range lines {
			if reader.TimeWindowSecs <= 0 {
				lastTime, err = time.Parse(time.RFC3339Nano, line.GetField(timeField))
			}
			//linestr := line.GetField("@timestamp") + " " + line.GetField("kubernetes.container_name") + " NS: " + line.GetField("kubernetes.namespace_name") + " [" + line.GetField("level") + "] " + line.Message()
			//fmt.Println("***** line: >> " + linestr)
			reader.ConsumerDictionary.Distribute(line)
		}
		reader.ConsumerDictionary.DistributeBatchDone()

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
