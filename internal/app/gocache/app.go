package gocache

import (
	"fmt"
	db "github.com/TrumanDu/gocache/internal/app/gocache/db"
)

func Run() {

	db.Set("truman", "trumandu")

	fmt.Println(db.Get("truman"))
}