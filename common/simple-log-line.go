package common

type SimpleLine struct {
	MyMessage string
	Type      string
	Terms     []string
	Fields    map[string]string
}

func (line *SimpleLine) Message() string {
	return line.MyMessage
}
func (line *SimpleLine) GetField(fieldName string) string {
	val := line.Fields[fieldName]
	return val
}

// GetTerms returns the terms representing the line, this implementation also caches them
func (line *SimpleLine) GetTerms() []string {
	var numericCharAcceptRatio float32 = 0.2 // the maximum ratio of numbers/other chars in a term (more will delete the term)
	if line.Terms == nil {
		line.Terms = getWordTerms(line.Message(), numericCharAcceptRatio, lineSeparators)
	}

	return line.Terms
}
