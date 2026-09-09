package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	blocks := make([][]byte, 0, 128)
	deadline := time.Now().Add(4 * time.Minute)
	for time.Now().Before(deadline) {
		block := make([]byte, 1024*1024)
		for index := range block {
			block[index] = byte(index)
		}
		blocks = append(blocks, block)
		if len(blocks) >= 128 {
			blocks = blocks[1:]
			runtime.GC()
		}
		fmt.Printf("bloques_mib=%d\n", len(blocks))
		time.Sleep(time.Second)
	}
}
