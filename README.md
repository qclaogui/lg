# lg

Project logger packaging

## Getting Started

Suitable for project log in the container

```
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

```

## Versioning

Using [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/qclaogui/lg/tags). 


## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
