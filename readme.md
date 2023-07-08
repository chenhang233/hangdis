### 目录结构:

-   main 函数，执行入口
- config: 配置文件
- interface: 模块间的接口定义
- utils: 工具类，
- tcp: 服务器请求处理
- redis: redis 协议解析模块
  - client 命令行客户端
  - connection 客户端连接对象
  - parser 应用层协议解析
  - protocol 应用层协议回复
  - server 请求转发处理
- datastruct: redis 的各类数据结构实现
    - dict: hash 表
    - list: 链表
    - set： 集合
    - sortedset: 有序集合
    - bitmap 位运算
- database: 存储引擎核心
    - server.go:  服务实例
    - database.go: 单个database
    - router.go: 注册处理回调函数
    - universal.go: 通用命令
    - string.go: 字符串命令
    - list.go: 列表命令
    - hash.go: 哈希表命令
    - set.go: 集合命令
    - sortedset.go: 有序集合命令
    - pubsub.go: 发布订阅命令
    - geo.go: GEO 相关命令
    - systemcmd.go: 系统命令
- aof: AOF 持久化

### 实现
- 自动过期功能(TTL)
-  string, list, hash, set, sorted set, bitmap 数据结构
- 发布订阅

### 实现中

- AOF 持久化及 AOF 重写
- 加载和导出 RDB 文件


### 手动打包
```jsx
    go env 查看配置环境
    go build -o hangdis.exe main.go
    cd redis/client/cmd/
    go build -o client.exe main.go
```