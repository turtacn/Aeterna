package main

import (
	"fmt"
	"os"

	"github.com/turtacn/Aeterna/internal/cli"
	"github.com/turtacn/Aeterna/pkg/logger"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			if logger.Log != nil {
				logger.Log.Error("Panic recovered", "panic", r)
			} else {
				fmt.Fprintf(os.Stderr, "Panic recovered: %v\n", r)
			}
			os.Exit(1)
		}
	}()

	cli.Execute()
}

// Personal.AI order the ending
