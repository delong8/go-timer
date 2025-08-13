# go-timer

A simple timer utility package for Go.

## Installation

```bash
go get github.com/delong8/go-timer
```

## Usage

```go
package main

import (
	"fmt"
	"time"
	
	"github.com/delong8/go-timer"
)

func main() {
	timer.Init()
	timer.RegisteDaily("my_daily_task", "9:00", func() {
		fmt.Println("Daily task executed!")
	})
	timer.RegisteInterval("my_interval_task", 5, func(){
		fmt.Println("Interval task executed!")
	})
}
```

## License

MIT