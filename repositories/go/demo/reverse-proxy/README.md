# httputil.ReverseProxy 示例

这个示例使用标准库的 `httputil.NewSingleHostReverseProxy`，将收到的 HTTP
请求转发到指定的上游服务，并在上游不可用时返回 `502 Bad Gateway`。

在 `repositories/go` 目录下，先启动一个用于测试的上游服务：

```sh
python3 -m http.server 8081
```

再启动反向代理：

```sh
go run ./demo/reverse-proxy
```

现在访问代理地址；请求路径会被转发给上游服务：

```sh
curl -i http://localhost:8080/
```

可通过参数修改监听地址或上游服务地址：

```sh
go run ./demo/reverse-proxy -listen :9090 -target https://example.com
```

`NewSingleHostReverseProxy` 会重写转发请求的协议、主机和路径，使其指向
`-target` 指定的上游；同时会保留并追加 `X-Forwarded-For` 等代理相关头部。
