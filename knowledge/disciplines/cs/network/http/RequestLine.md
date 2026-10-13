# 请求行

请求行（Request Line）是 HTTP/1.x 请求报文的第一行，用于说明请求方法、请求目标和 HTTP 版本。

请求行之后是[请求头](Headers.md)，用于携带请求的附加信息。

## 1. 结构

```text
method SP request-target SP HTTP-version CRLF
```

三个字段之间使用空格（`SP`）分隔，以回车换行（`CRLF`）结束。

```http
GET /users?page=1 HTTP/1.1
```

- [请求方法](Methods.md)：示例中的 `GET`，表示希望对目标资源执行的操作。
- [请求目标](RequestTarget.md)：示例中的 `/users?page=1`，包含路径和查询字符串。这是直接向源站发送请求时的常见形式。
- HTTP 版本：示例中的 `HTTP/1.1`。

## 2. HTTP/2 与 HTTP/3

HTTP/2 与 HTTP/3 没有文本形式的请求行，而是通过 `:method`、`:scheme`、`:authority`、`:path` 等伪首部表达请求控制信息。
