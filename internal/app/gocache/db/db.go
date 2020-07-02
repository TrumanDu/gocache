package db

var db = make(map[string]string)

func Get(key string) string {
	return db[key]
}

func Set(key string, val string) bool {
	db[key] = val
	return true
}
