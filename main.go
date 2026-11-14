package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	PATCH  HTTPMethod = "PATCH"
	DELETE HTTPMethod = "DELETE"
)

type ArgInfo struct {
	port    int
	timeout int
	host    string
	method  HTTPMethod
	isHttps bool
	headers []string
}

/*
 * Validates HTTP method requested by the user so it conforms to a HTTP /1.1
 * standard method as defined in
 */
func validateHTTPMethod(method string) (HTTPMethod, error) {
	normalized := HTTPMethod(strings.ToUpper(method))

	switch normalized {
	case GET, POST, PUT, PATCH, DELETE:
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid HTTP method %q", method)
	}
}

func validateHTTPHost(host string) string {
	return host
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

	var headers headerFlags

	var timeout = flag.Int("timeout", 10000, "Timeout (in milliseconds)")
	var method = flag.String("method", "GET", "HTTP Method")
	flag.Var(&headers, "header", "HTTP header")

	flag.Parse()

	// Require a host for the request
	if flag.NArg() != 1 {
		log.Fatal("expected exactly one host argument")
	}

	// Remove the http / https protocol in case the user introduces it
	// and flag if the request is http or https
	host := flag.Arg(0)

	scheme := regexp.MustCompile(`(?i)^https?://`).FindString(host)
	isHttps := strings.EqualFold(scheme, "https://")

	host = host[len(scheme):]
	parsedHost, err := url.Parse("//" + host)

	if err != nil || parsedHost.Hostname() == "" || parsedHost.Path != "" || parsedHost.RawQuery != "" || parsedHost.Fragment != "" || parsedHost.User != nil {
		log.Fatal("invalid host argument")
	}

	host = parsedHost.Hostname()
	host = validateHTTPHost(host)

	// Extract the defined port from the host string or set a default based
	// on HTTP or HTTPS connection types
	port := 443

	if portString := parsedHost.Port(); portString != "" {
		port, err = strconv.Atoi(portString)
		if err != nil || port < 1 || port > 65535 {
			log.Fatal("invalid port")
		}
	} else if !isHttps {

		port = 80

	}

	// Validate HTTP request method
	httpMethod, err := validateHTTPMethod(*method)

	if err != nil {
		log.Fatal(err)
	}

	return ArgInfo{port: port, timeout: *timeout, host: host, method: httpMethod, isHttps: isHttps, headers: []string(headers)}
}

func main() {

	// CMD argument flags

	var args = parseArgs()

	// Start TCP connection (TLS handshake is already included with tls.Dial)
	var tcpConnectionString string = net.JoinHostPort(args.host, strconv.Itoa(args.port))

	conn, err := tls.Dial("tcp", tcpConnectionString, nil)
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}

	defer conn.Close()

	// Avoid waiting indefinitely for the server.
	if err := conn.SetDeadline(time.Now().Add(time.Duration(args.timeout) * time.Millisecond)); err != nil {
		fmt.Println("Deadline error:", err)
		return
	}

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	var httpReqMessage = fmt.Sprintf("%s / HTTP/1.1\r\n", args.method)

	httpReqMessage += fmt.Sprintf("Host: %s\r\n", args.host)

	for _, header := range args.headers {

		httpReqMessage += fmt.Sprintf("%s\r\n", header)

	}

	httpReqMessage += "Connection: close\r\n\r\n"

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

	var resHeaders []string
	var body []string

	var isHeaderSection bool = true

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
	}

	fmt.Printf("Headers:\n")

	for _, hdr := range resHeaders {
		fmt.Printf("%s", hdr)
	}

	fmt.Printf("Body:\n")

	for _, bline := range body {
		fmt.Printf("%s", bline)
	}

}
