

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