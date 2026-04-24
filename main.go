package main

import (
	"context"
	"os"

	"github.com/hczs/pxy/cmd"
)

func main() {
	// Go 程序的入口只有 main.main；这里像 Java 的 main 方法一样只做启动转发。
	os.Exit(cmd.Run(context.Background(), os.Args, os.Stdout, os.Stderr))
}
