<div align="center">
  <h1>lg</h1>
</div>
<p align="center">
<a href="https://travis-ci.org/qclaogui/lg"><img src="https://travis-ci.org/qclaogui/lg.svg?branch=master"></a>
<a href="https://goreportcard.com/report/github.com/qclaogui/lg"><img src="https://goreportcard.com/badge/github.com/qclaogui/lg?v=1" /></a>
<a href="https://godoc.org/github.com/qclaogui/lg"><img src="https://godoc.org/github.com/qclaogui/lg?status.svg"></a>
<a href="https://github.com/qclaogui/lg/blob/master/LICENSE"><img src="https://img.shields.io/github/license/qclaogui/lg.svg" alt="License"></a>
</p>

Project logger packaging for [github.com/uber-go/zap](https://github.com/uber-go/zap)

## Getting Started

Suitable for project log in the container


### Installation

`go get -u github.com/qclaogui/lg`


### Usage [Run in Playground](https://play.golang.org/p/_q67O0B0Dd5)



```go
package main

import (
	"github.com/qclaogui/lg"
)

func main() {
	//lg.TimeFormat = time.RFC3339Nano

	// lvl - global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	_ = lg.Init(-1, "demo-project")

	lg.APPLog.Info("Happy Goding!")
}
// Output: {"level":"info","ts":1564834577.710078,"msg":"Happy Goding!","info":{"project":"demo-project","hostname":"qclaogui.local"}}
```

## Versioning

Using [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/qclaogui/lg/tags). 


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
