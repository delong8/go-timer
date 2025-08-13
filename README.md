# go-timer

A simple timer utility package for Go.

## Installation

```bash
go get github.com/delong/go-timer
```

## Usage

```go
package main

import (
	"fmt"
	"time"
	
	"github.com/delong/go-timer"
)

func main() {
	// Create a new timer
	t := timer.New()
	
	// Wait for a moment
	time.Sleep(time.Second)
	
	// Get elapsed time
	fmt.Printf("Elapsed time: %v\n", t.Elapsed())
	
	// Reset the timer
	t.Reset()
	
	// Wait again
	time.Sleep(time.Second)
	
	// Get elapsed time after reset
	fmt.Printf("Elapsed time after reset: %v\n", t.Elapsed())
}
```

## API

- `timer.New()` - Creates and starts a new timer
- `timer.Elapsed()` - Returns the time elapsed since the timer was started
- `timer.Reset()` - Resets the timer to the current time

## License

MIT