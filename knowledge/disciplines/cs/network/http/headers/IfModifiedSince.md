# If-Modified-Since

`If-Modified-Since` 是 GET 或 HEAD 条件请求首部，通常携带之前响应中的 `Last-Modified` 时间，用于验证缓存表示是否发生变化。

## 1. 格式

```http
If-Modified-Since: Tue, 15 Nov 1994 08:12:31 GMT
```

## 2. 处理结果

- 表示未在指定时间之后修改：对于原本会成功返回表示的请求，服务器返回 `304 Not Modified`，客户端复用缓存。
- 表示已修改：正常处理请求，GET 通常返回 `200` 和最新内容，无须另发一次请求。

修改时间只有秒级精度，无法可靠区分同一秒内的多次修改。

## 3. 忽略条件

请求包含 `If-None-Match`、方法不是 GET 或 HEAD、字段值不是有效 HTTP 日期，或服务器没有所选表示的修改时间时，服务器必须忽略该条件字段。
