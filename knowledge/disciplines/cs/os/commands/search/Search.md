# 搜索命令

Linux、macOS 等系统中的搜索命令可以按搜索对象分为文件内容搜索、文件路径查找和命令位置查找。ripgrep 是跨平台工具，命令名为 `rg`，可用于 Linux、macOS 和 Windows。

## 1. 常用命令

| 命令 | 搜索对象与用途 | 示例 |
| --- | --- | --- |
| `grep` | 搜索文件内容，支持正则表达式 | `grep -rn 'TODO' .` |
| `rg` | 搜索文件内容，默认递归搜索目录 | `rg -n 'TODO' .` |
| `find` | 按名称、路径、类型、时间等条件查找文件 | `find . -type f -name '*.py'` |
| `locate` | 从预先建立的索引中查找文件路径，索引可能不是最新的 | `locate report.pdf` |
| `command -v` | 查询 Shell 如何解析命令，结果可能是可执行文件路径、内建命令或别名等 | `command -v python` |
| `which` | 通常用于在 `PATH` 中查找可执行文件，具体行为依实现而异 | `which python` |

`grep` 和 `find` 通常随系统提供，`rg` 和 `locate` 是否可用取决于安装情况。

## 2. 文件内容搜索

`grep` 和 `rg` 都会读取文本、寻找满足搜索条件的内容，再输出匹配结果。两者用途相似，但具体实现、参数和默认行为不同。

| 方面 | 常见的 GNU grep | ripgrep |
| --- | --- | --- |
| 目录搜索 | 通常需要加 `-r` 或 `-R` | 默认递归搜索 |
| 文件筛选 | 按参数控制，不默认遵守 `.gitignore` | 默认遵守 `.gitignore` 等忽略规则，并跳过隐藏文件 |
| 多文件处理 | 通常逐个处理 | 支持多线程并行搜索 |
| 默认正则语法 | 基本正则表达式；`-E` 使用扩展正则表达式 | Rust 正则库支持的语法 |

两者都对二进制文件有特殊处理，具体行为取决于命令和参数；日常代码搜索通常不会像搜索普通文本一样完整输出二进制内容。

```bash
# 递归搜索，显示文件名、行号和匹配内容
grep -rn 'TODO' .
rg -n 'TODO' .

# 按普通字符串搜索，避免把点号等字符解释为正则语法
grep -Fn 'config.json' settings.txt
rg -Fn 'config.json' settings.txt

# 只搜索 Python 文件中的内容
rg -n -g '*.py' 'TODO' .
```

上面的 `-g '*.py'` 筛选搜索的文件，`'TODO'` 匹配文件内容，两者作用不同。由于默认筛选规则不同，`grep` 和 `rg` 搜索同一目录时，结果不一定相同。

工具的文件筛选规则、匹配引擎与性能原理见 [ripgrep](../../../../../tools/ripgrep/Ripgrep.md)。

## 3. 文件路径查找

`find` 遍历指定目录，再按条件筛选目录项；`locate` 查询预先建立的路径索引。

```bash
# 文件名等于 config.json，目录不限
find . -type f -name 'config.json'

# 文件名以 .py 结尾
find . -type f -name '*.py'

# 路径符合指定条件
find . -type f -path './src/config.json'

# 从索引中搜索包含 report.pdf 的路径
locate report.pdf

# 列出 ripgrep 默认搜索范围内的文件路径
rg --files
```

`find -name` 匹配文件名，`find -path` 匹配从搜索起点生成的路径。通配符应加引号，避免先被 Shell 展开。`rg --files` 也会受忽略规则和隐藏文件规则影响，因此不一定列出目录中的所有文件。

## 4. 路径匹配方式

### 4.1 精准匹配

“精准路径匹配”不是统一的命令行术语，通常表示路径必须完全等于指定值。使用时应说明比较的是绝对路径、相对路径，还是经过规范化处理的路径。

例如存在以下文件：

```text
src/config.json
test/config.json
src/config.json.bak
```

| 条件 | 匹配结果 |
| --- | --- |
| 文件名包含 `config` | 三个文件 |
| 文件名等于 `config.json` | 前两个文件 |
| 相对路径等于 `src/config.json` | 只有第一个文件 |

在上述目录的父级执行以下命令，会匹配路径为 `./src/config.json` 的普通文件：

```bash
find . -type f -path './src/config.json'
```

这里的模式没有通配符，因此按给定路径匹配。`find -path` 本身支持通配符，并非专门的路径相等比较操作。路径字符串相等也不等于文件身份相同，例如不同的符号链接可能指向同一个文件。

### 4.2 通配符匹配

通配符使用 `*`、`?` 等符号描述一组名称或路径，例如 `*.py` 表示以 `.py` 结尾的名称。不同工具对路径分隔符、`**` 和隐藏文件的处理可能不同，不能直接假定 Shell、`find` 和 `rg -g` 的规则完全一致。

```bash
find . -type f -name '*.py'
rg --files -g '*.py'
```

### 4.3 正则表达式匹配

正则表达式描述文本模式，与通配符是不同的语法。例如，正则中的 `.` 表示任意字符，`*` 表示前一个元素重复零次或多次。要匹配字面量点号，应写成 `\.`。

如果已有一个文件列表 `paths.txt`，每行保存一个普通路径字符串，可以用以下方式筛选：

```bash
# 正则匹配整行路径：^ 和 $ 限定行首与行尾
rg '^src/config\.json$' paths.txt

# 使用固定字符串整行匹配，无需转义点号
rg -Fx 'src/config.json' paths.txt
```

这里搜索的是列表中的文本行；若文件名本身包含换行符，逐行列表不能可靠表示路径，应使用支持 NUL 分隔的工具和参数。
