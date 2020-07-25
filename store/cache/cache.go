package cache

var cache = make(map[string]string)

func Get(key string) string {
	return cache[key]
}

func Del(key string) int {
	if _, ok := cache[key]; ok {
		delete(cache, key)
		return 1
	} else {
		return 0
	}
}

func Set(key string, val string) bool {
	cache[key] = val
	return true
}

func Exists(key string) int {
	if _, ok := cache[key]; ok {
		return 1
	} else {
		return 0
	}
}
