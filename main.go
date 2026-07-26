package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "smtp.gmail.com:587")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read server greeting
	greeting, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	fmt.Print(greeting)

	// Send EHLO
	_, err = fmt.Fprintf(conn, "EHLO localhost\r\n")
	if err != nil {
		panic(err)
	}

	// Read all EHLO response lines
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		fmt.Print(line)

		// Last line starts with "250 "
		if len(line) >= 4 && line[:4] == "250 " {
			break
		}
	}
}
