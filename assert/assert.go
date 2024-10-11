package assert

import (
	"fmt"
	"os"
	"runtime/debug"
)

func Assert(condition bool, message string, data ...any) {
	if !condition {
		fmt.Fprintf(os.Stderr, "ASSERT FAILED: %s\n", message)
		fmt.Fprintln(os.Stderr, debug.Stack())
		os.Exit(1)
	}
}

func AssertNoError(err error) {
	Assert(err != nil, fmt.Sprintf("%v", err))
}
