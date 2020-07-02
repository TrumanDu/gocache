package gocache

import (
	"fmt"
	db "gocache/internal/app/gocache/db"
)

func Run() {

	db.Set("truman", "trumandu")

	fmt.Println(db.Get("truman"))
}
