//go:build with_extend

package main

import (
	//注册扩展组件库
	// 使用`go build -tags with_extend .`把扩展组件编译到运行文件
	_ "github.com/rulego/rulego-components/external/kafka"
	_ "github.com/rulego/rulego-components/external/redis"
	_ "github.com/rulego/rulego-components/filter"
	_ "github.com/rulego/rulego-components/transform"
)
