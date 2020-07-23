# 网路IO模型开发
redis选用的单Reactor模型，虽然go 编程模型对于goroutine创建属于轻量级的，比
线程耗的资源更低，但是一个goroutine stack也会占用2k-8k。对于百万级连接，内存
占用也会很高，达到几十G,为了更好的性能，和更贴近redis设计。我这边也采用Reactor。
## Reactor模型
![](https://img2018.cnblogs.com/blog/1485398/201810/1485398-20181022232220631-1867817712.jpg)

Reactor模型其实就是IO多路复用+池化技术。

Reactor架构模式允许事件驱动的应用通过多路分发的机制去处理来自不同客户端的多个请求。




## 参考
1. [百万 Go TCP 连接的思考: epoll方式减少资源占用](https://colobu.com/2019/02/23/1m-go-tcp-connection/)
2. [smallnest/1m-go-tcp-server](https://github.com/smallnest/1m-go-tcp-server)
3. [epoll多路复用-----epoll_create1()、epoll_ctl()、epoll_wait()](https://blog.csdn.net/displayMessage/article/details/81151646)
4. [Minimal viable epoll package for go](https://gist.github.com/ast/a41816345e94e065890440e87e41a219)