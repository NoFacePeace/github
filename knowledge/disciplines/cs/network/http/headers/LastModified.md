# Last-Modified

`Last-Modified` 是 HTTP 响应首部，表示源服务器认为所选资源表示最后修改的日期和时间。它不同于表示报文生成时间的 `Date`。

## 1. 格式

使用 HTTP 日期格式，以 GMT 表示时间，精确到秒。

```http
Last-Modified: Tue, 15 Nov 1994 08:12:31 GMT
```

## 2. 条件请求

客户端可以保存该时间，在后续 GET 或 HEAD 请求中使用 `If-Modified-Since`：

```http
If-Modified-Since: Tue, 15 Nov 1994 08:12:31 GMT
```

资源未在该时间之后修改时，服务器可以返回 `304 Not Modified`，不发送响应体，让客户端继续使用缓存表示；发生修改时通常返回 `200 OK` 和最新内容。

## 3. 精度与限制

该字段只有秒级精度，可能无法区分同一秒内的多次修改。服务器无法可靠确定修改时间时，可以不发送该字段。需要更精确的表示验证时，可以使用 `ETag`。
