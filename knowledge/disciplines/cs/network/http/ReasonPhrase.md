# 原因短语

原因短语（Reason Phrase）是 HTTP/1.x 状态行中位于状态码之后的文本，用于向人说明响应的处理结果。

## 1. 示例

```http
HTTP/1.1 200 OK
```

其中 `OK` 是原因短语。其他常见示例包括 `404 Not Found` 中的 `Not Found`，以及 `500 Internal Server Error` 中的 `Internal Server Error`。

## 2. 与状态码的关系

响应的协议语义由状态码决定，原因短语仅供人阅读，可以使用不同文本或留空。客户端应依据状态码处理响应，不应依赖原因短语的具体内容。

原因短语为空时，HTTP/1.x 状态行仍需保留状态码后的空格，并以回车换行（`CRLF`）结束。

## 3. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 不传输原因短语，只通过 `:status` 伪首部表达状态码。
