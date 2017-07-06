package consumer

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

type myLine struct {
	message string
	aType   string
}

func (line myLine) Message() string {
	return line.message
}
func (line myLine) GetField(fieldName string) string {
	val := reflect.ValueOf(line).FieldByName(fieldName).String()
	return val
}

func TestWordParsing(t *testing.T) {
	groupper := NewAutoGroupper()
	terms := groupper.getWordTerms("the quick brown fox jumped(over)the-lazy-dog windows-81 6A8A31F6-4309-453D-84CF-202EA5EB7C75 23425")
	fmt.Println(terms)
	sqlLine := "%!(EXTRA string=LicenseDataStore.getUpsertQuery: insert into \"security_1\".\"LICENSE\" (\"AP_FEATURE_ID\", \"AP_FEATURE_VERSION\", \"AP_PRODUCT_CODE\", \"CAPACITY\", \"EXPIRATION_DATE\", \"ID\", \"LICENSE_TYPE\", \"ORIGINAL_CAPACITY\", \"PURCHASE_METHOD\", \"START_DATE\", \"STATUS\", \"SUBSCRIPTION_ID\") values ('1535792', 'WebMinutes', 'FTaaS', 99999997602944, '2480550399000', '87995036-8655-4dfa-98e6-5bcdd5e896cd', 'FREE', '99999999999999', '1', '1479859200000', 'ACTIVE', '54321')ON CONFLICT (\"SUBSCRIPTION_ID\") DO UPDATE SET \"AP_PRODUCT_CODE\" = EXCLUDED.\"AP_PRODUCT_CODE\",\n\"AP_FEATURE_ID\" = EXCLUDED.\"AP_FEATURE_ID\",\n\"AP_FEATURE_VERSION\" = EXCLUDED.\"AP_FEATURE_VERSION\",\n\"PURCHASE_METHOD\" = EXCLUDED.\"PURCHASE_METHOD\",\n\"LICENSE_TYPE\" = EXCLUDED.\"LICENSE_TYPE\",\n\"STATUS\" = EXCLUDED.\"STATUS\",\n\"ORIGINAL_CAPACITY\" = EXCLUDED.\"ORIGINAL_CAPACITY\",\n\"CAPACITY\" = EXCLUDED.\"CAPACITY\",\n\"START_DATE\" = EXCLUDED.\"START_DATE\",\n\"EXPIRATION_DATE\" = EXCLUDED.\"EXPIRATION_DATE\" RETURNING *;)\n"
	terms = groupper.getWordTerms(sqlLine)
	fmt.Println(terms)
	statsdLine := "     [32m'statsd-global.packets_received'[39m: [33m0[39m,\n"
	terms = groupper.getWordTerms(statsdLine)
	fmt.Println(terms)
}
func TestMain(m *testing.M) {
	testSetupFunction()
	retCode := m.Run()
	//myTeardownFunction()
	os.Exit(retCode)
}

var lines []myLine

func testSetupFunction() {
	lines = []myLine{}
	lines = append(lines, myLine{aType: "a", message: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, myLine{aType: "a", message: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, myLine{aType: "b", message: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, myLine{aType: "a", message: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, myLine{aType: "c", message: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, myLine{aType: "a", message: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, myLine{aType: "c", message: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, myLine{aType: "c", message: "the quick brown fox jumped(the almighty junkheap)the-lazy-dog"})
	lines = append(lines, myLine{aType: "a", message: "the quick brown fox jumped(what you wish for)the-lazy-dog"})
	lines = append(lines, myLine{aType: "a", message: "the quick brown fox jumped(over)the-lazy-dog"})
	lines = append(lines, myLine{aType: "b", message: "the quick brown fox jumped(234253)the-lazy-dog"})
	lines = append(lines, myLine{aType: "a", message: "the quick brown fox jumped(under)the-lazy-dog"})
	lines = append(lines, myLine{aType: "b", message: "the quick brown fox jumped(moron)the-lazy-dog"})
	lines = append(lines, myLine{aType: "b", message: "the quick brown fox jumped(ever in demand)the-lazy-dog"})
	lines = append(lines, myLine{aType: "b", message: "the quick brown fox jumped(Johnny Henkock)the-lazy-dog"})
	lines = append(lines, myLine{aType: "c", message: "the quick brown fox jumped(velvet granny)the-lazy-dog"})
}

func TestAutoGrouper(t *testing.T) {
	//GetWordTerms("the quick brown fox jumped(over)the-lazy-dog")

	groupper := NewAutoGroupper()
	for _, l := range lines {
		groupper.FindGroup(l)
	}
	for _, gr := range groupper.groups {
		fmt.Printf("%v\n", gr.String())
	}
}

func TestFanOut(t *testing.T) {
	consumer := FanOutConsumer{LogField: "aType", autoCreate: &AutoCreateConfig{consumerType: "Console"}}

	for _, l := range lines {
		consumer.Consume(l)
	}

}
