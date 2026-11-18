package src

import (
	"fmt"
	"net"
	"regexp"
	"slices"
	"strings"
)

/*
* Validates HTTP method requested by the user so it conforms to a HTTP /1.1
* standard method as defined in RFC 9112 and RFC 5789
 */
func ValidateHTTPMethod(method string) (HTTPMethod, error) {
	normalized := HTTPMethod(strings.ToUpper(method))

	switch normalized {
	case GET, POST, PUT, PATCH, DELETE:
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid HTTP method %q", method)
	}
}

// Validates an HTTP/HTTPS URI or an authority with an optional path and query.
// By default schema is set as HTTP
func ValidateHTTPHostAndPort(input string) (string, string, int, bool, error) {
	invalid := func() (string, string, int, bool, error) {
		return "", "", 0, false, fmt.Errorf("invalid HTTP URI %q", input)
	}

	lower := strings.ToLower(input)

	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		input = "http://" + input
	}

	u, err := ParseRFC3986URI(input)
	if err != nil || u.host == "" || strings.Contains(u.authority, "@") || strings.HasSuffix(u.authority, ":") {
		return invalid()
	}

	host := u.host
	if strings.HasPrefix(host, "[") {
		host = host[1 : len(host)-1]
		if net.ParseIP(host) == nil {
			return invalid()
		}
	}

	isPortUserProvided := false
	scheme := strings.ToLower(u.scheme)
	port := 80

	if scheme == "https" {
		port = 443
	}

	if u.port != -1 {
		port = u.port
		if port < 1 || port > 65535 {
			return invalid()
		}

		isPortUserProvided = true
	}

	return scheme, host, port, isPortUserProvided, nil
}

/* Validates HTTP request headers provided by the user input. Headers must follow
 * the format `<header-name>: <header-value>`. This method does not exclude nor
 * compress duplicated headers as they are processed at the request execution level.
 */
func ValidateHeaders(headers []string) error {

	// HTTP field names are tokens; values must not contain control characters.
	pattern := regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+:[^\\x00-\\x08\\x0a-\\x1f\\x7f]*$")

	for _, header := range headers {

		if !pattern.MatchString(header) {
			return fmt.Errorf("invalid header %q: expected <name>:<value> (space and value are optional)", header)
		}

		var sepIndex = strings.IndexByte(header, ':')

		var name = header[:sepIndex]
		var lowerName = strings.ToLower(name)

		if lowerName == "host" {
			return fmt.Errorf("invalid header %q: Host cannot be provided as a header and must be supplied as the last arg of the command", header)
		}

		if slices.Contains(FORBIDDEN_USER_SUPPLIED_HEADERS, lowerName) {
			return fmt.Errorf("invalid header %q: Forbidden header", header)
		}

	}

	return nil
}
