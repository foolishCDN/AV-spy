package utils

import (
	"fmt"
	"time"
)

// Elapsed
// Usage:
//
//  func foo() {
//     defer elapsed("page")()  // <-- The trailing () is the deferred call
//     // code to measure
//	}
//
// see https://stackoverflow.com/questions/45766572/is-there-an-efficient-way-to-calculate-execution-time-in-golang/45766707
func Elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}
