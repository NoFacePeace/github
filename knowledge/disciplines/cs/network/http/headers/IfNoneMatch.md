# If-None-Match

`If-None-Match` 是 HTTP 条件请求首部，携带客户端保存的实体标签。只有目标表示不匹配这些标签时，服务器才执行请求方法。

## 1. 格式

```http
If-None-Match: "v1"
```

可以发送逗号分隔的多个标签；`*` 表示对任何当前表示进行匹配。标签比较使用弱比较规则。

## 2. 处理结果

- GET 或 HEAD 请求的标签匹配：返回 `304 Not Modified`，客户端可以复用缓存。
- 其他方法的标签匹配：条件不满足，通常返回 `412 Precondition Failed`。
- 标签不匹配：正常处理请求，GET 通常返回 `200` 和最新内容，无须另发一次请求。

## 3. 优先级

请求同时携带 `If-None-Match` 和 `If-Modified-Since` 时，服务器忽略 `If-Modified-Since`，优先使用实体标签条件。
