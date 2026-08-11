package main

import (
	"log"

	"nofx-ta-lib-runtime/talib/runtimeapi"
)

func main() {
	if err := runtimeapi.Start(); err != nil {
		log.Printf("[ta-lib-runtime] 异常退出: %v", err)
	}
}
