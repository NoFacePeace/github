# POST

POST 用于让目标资源按照自身语义处理提交的数据，常用于创建资源、提交表单或执行操作，并不只用于创建。

## 1. 请求示例

```http
POST /users HTTP/1.1
Host: example.com
Content-Type: application/json
Content-Length: 16

{"name":"Alice"}
```

POST 通常通过请求体提交数据，也可以包含 URL 查询字符串。请求体可以是 JSON、表单、文件等格式，由 `Content-Type` 说明。

## 2. 安全性与幂等性

POST 不是安全方法，也不保证幂等。重复相同请求可能创建多个资源或重复执行操作；应用可以使用幂等键等机制避免重复处理。

请求体不出现在 URL 中，但并不因此获得加密保护，传输敏感数据仍应使用 HTTPS。

## 3. 响应与缓存

创建资源成功时，通常返回 201，并通过 `Location` 指出主要新建资源的地址。其他操作可以返回 200、202、204 等状态码，具体取决于处理结果。

POST 响应并非绝对不可缓存。按 HTTP 规范，需要提供显式新鲜度信息，并携带与 POST 目标 URI 相同的 `Content-Location`：

```http
HTTP/1.1 200 OK
Content-Location: /report
Cache-Control: max-age=60
```

假设该响应对应 `POST /report`，在满足其他缓存复用条件时，缓存可以用于后续对 `/report` 的 GET 或 HEAD 请求。

**缓存的 POST 响应不能直接用于满足再次 POST。** 再次 POST 仍需要由服务器处理，不能因为存在缓存就跳过提交或执行操作。缓存 POST 响应也不会使 POST 变成幂等方法。

浏览器、代理和 CDN 对 POST 响应缓存的实际支持较少，不能假定所有实现都支持。获取可缓存数据时，通常优先设计 GET 接口。
