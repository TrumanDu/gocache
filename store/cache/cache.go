package cache

var cache = make(map[string]string)

func Get(key string) string {
	return cache[key]
}

func Del(key string) bool {
	delete(cache, key)
	return true
}

func Set(key string, val string) bool {
	cache[key] = val
	return true
}
