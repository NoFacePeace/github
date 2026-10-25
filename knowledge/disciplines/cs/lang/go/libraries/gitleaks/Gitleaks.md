# Gitleaks Go 使用

Gitleaks 可以作为 Go 库在进程内检测文件内容中的敏感信息，无需通过 `os/exec` 启动命令行程序。工具用途与检测原理见 [Gitleaks](../../../../security/tools/Gitleaks.md)。

## 1. 安装依赖

```bash
go get github.com/zricethezav/gitleaks/v8@v8.28.0
```

本文使用 v8.28.0 的 API 形式。项目主页与 Go 模块导入路径可能不同，导入时以目标版本的 `go.mod` 为准。

## 2. 创建检测器

先加载并校验检测规则，再创建检测器。以下为已取得配置后的核心调用：

```go
import (
    gitleaksconfig "github.com/zricethezav/gitleaks/v8/config"
    "github.com/zricethezav/gitleaks/v8/detect"
)

func newDetector(raw gitleaksconfig.ViperConfig) (*detect.Detector, error) {
    cfg, err := raw.Translate()
    if err != nil {
        return nil, err
    }
    return detect.NewDetector(cfg), nil
}
```

`raw` 可以从 TOML 配置解析得到。应明确加载的规则来源和允许列表，避免无规则或意外放行导致检测范围不符合预期。

## 3. 检测文件内容

```go
// detector 已创建并加载规则；content 为调用方读取的文件正文。
findings := detector.Detect(detect.Fragment{
    FilePath: "config.yaml",
    Raw:      string(content),
})

for _, finding := range findings {
    ruleID := finding.RuleID
    secret := finding.Secret
    startLine := finding.StartLine
    startColumn := finding.StartColumn
    // 使用命中结果执行后续处理，避免将 secret 直接写入日志。
    _, _, _, _ = ruleID, secret, startLine, startColumn
}
```

输入是文件的逻辑路径与正文；输出是命中项集合，包括规则、敏感内容和位置信息。文件读取、目录遍历与结果处置由调用方组织。

## 4. 与脱敏流程衔接

```text
读取文件内容
  → Detect 检测
  → 校验命中位置并合并重叠范围
  → 根据范围替换敏感内容
  → Detect 再次检测
  → 根据复检结果返回内容或错误
```

位置换算应遵循所用版本的行列定义，不能直接把列号当作字符串字符索引。UTF-8 字符与多行命中尤其需要验证；需要保留来源行号时，替换过程也应保留换行。

## 5. 参考资料

- [项目源码](https://github.com/gitleaks/gitleaks)。
- [Go 检测包 API](https://pkg.go.dev/github.com/zricethezav/gitleaks/v8/detect)。
