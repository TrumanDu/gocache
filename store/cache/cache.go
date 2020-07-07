package cache

var cache = make(map[string]string)

func Get(key string) string {
	return cache[key]
}

func Set(key string, val string) bool {
	cache[key] = val
	return true
}
