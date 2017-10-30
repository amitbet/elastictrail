package common

type LogLine interface {
	Message() string
	GetField(fieldName string) string
	GetTerms() []string
}
type Countable interface {
	Count() int
}
