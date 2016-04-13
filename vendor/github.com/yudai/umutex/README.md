# Unblocking Mutex

This simple package provides unblocking mutexes for those who don't want to write many `select` clauses or get confused by numerous channels.

## Usage Example

```go
package main

import (
	"fmt"
	"github.com/yudai/umutex"
)

func main() {
	// Create mutex
	mutex := umutex.New()

	// First time, try should succeed
	if mutex.TryLock() {
		fmt.Println("SUCCESS")
	} else {
		fmt.Println("FAILURE")
	}

	// Second time, try should fail as it's locked
	if mutex.TryLock() {
		fmt.Println("SUCCESS")
	} else {
		fmt.Println("FAILURE")
	}

	// Unclock mutex
	mutex.Unlock()

	// Third time, try should succeed again
	if mutex.TryLock() {
		fmt.Println("SUCCESS")
	} else {
		fmt.Println("FAILURE")
	}
}
```

The output is;

```sh
SUCCESS
FAILURE
SUCCESS
```

`ForceLock()` method is also availale for normal blocking lock.
