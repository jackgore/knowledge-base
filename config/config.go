package config

type Config interface {
	GetStrings(keys []string) []string
	GetString(key string) string
	GetInts(keys []string) []int
	GetInt(key string) int
}
