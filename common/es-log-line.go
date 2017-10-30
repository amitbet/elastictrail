package common

import "strings"

// LogLine is a representation of one log line with several dynamic property fields
type ESLogLine struct {
	Content       map[string]interface{}
	messageFields []string
	timeFields    []string
	levelFields   []string
	Terms         []string
}

func NewESLogLine(
	content map[string]interface{},
	messageFields []string,
	timeFields []string,
	levelFields []string,
) *ESLogLine {
	line := ESLogLine{Content: content, messageFields: messageFields, timeFields: timeFields, levelFields: levelFields}
	return &line
}

var lineSeparators map[rune]bool = getSeperators()

func getSeperators() map[rune]bool {
	strSeparators := "`:\\/;'=-_+~<>[]{}!@#$%^&*().,?\"| \t\n"
	separators := map[rune]bool{}
	for _, sep := range strSeparators {
		separators[sep] = true
	}
	return separators
}

// GetTerms returns the terms representing the line, this implementation also caches them
func (line *ESLogLine) GetTerms() []string {
	var numericCharAcceptRatio float32 = 0.2 // the maximum ratio of numbers/other chars in a term (more will delete the term)
	if line.Terms == nil {
		line.Terms = getWordTerms(line.Message(), numericCharAcceptRatio, lineSeparators)
	}

	return line.Terms
}

func countNumericChars(term string) int {
	numCounter := 0
	for _, ch := range term {
		if ch >= '0' && ch <= '9' {
			numCounter++
		}
	}
	return numCounter
}

func getWordTerms(line string, numericCharAcceptRatio float32, separators map[rune]bool) []string {
	terms := []string{}

	runes := []rune{}
	order := 1
	for _, ch := range line {
		if separators[ch] {
			term := string(runes)
			numericCount := countNumericChars(term)
			numericRatio := float32(numericCount) / float32(len(term))
			term = strings.Trim(term, string([]rune{27})+"\t\r\n ")
			//strings.TrimSpace()
			if runes != nil && term != "" && len(term) > 0 && numericRatio <= numericCharAcceptRatio {
				terms = append(terms, term)
				order++
			}
			runes = []rune{}
		} else {
			runes = append(runes, ch)
		}
	}
	//add last term
	if len(runes) > 0 {
		term := string(runes)

		numericCount := countNumericChars(term)
		numericRatio := float32(numericCount) / float32(len(term))
		term = strings.Trim(term, string([]rune{27})+"\t\r\n ")
		//strings.TrimSpace()
		if runes != nil && term != "" && len(term) > 0 && numericRatio <= numericCharAcceptRatio {
			terms = append(terms, term)
			order++
		}
	}
	return terms
}

// GetFirstField returns the first existing field value for the given fields list or empty string if none exist
func (line *ESLogLine) GetFirstField(fieldNames []string) string {
	for _, name := range fieldNames {
		if val := line.GetField(name); val != "" {
			return val
		}
	}
	return ""
}

// GetField returns the field value for the given field name
func (line *ESLogLine) GetField(fieldName string) string {
	cont := line.Content[fieldName]
	if cont == nil {
		return ""
	}

	//bytes, _ := json.Marshal(line.Content)
	//fmt.Println("message content:" + string(bytes))

	item := cont.([]interface{})[0]
	if item == nil {
		return ""
	}
	return item.(string)
}

func (line *ESLogLine) String(fieldName string) string {
	return ""
}

// Message returns the first message field that exists in the line definition
func (line *ESLogLine) Message() string {

	//fmt.Printf("#Message:%s:%v\n", line.messageField, line.getString(line.messageField)) //line.Content[line.messageField])
	return line.GetFirstField(line.messageFields)
	//return ""
}

// Time returns the first message field that exists in the line definition
func (line *ESLogLine) Time() string {
	//fmt.Printf("#Time:%s cont: %s\n", line.timeField, line.getString(line.timeField))
	return line.GetFirstField(line.timeFields)
	//return ""
}

// Level returns the first message field that exists in the line definition
func (line *ESLogLine) Level() string {
	return line.GetFirstField(line.levelFields)
	//return ""
}
