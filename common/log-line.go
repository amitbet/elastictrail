package common

type LogLine interface {
	Message() string
	GetField(fieldName string) string
}

// LogLine is a representation of one log line with several dynamic property fields
type ESLogLine struct {
	Content       map[string]interface{}
	messageFields []string
	timeFields    []string
	levelFields   []string
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
