package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	elastic "gopkg.in/olivere/elastic.v3"
)

// type Configuration struct {
// 	eshost string
// 	esport string
// }

//var configuration Configuration
func connectToElastic(hostURL string) *elastic.Client {
	// Create a context
	//ctx := context.Background()

	// Create a client
	client, err := elastic.NewClient()
	if err != nil {
		// Handle error
		panic(err)
	}

	// Ping the Elasticsearch server to get e.g. the version number
	//info, code, err := client.Ping(hostURL).Timeout("2").Do()
	info, code, err := client.Ping(hostURL).Do()

	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s", code, info.Version.Number)

	return client
}

func main() {
	var config = viper.New()
	config.SetConfigName("conf")
	config.SetConfigType("json")
	config.AddConfigPath(".") // look for config in the working directory

	err := config.ReadInConfig() // Find and read the config file
	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}

	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})

	viper.SetDefault("eshost", "localhost")
	viper.SetDefault("esport", "9200")

	host := config.Get("eshost")
	port := config.Get("esport")

	client := connectToElastic("http://" + host.(string) + ":" + port.(string))
	// Search with a term query
	termQuery := elastic.NewTermQuery("user.keyword", "olivere")
	searchResult, err := client.Search().
		Index("twitter").           // search in index "twitter"
		Query(termQuery).           // specify the query
		Sort("user.keyword", true). // sort by "user" field, ascending
		From(0).Size(10).           // take documents 0-9
		Pretty(true).               // pretty print request and response JSON
		Do()                        // execute
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	// for _, item := range searchResult.Each(reflect.TypeOf(typ)) {
	// 	if t, ok := item.(Tweet); ok {
	// 		fmt.Printf("Tweet by %s: %s\n", t.User, t.Message)
	// 	}
	// }
	// file, _ := os.Open("conf.json")
	// decoder := json.NewDecoder(file)
	// configuration := Configuration{
	// 	eshost: "localhost",
	// 	esport: "9200",
	// }

	// err = decoder.Decode(&configuration)
	// if err != nil {
	// 	fmt.Println("error:", err)
	// }
}
