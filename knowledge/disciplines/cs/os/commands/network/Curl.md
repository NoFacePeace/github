# curl

`curl` 是通过 URL 传输数据的命令行工具，支持 HTTP、HTTPS 等协议，常用于接口请求、文件上传下载和网络排查。

## 1. 基本用法

```sh
curl <URL>
```

HTTP 请求默认使用 GET，响应正文默认输出到终端。`<URL>` 是占位符，执行前替换为实际地址。

## 2. 常用参数

| 参数 | 用途 |
| --- | --- |
| `-X <方法>` | 显式指定请求方法，例如 DELETE |
| `-H '<名称>: <值>'` | 添加请求头，可以重复使用 |
| `-d '<数据>'` | 发送请求数据，默认使用 POST 和表单内容类型 |
| `--data-binary @<文件>` | 按原始字节发送文件内容 |
| `-F '<字段>=<值>'` | 发送 multipart 表单，使用 `@文件路径` 上传文件 |
| `-i` | 同时输出响应头和正文 |
| `-I` | 发起 HEAD 请求，只获取响应头 |
| `-v` | 输出连接、TLS 和请求响应等调试信息 |
| `-L` | 跟随 HTTP 重定向 |
| `-o <文件>` | 将响应正文保存到指定文件 |
| `-O` | 使用 URL 路径中的文件名保存响应正文 |
| `-sS` | 隐藏进度信息，但保留错误信息 |
| `--fail` | 将 HTTP 400 及以上状态视为失败，并不输出响应正文 |
| `--connect-timeout <秒>` | 限制连接阶段的耗时 |
| `--max-time <秒>` | 限制整个传输的耗时 |
| `-w '<格式>'` | 传输完成后输出状态码、耗时等信息 |

使用 `-d` 或 `-F` 时通常无需再写 `-X POST`。`curl -I` 会改变请求方法；要查看 GET 请求的响应头，可使用 `-i`。

## 3. 示例

示例中的域名和接口路径需替换为实际服务地址。

### 3.1 发起 GET 请求

```sh
curl -i 'https://example.com/api/items?limit=10'
```

URL 使用引号包裹，避免其中的 `&` 等字符被 Shell 解释。

### 3.2 发送 JSON

```sh
curl 'https://example.com/api/items' \
  -H 'Content-Type: application/json' \
  -d '{"name":"demo"}'
```

`-d` 默认发起 POST，但默认内容类型是表单，因此发送 JSON 时需要指定 `Content-Type`。

### 3.3 携带访问令牌

假设已设置 `TOKEN` 环境变量：

```sh
curl 'https://example.com/api/items' \
  -H "Authorization: Bearer $TOKEN"
```

使用 `-v` 调试时，请求头可能包含认证信息，分享输出前应移除敏感内容。

### 3.4 下载文件

```sh
curl --fail -L -o archive.tar.gz 'https://example.com/archive.tar.gz'
```

`-L` 跟随重定向，`-o` 指定本地文件名。`--fail` 可以避免将 HTTP 错误页当作成功下载的内容。

### 3.5 上传文件

```sh
curl 'https://example.com/api/upload' -F 'file=@./report.txt'
```

这是 multipart 表单上传，服务端需要支持对应接口和 `file` 字段。

### 3.6 查看状态码与耗时

```sh
curl -sS -o /dev/null \
  --connect-timeout 5 --max-time 15 \
  -w 'HTTP %{http_code}\nTotal %{time_total}s\n' \
  'https://example.com'
```

默认情况下，收到 HTTP 404 或 500 并不一定让 `curl` 返回非零退出码；脚本需要将 HTTP 错误视为失败时，可添加 `--fail`。

[返回网络命令目录](Network.md)
