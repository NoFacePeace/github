# 状态行

状态行（Status Line）是 HTTP/1.x 响应报文的第一行，用于说明 HTTP 版本和请求的处理结果。

状态行之后是[响应头](headers/Headers.md)，用于携带响应的附加信息。

## 1. 结构

```text
HTTP-version SP status-code SP [reason-phrase] CRLF
```

字段之间使用空格（`SP`）分隔，以回车换行（`CRLF`）结束。原因短语可以为空，但状态码后的空格仍需保留。

```http
HTTP/1.1 200 OK
```

- [HTTP 版本](Version.md)：示例中的 `HTTP/1.1`。
- [状态码](StatusCode.md)：示例中的 `200`，是表示处理结果的三位数字。
- [原因短语](ReasonPhrase.md)：示例中的 `OK`，是供人阅读的说明；客户端应依据状态码处理响应，而非依赖原因短语。

## 2. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 没有文本形式的状态行，通过 `:status` 伪首部表达状态码，不传输原因短语。
