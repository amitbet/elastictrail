package main

import (
	"elastictrail/logger"
	"elastictrail/reader"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

func main() {
	paramConfFile := flag.String("conf", "", "the config file to use, if not provided will try to use config.json in current dir")

	runDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	configFile := path.Join(runDir, "config.json")
	if *paramConfFile != "" {
		configFile = *paramConfFile
	}

	flag.Parse()
	logger.Infof("reading configuration file: %s", configFile)
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		logger.Error("error reading file: ", configFile, err)
		return
	}
	jsonStr := string(buf)
	m := reader.LogExporterConfig{}

	json.Unmarshal([]byte(jsonStr), &m)

	logger.Info("starting polling processes with config:" + jsonStr)
	m.Run()
}
