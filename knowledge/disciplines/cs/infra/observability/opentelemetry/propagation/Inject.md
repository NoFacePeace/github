# Inject

`Inject` 从当前上下文读取传播器支持的信息，并将其写入出站请求的载体。

## 1. 写入 HTTP 请求头

订单服务调用库存服务时，链路传播器可以写入 `traceparent`，Baggage 传播器可以写入 `baggage: region=sg`。

```go
propagator.Inject(ctx, propagation.HeaderCarrier(outgoingReq.Header))
```
