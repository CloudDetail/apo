package utils

type Filter interface {
	ExtractFilterStr() []string
}
