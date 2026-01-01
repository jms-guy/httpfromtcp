package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connection accepted")

		messages := getLinesChannel(conn)

		for m := range messages {
			fmt.Println(m)
		}
		fmt.Println("Connection closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	messChannel := make(chan string)

	go func() {
		defer close(messChannel)
		messages := make([]byte, 8)
		currLine := ""

		for {
			clear(messages)
			numBytes, err := f.Read(messages)
			if err != nil {
				if err == io.EOF {
					return
				} else {
					log.Fatal(err)
				}
			}
			if numBytes > 0 {
				parts := strings.Split(string(messages), "\n")
				currLine += parts[0]
				if len(parts) > 1 {
					messChannel <- currLine
					currLine = parts[1]
				}
			}
			if numBytes < 8 {
				messChannel <- currLine
			}
		}
	}()

	return messChannel
}
