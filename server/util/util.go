package util

import (
	"fmt"
	"os"
)

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func RaiseCustomError(msg string) {
	fmt.Fprintf(os.Stderr, "Fatal error: %s", msg)
	os.Exit(1)
}
