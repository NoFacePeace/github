# WebDAV

WebDAV（Web Distributed Authoring and Versioning，Web 分布式创作和版本控制）是一组基于 HTTP 的扩展协议。它在 HTTP 读取和提交资源的能力之上，增加了目录管理、资源属性、复制、移动和锁定等功能，使客户端能够像操作远程文件系统一样管理服务器上的文件。

WebDAV 适用于网盘、NAS、文档管理系统、团队文件共享和远程文件挂载等场景。它定义的是远程资源管理协议，不是磁盘文件系统；服务器可以将 WebDAV 资源映射到本地文件、对象存储、数据库或其他后端。

## 1. 所在网络层

WebDAV 属于 OSI 模型的应用层（第 7 层），也是 TCP/IP 模型中的应用层协议。它复用 HTTP 的 URI、请求、响应、首部字段和状态码，并通过 HTTPS 获得传输加密。

典型协议栈如下：

```text
WebDAV
  -> HTTP / HTTPS
      -> TCP
          -> IP
              -> Ethernet / Wi-Fi
```

WebDAV 通常不使用独立端口：

- HTTP WebDAV 通常使用 TCP `80` 端口。
- HTTPS WebDAV 通常使用 TCP `443` 端口。

## 2. 资源模型

WebDAV 使用 URI 标识资源，并将资源分为两类：

- **普通资源**：通常表示文件或其他可以读取、写入的内容。
- **集合（Collection）**：类似文件系统中的目录，可以包含普通资源和其他集合。

例如：

```text
https://dav.example.com/documents/
https://dav.example.com/documents/report.pdf
```

其中，`documents/` 可以是集合，`report.pdf` 可以是集合中的普通资源。URI 末尾的 `/` 常用于表示集合，但客户端仍应以服务器返回的资源信息为准。

除了资源内容，WebDAV 还支持资源属性（Property）。属性可以描述资源名称、内容长度、媒体类型、创建时间和最后修改时间等元数据。属性通常使用 XML 表示，并分为：

- **活属性（Live Property）**：由服务器维护，例如内容长度和最后修改时间。
- **死属性（Dead Property）**：由客户端设置并由服务器保存的自定义元数据。

## 3. 常见方法

WebDAV 保留 HTTP 的 `GET`、`HEAD`、`PUT`、`DELETE` 等方法，并增加用于管理远程资源的方法：

| 方法 | 主要用途 |
| --- | --- |
| `PROPFIND` | 查询资源属性或集合中的成员。 |
| `PROPPATCH` | 设置或删除资源属性。 |
| `MKCOL` | 创建集合，作用类似创建目录。 |
| `COPY` | 复制资源或集合。 |
| `MOVE` | 移动资源或集合，也可以用于重命名。 |
| `LOCK` | 锁定资源，避免多个客户端同时修改造成冲突。 |
| `UNLOCK` | 使用锁令牌解除资源锁定。 |

常见 HTTP 方法在 WebDAV 中的用途包括：

| 方法 | 主要用途 |
| --- | --- |
| `GET` | 下载或读取资源内容。 |
| `PUT` | 创建或完整替换资源内容。 |
| `DELETE` | 删除资源或集合。 |
| `OPTIONS` | 查询服务器支持的协议能力和方法。 |

服务器不一定支持所有 WebDAV 功能。客户端通常先发送 `OPTIONS` 请求，并根据响应中的 `Allow` 和 `DAV` 首部判断服务器支持的方法与 WebDAV 能力等级。

## 4. 典型交互

客户端访问远程目录时，通常通过 `PROPFIND` 查询集合及其成员。`Depth` 请求首部用于指定查询深度：

- `Depth: 0`：只查询目标资源本身。
- `Depth: 1`：查询目标集合及其直接成员。
- `Depth: infinity`：递归查询全部后代资源，服务器可能因开销过大而拒绝。

下面的请求查询 `/documents/` 集合及其直接成员：

```http
PROPFIND /documents/ HTTP/1.1
Host: dav.example.com
Depth: 1
Content-Type: application/xml

<?xml version="1.0" encoding="utf-8"?>
<d:propfind xmlns:d="DAV:">
  <d:prop>
    <d:displayname/>
    <d:getcontentlength/>
    <d:getlastmodified/>
    <d:resourcetype/>
  </d:prop>
</d:propfind>
```

