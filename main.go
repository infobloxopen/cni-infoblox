package main

import (
	"os"
)

func main() {
	if len(os.Args) > 1 {
		config := LoadConfig()
		runDaemon(config)
	} else {
		runPlugin()
	}
}
