package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
)

func readSMTPResponse(reader *bufio.Reader) ([]string, error) {
	var lines []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")
		lines = append(lines, line)

		fmt.Println(line)

		if len(line) >= 4 && line[3] == ' ' {
			break
		}
	}

	return lines, nil
}

func main() {

	conn, err := net.Dial("tcp", "smtp.gmail.com:587")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Server Greeting
	_, err = readSMTPResponse(reader)
	if err != nil {
		panic(err)
	}

	fmt.Println()

	// First EHLO
	fmt.Fprintf(conn, "EHLO localhost\r\n")

	lines, err := readSMTPResponse(reader)
	if err != nil {
		panic(err)
	}

	// Verify STARTTLS support
	startTLS := false
	for _, line := range lines {
		if strings.Contains(line, "STARTTLS") {
			startTLS = true
			break
		}
	}

	if !startTLS {
		panic("Server does not support STARTTLS")
	}

	fmt.Println()
	fmt.Println(">>> STARTTLS")

	fmt.Fprintf(conn, "STARTTLS\r\n")

	_, err = readSMTPResponse(reader)
	if err != nil {
		panic(err)
	}

	// Upgrade existing connection
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: "smtp.gmail.com",
	})

	err = tlsConn.Handshake()
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println("TLS Handshake Successful")
	fmt.Println()

	reader = bufio.NewReader(tlsConn)

	// RFC requires EHLO again after STARTTLS
	fmt.Fprintf(tlsConn, "EHLO localhost\r\n")

	_, err = readSMTPResponse(reader)
	if err != nil {
		panic(err)
	}
}
