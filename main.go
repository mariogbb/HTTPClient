package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"time"
)

func main() {

	// CMD argument flags

	var port = flag.Int("port", 443, "Connection port")
	var timeout = flag.Int("timeout", 10000, "Timeout (in milliseconds)")
	var host = flag.String("host", "example.com", "Connection host (example.com)")
	var method = flag.String("method", "GET", "HTTP Method")

	flag.Parse()

	// Start TCP connection (TLS handshake is already included with tls.Dial)

	var tcpConnectionString string = fmt.Sprintf("%s:%d", *host, *port)

	conn, err := tls.Dial("tcp", tcpConnectionString, nil)
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}

	defer conn.Close()

	// Avoid waiting indefinitely for the server.
	if err := conn.SetDeadline(time.Now().Add(time.Duration(*timeout) * time.Millisecond)); err != nil {
		fmt.Println("Deadline error:", err)
		return
	}

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	var httpReqMessage = fmt.Sprintf("%s / HTTP/1.1\r\nHost: %s\r\n\r\n", *method, *host)

	fmt.Println(httpReqMessage)

	_, err = writer.WriteString(httpReqMessage)
	if err != nil {
		fmt.Println("Write error:", err)
		return
	}

	if err := writer.Flush(); err != nil {
		fmt.Println("Flush error:", err)
		return
	}

	statusLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	fmt.Printf("Server status line: %s", statusLine)
}
