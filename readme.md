### 手动打包
```jsx
    go env 查看配置环境
    go build -o hangdis.exe main.go
    cd redis/client/cmd/
    go build -o client.exe main.go
```

### 实现
- 自动过期功能(TTL)

### 实现中

-  string, list, hash, set, sorted set, bitmap 数据结构
- 发布订阅
- AOF 持久化及 AOF 重写
- 加载和导出 RDB 文件