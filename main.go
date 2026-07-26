package main

import (
	"fmt"
	"net"
)

func main() {

	conn, err := net.Dial("tcp", "smtp.gmail.com:587")

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(buffer[:n]))

}
