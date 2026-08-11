package main

import (
	"log"

	"nofx-trendln/runtimeapi"
)

func main() {
	if err := runtimeapi.Start(); err != nil {
		log.Printf("[trendln-runtime] 异常退出: %v", err)
	}
}
