package main

import (
	"log"

	"nofx-pandas-ta/runtimeapi"
)

func main() {
	if err := runtimeapi.Start(); err != nil {
		log.Printf("[pandas-ta-runtime] 异常退出: %v", err)
	}
}
