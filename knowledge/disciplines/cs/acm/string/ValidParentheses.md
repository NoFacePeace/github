# 有效的括号

## 1. 题目描述

[LeetCode 20：有效的括号](https://leetcode.cn/problems/valid-parentheses/)。

给定一个只包含字符 `'('`、`')'`、`'['`、`']'`、`'{'`、`'}'` 的字符串 `s`，判断该字符串是否有效。

有效字符串必须满足：

- 左括号必须用相同类型的右括号闭合。
- 左括号必须以正确的顺序闭合。
- 每个右括号都有对应的左括号。

## 2. 示例

```text
输入：s = "()"
输出：true
```

```text
输入：s = "()[]{}"
输出：true
```

```text
输入：s = "(]"
输出：false
解释：左右括号的类型不匹配。
```

```text
输入：s = "([)]"
输出：false
解释：虽然每种括号的数量相同，但闭合顺序不正确。
```

## 3. Go 实现

使用[栈](../../data-structures/stack/Stack.md)保存尚未匹配的左括号。遇到右括号时，检查它是否与栈顶的左括号匹配，匹配则出栈，否则返回 `false`；遍历结束后，栈为空才表示所有括号均已匹配。

```go
func isValid(s string) bool {
    m := map[string]string{
        ")": "(",
        "}": "{",
        "]": "[",
    }
    stack := []string{}
    for i := range s {
        str := s[i:i+1]
        val, ok := m[str]
        if !ok {
            stack = append(stack, str)
            continue
        }
        if len(stack) == 0 {
            return false
        }
        n := len(stack)
        if stack[n-1] != val {
            return false
        }
        stack = stack[:n-1]
    }
    if len(stack) == 0 {
        return true
    }
    return false
}
```
