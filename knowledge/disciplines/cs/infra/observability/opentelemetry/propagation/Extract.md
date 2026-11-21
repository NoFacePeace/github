# Extract

`Extract` 从入站请求的载体读取传播器支持的信息，并返回包含提取结果的上下文。

## 1. 从 HTTP 请求头恢复上下文

库存服务收到请求后，从请求头提取链路信息和 Baggage。用返回的上下文创建 Span，就可以接续同一条链路。下游代码也可以从该上下文读取 `region=sg`。

```go
ctx := propagator.Extract(incomingReq.Context(), propagation.HeaderCarrier(incomingReq.Header))
```
