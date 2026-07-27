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

	// Verify STARTTLS support
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

	// Upgrade existing connection
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
	fmt.Fprintf(tlsConn, "AUTH LOGIN\r\n")
	readSMTPResponse(reader)

	username := base64.StdEncoding.EncodeToString([]byte(email))
	fmt.Fprintf(tlsConn, "%s\r\n", username)
	readSMTPResponse(reader)

	pass := base64.StdEncoding.EncodeToString([]byte(password))
	fmt.Fprintf(tlsConn, "%s\r\n", pass)
	readSMTPResponse(reader)

	// MAIL FROM
	fmt.Println("\n>>> MAIL FROM")
	fmt.Fprintf(tlsConn, "MAIL FROM:<%s>\r\n", email)
	readSMTPResponse(reader)

	// RCPT TO
	recipient := os.Getenv("SMTP_RCPT") // Change if you want

	fmt.Println("\n>>> RCPT TO")
	fmt.Fprintf(tlsConn, "RCPT TO:<%s>\r\n", recipient)
	readSMTPResponse(reader)

	// DATA
	fmt.Println("\n>>> DATA")
	fmt.Fprintf(tlsConn, "DATA\r\n")
	readSMTPResponse(reader)

	// Message
	message := `From: ` + email + `
To: ` + recipient + `
Subject: Tiny SMTP Client

Assalamu Alaikum!

This email was sent manually using my own SMTP client written in Go.

Regards,
Ahmad
`

	// SMTP requires CRLF line endings
	message = strings.ReplaceAll(message, "\n", "\r\n")

	fmt.Fprintf(tlsConn, "%s\r\n.\r\n", message)
	readSMTPResponse(reader)

	// QUIT
	fmt.Println("\n>>> QUIT")
	fmt.Fprintf(tlsConn, "QUIT\r\n")
	readSMTPResponse(reader)
}
