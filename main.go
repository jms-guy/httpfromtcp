package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal(err)
	}

	messages := make([]byte, 8)

	for {
		numBytes, err := f.Read(messages)
		if numBytes > 0 {
			fmt.Printf("read: %s\n", string(messages)[:numBytes])
		}
		if err != nil {
			if err == io.EOF {
				os.Exit(0)
			} else {
				log.Fatal(err)
			}
		}
	}
}
