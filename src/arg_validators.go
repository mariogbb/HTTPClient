package src

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"slices"
	"strconv"
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

	// URI characters and percent escapes
	uriPattern := regexp.MustCompile(`^(?i:(?:https?://)?)(?:\[[0-9A-Fa-f:.]+\]|(?:[A-Za-z0-9._~!$&'()*+,;=-]|%[0-9A-Fa-f]{2})+)(?::[0-9]+)?(?:/(?:[A-Za-z0-9._~!$&'()*+,;=:@/-]|%[0-9A-Fa-f]{2})*)?(?:\?(?:[A-Za-z0-9._~!$&'()*+,;=:@/?-]|%[0-9A-Fa-f]{2})*)?(?:#(?:[A-Za-z0-9._~!$&'()*+,;=:@/?-]|%[0-9A-Fa-f]{2})*)?$`)

	if !uriPattern.MatchString(input) {
		return invalid()
	}

	lower := strings.ToLower(input)

	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		input = "http://" + input
	}

	// @TODO: Implement a custom validator for URL. For now it works as a way
	// to extract URL data
	u, err := url.Parse(input)

	if err != nil || u.Hostname() == "" {
		return invalid()
	}

	if strings.HasPrefix(u.Host, "[") && net.ParseIP(u.Hostname()) == nil {
		return invalid()
	}

	isPortUserProvided := false
	scheme := strings.ToLower(u.Scheme)
	port := 80

	if scheme == "https" {
		port = 443
	}

	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())

		if err != nil || port < 1 || port > 65535 {
			return invalid()
		}

		isPortUserProvided = true
	}

	return scheme, u.Hostname(), port, isPortUserProvided, nil
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
