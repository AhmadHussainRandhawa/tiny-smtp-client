package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
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

	// Read the attachment
	attachmentPath := "attachments/image.png"

	fileData, err := os.ReadFile(attachmentPath)
	if err != nil {
		panic(err)
	}

	// Convert binary file data into Base64 text
	encodedFile := base64.StdEncoding.EncodeToString(fileData)

	// MIME message
	// MIME message
	boundary := "BOUNDARY_123"

	message := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: Image from my Go SMTP Client\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: multipart/mixed; boundary=\"%s\"\r\n"+
			"\r\n"+
			"--%s\r\n"+
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n"+
			"\r\n"+
			"Assalam u Alaikum!\r\n"+
			"\r\n"+
			"This email contains an image attachment sent manually by my\r\n"+
			"own SMTP client written in Go.\r\n"+
			"\r\n"+
			"Regards,\r\n"+
			"Ahmad\r\n"+
			"\r\n"+
			"--%s\r\n"+
			"Content-Type: image/png; name=\"%s\"\r\n"+
			"Content-Transfer-Encoding: base64\r\n"+
			"Content-Disposition: attachment; filename=\"%s\"\r\n"+
			"\r\n"+
			"%s\r\n"+
			"\r\n"+
			"--%s--\r\n",
		email,
		recipient,
		boundary,
		boundary,
		boundary,
		filepath.Base(attachmentPath),
		filepath.Base(attachmentPath),
		encodedFile,
		boundary,
	)

	// End SMTP DATA using <CRLF>.<CRLF>.
	fmt.Fprintf(tlsConn, "%s.\r\n", message)
	readSMTPResponse(reader)

	// QUIT
	fmt.Println("\n>>> QUIT")
	fmt.Fprintf(tlsConn, "QUIT\r\n")
	readSMTPResponse(reader)
}
