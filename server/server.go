package gocache

import (
	"fmt"
	cache "github.com/TrumanDu/gocache/store/cache"
)

func Run() {
	// 初始化socket 端口监听
	// epoll 建立client连接
	// 处理请求：解析命令
	// 主线程执行命令
	// 将结果返回给client
	cache.Set("truman", "trumandu")

	fmt.Println(cache.Get("truman"))
}
