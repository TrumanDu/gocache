# go cache:shit:
go cache just for learn redis design and golang

:shit:一样的代码。。。。go语言初学者，请勿吐槽！
## 计划完成功能：
功能点|进度|备注
---|---|---
网络编程（Reactor）|100%|linux epoll实现
Redis协议解析（支持RESP2/3）|100%|仅实现了部分类型，详见以下
多线程IO|100%|
string 常用操作|100%|
hash 常用操作||
aof持久化|90%|
集群分布式||

### 网络编程（Reactor）
[网路IO模型开发](docs/网路IO模型开发.md)

### Redis协议解析
目前支持如下类型：

Type|Comment
---|---
Array|an ordered collection of N other types
Blob string| binary safe strings
Simple string|a space efficient non binary safe string
Simple error|a space efficient non binary safe error code and message
Number|an integer in the signed 64 bit range
Null|RESP2.0 null
### 多线程IO

### 支持命令
忽略大小写

命令|demo|返回值|备注
---|---|---|---
ping| |PONG|
quit| |OK||
set|set truman truman|OK| 
get|get truman |truman or null| 
del|del truman |1 or 0| 
exists|exists truman |1 or 0| 

### AOF持久化实现
[AOF持久化实现](docs/aof实现.md)
## 支持平台
因为使用linux epoll，目前仅打算支持linux

经测试支持**jedis,redis-cli**等客户端

## 涉及技术方向：
- 高并发
- 网络编程



