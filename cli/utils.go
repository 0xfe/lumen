package cli

import (
	"fmt"
	"os"
)

func showSuccess(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func showError(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args)
}
