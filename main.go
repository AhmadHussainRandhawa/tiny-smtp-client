package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"os"
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

	email := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")

	if email == "" || password == "" {
		panic("SMTP_EMAIL or SMTP_PASSWORD not set")
	}

	conn, err := net.Dial("tcp", "smtp.gmail.com:587")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	readSMTPResponse(reader)

	fmt.Fprintf(conn, "EHLO localhost\r\n")
	lines, _ := readSMTPResponse(reader)

	startTLS := false
	for _, line := range lines {
		if strings.Contains(line, "STARTTLS") {
			startTLS = true
			break
		}
	}

	if !startTLS {
		panic("STARTTLS not supported")
	}

	fmt.Fprintf(conn, "STARTTLS\r\n")
	readSMTPResponse(reader)

	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: "smtp.gmail.com",
	})

	if err := tlsConn.Handshake(); err != nil {
		panic(err)
	}

	reader = bufio.NewReader(tlsConn)

	fmt.Fprintf(tlsConn, "EHLO localhost\r\n")
	readSMTPResponse(reader)

	// AUTH LOGIN
	fmt.Println("\n>>> AUTH LOGIN")

	fmt.Fprintf(tlsConn, "AUTH LOGIN\r\n")
	readSMTPResponse(reader)

	// Username
	username := base64.StdEncoding.EncodeToString([]byte(email))
	fmt.Fprintf(tlsConn, "%s\r\n", username)
	readSMTPResponse(reader)

	// Password
	pass := base64.StdEncoding.EncodeToString([]byte(password))
	fmt.Fprintf(tlsConn, "%s\r\n", pass)
	readSMTPResponse(reader)

	fmt.Println("\nAuthentication Complete")
}
