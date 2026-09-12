# 微信小程序 `V1MMWX` 包解密流程

## 1. 适用对象

本文针对从桌面版微信缓存中获取的、文件头为 `V1MMWX` 的小程序包。

本次分析的包信息：

```text
文件：my.wxapkg
AppID：wx9f30f1cea85e1e8c
原始大小：约 7.4 MB
```

## 2. 整体流程

```text
读取 wxapkg
    ↓
识别文件头 V1MMWX
    ↓
使用 AppID 派生 AES 密钥
    ↓
AES-256-CBC 解密前 1024 字节
    ↓
对剩余内容执行 XOR
    ↓
处理偏移 1023 处的 1 字节重叠
    ↓
校验普通 wxapkg 文件头
    ↓
解析文件索引并提取资源
```

## 3. 加密包结构

文件前 6 字节是格式标识：

```text
V1MMWX
```

去掉这 6 字节后，剩余内容分成两部分：

```text
payload[0:1024]       AES 加密部分
payload[1024:]         XOR 加密部分
```

## 4. 派生 AES 密钥

使用 AppID 通过 PBKDF2-HMAC-SHA1 派生 32 字节密钥：

```text
密码：AppID
Salt：saltiest
哈希：HMAC-SHA1
迭代次数：1000
输出长度：32 字节
```

本次使用的 AppID：

```text
wx9f30f1cea85e1e8c
```

等价的 Python 写法：

```python
import hashlib

key = hashlib.pbkdf2_hmac(
    "sha1",
    appid.encode(),
    b"saltiest",
    1000,
    32,
)
```

## 5. AES 解密前 1024 字节

算法参数：

```text
算法：AES-256-CBC
Key：上一步派生的 32 字节密钥
IV：the iv: 16 bytes
数据：payload[0:1024]
```

解密后得到前半部分明文：

```python
prefix = AES_256_CBC_DECRYPT(
    payload[0:1024],
    key,
    iv=b"the iv: 16 bytes",
)
```

## 6. XOR 解密剩余内容

XOR 密钥取 AppID 的倒数第二个字符的 ASCII 值：

```python
xor_key = ord(appid[-2])
```

本次 AppID 的倒数第二个字符为 `c`：

```text
ASCII('c') = 0x63
```

对剩余数据逐字节执行 XOR：

```python
suffix = bytes(
    value ^ xor_key
    for value in payload[1024:]
)
```

## 7. 处理 1 字节重叠

该格式存在一个特殊偏移：

```text
AES 明文：覆盖明文偏移 0～1023
XOR 明文：从明文偏移 1023 开始
```

因此不能直接拼接完整的 `prefix + suffix`，而应丢弃 AES 结果的最后 1 字节：

```python
plaintext = prefix[:1023] + suffix
```

## 8. wxapkg 文件头校验

解密后的内容应符合普通 wxapkg 格式：

```text
偏移 0：0xBE
偏移 13：0xED
```

示例校验：

```python
if plaintext[0] != 0xBE:
    raise ValueError("invalid wxapkg header")

if plaintext[13] != 0xED:
    raise ValueError("invalid wxapkg header")
```

如果校验失败，通常意味着 AppID 不正确、包格式不同，或 AES/XOR 参数不匹配。

## 9. 解析 wxapkg 索引

wxapkg 使用大端序的 4 字节整数。头部结构可概括为：

```text
偏移 0：1 字节，固定为 0xBE
偏移 1：4 字节，未知字段
偏移 5：4 字节，索引长度
偏移 9：4 字节，数据区长度
偏移 13：1 字节，固定为 0xED
偏移 14：4 字节，文件数量
```

随后依次读取每个文件的索引：

```text
文件名长度：4 字节
文件名：UTF-8，长度由上一个字段指定
文件偏移：4 字节
文件大小：4 字节
```

提取逻辑：

```python
for entry in entries:
    name = entry.name
    offset = entry.offset
    size = entry.size
    content = plaintext[offset:offset + size]
    save(name, content)
```

## 10. 本次解包结果

本次使用上述流程成功解密并解析：

```text
提取文件数：818
```

典型文件包括：

```text
app-config.json
app-service.js
app-wxss.js
appservice.app.js
chunk_*.appservice.js
chunk_*.webview.js
页面目录和静态资源
```

## 11. 完整伪代码

```python
data = read_file("my.wxapkg")
assert data.startswith(b"V1MMWX")

payload = data[6:]

key = pbkdf2_hmac_sha1(
    password=appid,
    salt=b"saltiest",
    iterations=1000,
    output_length=32,
)

prefix = aes_256_cbc_decrypt(
    payload[:1024],
    key=key,
    iv=b"the iv: 16 bytes",
)

xor_key = ord(appid[-2])
suffix = xor_each_byte(payload[1024:], xor_key)

plaintext = prefix[:1023] + suffix

assert plaintext[0] == 0xBE
assert plaintext[13] == 0xED

entries = parse_wxapkg_index(plaintext)

for entry in entries:
    write_file(
        entry.name,
        plaintext[entry.offset:entry.offset + entry.size],
    )
```

## 12. 注意事项

- 该流程只适用于能够确认是 `V1MMWX` 格式的包。
- AppID 必须与该小程序包对应，否则无法通过 wxapkg 文件头校验。
- 解包得到的是发布产物，不一定是开发者项目中的原始源码。
- JS 可能已经经过压缩、编译或混淆。
- 分包、插件包可能需要分别处理。
- 仅应分析自己拥有或获得授权的软件包。

## 13. 参考资料

- [legendxcheng/wxapkg_unpacker](https://github.com/legendxcheng/wxapkg_unpacker)
- [PC 微信小程序 V1MMWX 加密包逆向解析](https://blog.csdn.net/weixin_29327525/article/details/162569282)
