# Content-Type

`Content-Type` 可用于 HTTP 请求和响应，说明所关联内容的媒体类型，帮助接收方理解和处理内容。

## 1. 格式

字段值由类型、子类型和可选参数组成。

```http
Content-Type: text/html; charset=utf-8
```

其中 `text/html` 是媒体类型，`charset=utf-8` 是字符集参数。参数的含义取决于具体媒体类型，并非所有类型都支持相同参数。

## 2. 常见媒体类型

| 媒体类型 | 内容 |
| --- | --- |
| application/json | JSON 数据 |
| text/plain | 纯文本 |
| text/html | HTML 文档 |
| application/x-www-form-urlencoded | URL 编码的表单数据 |
| multipart/form-data | 多部分表单数据，常用于文件上传，需要 `boundary` 参数 |
| application/octet-stream | 通用二进制数据 |

## 3. 请求与响应

- 请求中：说明客户端发送的请求体格式。
- 响应中：说明服务器返回的内容格式；HEAD 响应也可以说明对应 GET 响应的内容格式。

`Content-Type` 描述当前内容的格式；`Accept` 表达客户端希望收到哪些媒体类型；`Content-Encoding` 描述 gzip 等内容编码。
