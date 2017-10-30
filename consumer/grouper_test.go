package consumer

import (
	"elastictrail/common"
	"fmt"
	"os"
	"testing"
)

func TestWordParsing(t *testing.T) {
	//groupper := NewAutoGroupper()
	line := common.SimpleLine{MyMessage: "the quick brown fox jumped(over)the-lazy-dog windows-81 6A8A31F6-4309-453D-84CF-202EA5EB7C75 23425"}
	terms := line.GetTerms()
	fmt.Println(terms)

	sqlLine := common.SimpleLine{MyMessage: "%!(EXTRA string=LicenseDataStore.getUpsertQuery: insert into \"security_1\".\"LICENSE\" (\"AP_FEATURE_ID\", \"AP_FEATURE_VERSION\", \"AP_PRODUCT_CODE\", \"CAPACITY\", \"EXPIRATION_DATE\", \"ID\", \"LICENSE_TYPE\", \"ORIGINAL_CAPACITY\", \"PURCHASE_METHOD\", \"START_DATE\", \"STATUS\", \"SUBSCRIPTION_ID\") values ('1535792', 'WebMinutes', 'FTaaS', 99999997602944, '2480550399000', '87995036-8655-4dfa-98e6-5bcdd5e896cd', 'FREE', '99999999999999', '1', '1479859200000', 'ACTIVE', '54321')ON CONFLICT (\"SUBSCRIPTION_ID\") DO UPDATE SET \"AP_PRODUCT_CODE\" = EXCLUDED.\"AP_PRODUCT_CODE\",\n\"AP_FEATURE_ID\" = EXCLUDED.\"AP_FEATURE_ID\",\n\"AP_FEATURE_VERSION\" = EXCLUDED.\"AP_FEATURE_VERSION\",\n\"PURCHASE_METHOD\" = EXCLUDED.\"PURCHASE_METHOD\",\n\"LICENSE_TYPE\" = EXCLUDED.\"LICENSE_TYPE\",\n\"STATUS\" = EXCLUDED.\"STATUS\",\n\"ORIGINAL_CAPACITY\" = EXCLUDED.\"ORIGINAL_CAPACITY\",\n\"CAPACITY\" = EXCLUDED.\"CAPACITY\",\n\"START_DATE\" = EXCLUDED.\"START_DATE\",\n\"EXPIRATION_DATE\" = EXCLUDED.\"EXPIRATION_DATE\" RETURNING *;)\n"}
	fmt.Println(sqlLine.GetTerms())

	statsdLine := common.SimpleLine{MyMessage: "     [32m'statsd-global.packets_received'[39m: [33m0[39m,\n"}
	fmt.Println(statsdLine.GetTerms())

	line = common.SimpleLine{MyMessage: "BasicFilter.composeFilter - Unknown filterKey: TENANTID"}
	fmt.Println(line.GetTerms())

	line = common.SimpleLine{MyMessage: "INVOKE Start. DispID = 6,ConnectionIndex = 0 Func=CProxy_IQTASUnitExecutionEngineEvents<class CQTASUnitExecutionEngine>::FireHelper File=e:\\ft\\qtp\\win32_release\\14.2.3667.0_clean\\qtp\\backend\\executionengine\\app\\qtexecutionengine\\qtasunitexecutionenginecp.h Line=43 ThreadID=6848"}
	fmt.Println(line.GetTerms())

	tokens := SplitWithMultiDelims(line.Message(), "`:\\/;'=-_+~<>[]{}!@#$%^&*().,?\"| \t\n")
	fmt.Println(tokens)
}

func TestMain(m *testing.M) {
	testSetupFunction()
	retCode := m.Run()
	//myTeardownFunction()
	os.Exit(retCode)
}

var lines []common.SimpleLine

func testSetupFunction() {
	lines = []common.SimpleLine{}
	lines = append(lines, common.SimpleLine{Type: "a", MyMessage: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "a", MyMessage: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "b", MyMessage: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "a", MyMessage: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "c", MyMessage: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "a", MyMessage: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "c", MyMessage: "the quick brown fox jumped(biggolo-the-jogolo)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "c", MyMessage: "the quick brown fox jumped(the almighty junkheap)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "a", MyMessage: "the quick brown fox jumped(what you wish for)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "a", MyMessage: "the quick brown fox jumped(over)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "b", MyMessage: "the quick brown fox jumped(234253)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "a", MyMessage: "the quick brown fox jumped(under)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "b", MyMessage: "the quick brown fox jumped(moron)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "b", MyMessage: "the quick brown fox jumped(ever in demand)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "b", MyMessage: "the quick brown fox jumped(Johnny Henkock)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "c", MyMessage: "the quick brown fox jumped(velvet granny)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "c", MyMessage: "the quick brown fox jumped(the almighty junkheap)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "a", MyMessage: "the quick brown fox jumped(what you wish for)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "a", MyMessage: "the quick brown fox jumped(over)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "b", MyMessage: "the quick brown fox jumped(234253)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "a", MyMessage: "the quick brown fox jumped(under)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "b", MyMessage: "the quick brown fox jumped(moron)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "b", MyMessage: "the quick brown fox jumped(ever in demand)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "b", MyMessage: "the quick brown fox jumped(Johnny Henkock)the-lazy-dog"})
	lines = append(lines, common.SimpleLine{Type: "c", MyMessage: "the quick brown fox jumped(velvet granny)the-lazy-dog"})
}

func TestAutoGrouper(t *testing.T) {
	//GetWordTerms("the quick brown fox jumped(over)the-lazy-dog")

	groupper := NewAutoGroupper()
	for _, l := range lines {
		groupper.FindGroup(&l)
	}
	for _, gr := range groupper.groups {
		fmt.Printf("%v\n", gr.String())
	}
}

func TestLineGroup(t *testing.T) {
	ag := NewAutoGroupper()
	//GetWordTerms("the quick brown fox jumped(over)the-lazy-dog")
	lineMessage := common.SimpleLine{MyMessage: "- remove JSON data from storage: 382993336 Func=MobileRtidUtils::JsonDataStorage::Remove File=e:\\ft\\qtp\\win32_release\\14.2.3667.0_clean\\qtp\\addins\\mobilepackage\\app\\mobilepackage\\MobileRtidUtils.h Line=78 ThreadID=652"}
	lg := NewLineGroup(&lineMessage)
	lg.lines[lineMessage.Message()] = true
	ag.lineCount++
	lg.generateTemplate(false)
	fmt.Println(lg.Template)
}