服务器通常返回 `207 Multi-Status`，响应体使用 XML 分别描述多个资源的处理结果：

```http
HTTP/1.1 207 Multi-Status
Content-Type: application/xml; charset=utf-8
```

上传文件时可以使用 `PUT`：

```http
PUT /documents/report.txt HTTP/1.1
Host: dav.example.com
Content-Type: text/plain
Content-Length: 12

hello webdav
```

创建目录时可以使用 `MKCOL`：

```http
MKCOL /documents/archive/ HTTP/1.1
Host: dav.example.com
```

## 5. 状态码

WebDAV 复用 HTTP 状态码，并补充了一些适合批量资源操作和依赖关系的状态码：

| 状态码 | 含义 |
| --- | --- |
| `201 Created` | 资源或集合创建成功。 |
| `204 No Content` | 操作成功，但没有响应内容。 |
| `207 Multi-Status` | 响应包含多个资源各自的处理结果。 |
| `403 Forbidden` | 服务器拒绝执行请求。 |
| `404 Not Found` | 目标资源不存在。 |
| `405 Method Not Allowed` | 目标资源不允许使用该方法。 |
| `409 Conflict` | 资源状态冲突，例如父集合不存在。 |
| `423 Locked` | 资源已被锁定，当前请求缺少有效锁令牌。 |
| `424 Failed Dependency` | 当前操作因依赖的其他操作失败而失败。 |
| `507 Insufficient Storage` | 服务器没有足够空间完成请求。 |

`207 Multi-Status` 只表示响应中包含多个结果，不代表其中每个操作都成功。客户端需要继续解析 XML 响应体中的每个 `response` 和对应状态。

## 6. 锁定与并发控制

多个客户端同时编辑同一个资源时，后提交的内容可能覆盖先提交的内容。WebDAV 可以通过 `LOCK` 创建写锁，并在响应中返回锁令牌。客户端修改或解锁资源时需要携带该令牌。

```text
客户端 A -- LOCK --> 服务器
客户端 A <-- 锁令牌 -- 服务器
客户端 A -- PUT + 锁令牌 --> 服务器
客户端 A -- UNLOCK + 锁令牌 --> 服务器
```

锁通常有超时时间，不应被视为永久锁。锁定也不能代替所有并发控制机制，客户端还可以使用 HTTP 条件请求首部，例如通过 `If-Match` 和 `ETag` 检测资源是否已被其他客户端修改。

## 7. 认证与安全

WebDAV 本身不规定唯一的身份认证方式，通常复用 HTTP 认证机制或由反向代理、应用服务器统一认证。常见方式包括 Basic、Digest、Bearer Token、客户端证书和基于 Cookie 的会话认证。

部署 WebDAV 时需要注意：

- 优先使用 HTTPS，避免账号、密码和文件内容以明文传输。
- Basic 认证只对凭据进行编码，不提供加密，必须与 HTTPS 配合使用。
- 按照最小权限原则限制用户可读取和修改的目录。
- 防止路径穿越、非法文件名和不受限制的文件上传。
- 限制单个文件大小、用户配额、请求体大小和递归查询深度。
- 谨慎记录请求日志，避免泄露认证信息、文件名和敏感路径。
- 对外提供服务时及时更新 Web 服务器和 WebDAV 组件。

## 8. 常见用途与局限

WebDAV 的常见用途包括：

- 将远程文件服务挂载为操作系统中的网络位置。
- 在 NAS、私有云和网盘之间同步或管理文件。
- 为文档管理系统提供统一的远程文件访问接口。
- 让编辑器或办公软件直接打开和保存服务器上的文件。

它的主要优点是基于 HTTP，容易复用现有的 HTTPS、认证、代理和防火墙基础设施，并且许多操作系统和文件管理工具原生支持。

WebDAV 也存在一些局限：

- 目录查询和批量属性响应使用 XML，解析和传输开销可能较大。
- 高延迟网络中的多次请求会影响挂载和浏览体验。
- 不同服务器和客户端支持的扩展能力可能不同。
- 文件锁、权限、文件名规则和大文件处理的兼容性需要实际验证。
- WebDAV 名称中虽然包含 Versioning，但基础 WebDAV 主要提供远程创作与资源管理；完整版本控制能力由其他扩展协议提供，并非所有服务器都支持。

因此，WebDAV 适合通用远程文件管理，但不能简单等同于本地文件系统、对象存储 API 或 Git 等版本控制系统。
