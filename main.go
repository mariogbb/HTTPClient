package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mariogbb/HTTPClient/src"
)

type ArgInfo struct {
	port               int
	isPortUserProvided bool
	timeout            int
	host               string
	schema             string
	method             src.HTTPMethod
	headers            []string
}

type headerFlags []string

func (h *headerFlags) String() string {
	return strings.Join(*h, ", ")
}

func (h *headerFlags) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func parseArgs() ArgInfo {

	// Register flags before processing
	var headers headerFlags
	flag.Var(&headers, "header", "HTTP header")

	var timeout = flag.Int("timeout", 10000, "Timeout (in milliseconds)")
	var method = flag.String("method", "GET", "HTTP Method")

	flag.Parse()

	// Require a host for the request. Host is always set as the last argument
	// and declared without using a flag
	if flag.NArg() != 1 {
		log.Fatal("expected exactly one host argument")
	}

	// Remove the http / https protocol in case the user introduces it
	// and flag if the request is http or https
	schema, host, port, isPortUserProvided, err := src.ValidateHTTPHostAndPort(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	// Validate request headers supplied by the user
	err = src.ValidateHeaders([]string(headers))

	if err != nil {
		log.Fatal(err)
	}

	// Validate HTTP request method
	httpMethod, err := src.ValidateHTTPMethod(*method)

	if err != nil {
		log.Fatal(err)
	}

	if err := src.ValidateHeaders(headers); err != nil {
		log.Fatal(err)
	}

	return ArgInfo{port: port, timeout: *timeout, host: host, method: httpMethod, schema: schema, headers: []string(headers), isPortUserProvided: isPortUserProvided}
}

func stablishTCPConnection(args ArgInfo) (*tls.Conn, *bufio.Writer, *bufio.Reader) {
	// Start TCP connection (TLS handshake is already included with tls.Dial)
	var tcpConnectionString string = net.JoinHostPort(args.host, strconv.Itoa(args.port))

	// @TODO: For now HTTPS connection is the only way to make requests.
	// An implementation with "net.Dial" for http:// schema will allow
	// the use of HTTP connections
	conn, err := tls.Dial("tcp", tcpConnectionString, nil)

	if err != nil {
		log.Fatalln("Connection error:", err)
	}

	// Avoid waiting indefinitely for the server.
	if err := conn.SetDeadline(time.Now().Add(time.Duration(args.timeout) * time.Millisecond)); err != nil {
		log.Fatalln("Deadline error:", err)
	}

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	return conn, writer, reader
}

func sendHTTPRequest(args ArgInfo, writer *bufio.Writer) {
	var httpReqMessage = fmt.Sprintf("%s / HTTP/1.1\r\n", args.method)

	// Keep URL authority in case user provides or doesn't provide a port
	if args.isPortUserProvided {
		httpReqMessage += fmt.Sprintf("Host: %s:%d\r\n", args.host, args.port)
	} else {
		httpReqMessage += fmt.Sprintf("Host: %s\r\n", args.host)
	}

	for _, header := range args.headers {

		httpReqMessage += fmt.Sprintf("%s\r\n", header)

	}

	httpReqMessage += "Connection: close\r\n\r\n"

	fmt.Println(httpReqMessage)

	_, err := writer.WriteString(httpReqMessage)

	if err != nil {
		log.Fatalln("Write error:", err)
	}

	if err := writer.Flush(); err != nil {
		log.Fatalln("Flush error:", err)
	}

}

func processHTTPResponse(args ArgInfo, reader *bufio.Reader) {

	var resHeaders []string
	var body []string

	var isHeaderSection bool = true
	// var contentLength int = 100000

	for {

		line, err := reader.ReadString('\n')

		if line == "\r\n" {
			isHeaderSection = false
		}

		if isHeaderSection {
			resHeaders = append(resHeaders, line)
		} else {
			body = append(body, line)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Read error:", err)
			return
		}
	}

	for _, hdr := range resHeaders {
		fmt.Printf("%s", hdr)
	}

	fmt.Printf("Body:\n")

	for _, bline := range body {
		fmt.Printf("%s", bline)
	}
}

func main() {

	// CMD argument flags
	var args = parseArgs()

	// Start TCP connection with the origin host
	conn, writer, reader := stablishTCPConnection(args)

	// Close connection at the end of main function
	defer conn.Close()

	// Send HTTP 1.1 request with the required parameters
	sendHTTPRequest(args, writer)

	// Process the HTTP 1.1 response
	processHTTPResponse(args, reader)

}
