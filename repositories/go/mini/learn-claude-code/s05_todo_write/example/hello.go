package main

// 对应 Python 版 s05_todo_write/example/hello.py：
//   def greet(name):
//       message = "Hello, " + name
//       print(message)
//
//   greet("Claude")
//
// 展示最小可运行的 Go 程序结构：函数定义 + 调用。

import "fmt"

// greet 把问候语拼好并打印，对齐 Python 版 greet(name)。
func greet(name string) {
	message := "Hello, " + name
	fmt.Println(message)
}

func main() {
	greet("Claude")
}
