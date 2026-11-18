package src

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var uriPattern = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9+.-]*):(?://([^/?#]*))?([^?#]*)(?:\?([^#]*))?(?:#(.*))?$`)
var authorityPattern = regexp.MustCompile(`^(?:((?:[A-Za-z0-9._~!$&'()*+,;=:-]|%[0-9A-Fa-f]{2})*)@)?(\[[^\]]+\]|(?:[A-Za-z0-9._~!$&'()*+,;=-]|%[0-9A-Fa-f]{2})*)(?::([0-9]*))?$`)
var ipvFuturePattern = regexp.MustCompile(`^[vV][0-9A-Fa-f]+\.[A-Za-z0-9._~!$&'()*+,;=:-]+$`)
var pathPattern = regexp.MustCompile(`^(?:[A-Za-z0-9._~!$&'()*+,;=:@/-]|%[0-9A-Fa-f]{2})*$`)
var queryFragmentPattern = regexp.MustCompile(`^(?:[A-Za-z0-9._~!$&'()*+,;=:@/?-]|%[0-9A-Fa-f]{2})*$`)
var ipv6GroupPattern = regexp.MustCompile(`^[0-9A-Fa-f]{1,4}$`)
var ipv4PartPattern = regexp.MustCompile(`^(?:0|[1-9][0-9]{0,2})$`)

// validIPv6 checks the IPv6address grammar, including an embedded IPv4 address.
func validIPv6(address string) bool {

	if !strings.Contains(address, ":") || strings.Count(address, "::") > 1 {
		return false
	}

	compressed := strings.Contains(address, "::")
	parts := strings.Split(address, "::")
	groups := 0

	for i, part := range parts {

		if part == "" {

			if !compressed {
				return false
			}

			continue
		}

		fields := strings.Split(part, ":")

		for j, field := range fields {

			if strings.Contains(field, ".") {

				if i != len(parts)-1 || j != len(fields)-1 {
					return false
				}

				octets := strings.Split(field, ".")

				if len(octets) != 4 {
					return false
				}

				for _, octet := range octets {

					if !ipv4PartPattern.MatchString(octet) {
						return false
					}

					n, _ := strconv.Atoi(octet)

					if n > 255 {
						return false
					}

				}

				groups += 2

			} else if !ipv6GroupPattern.MatchString(field) {

				return false

			} else {

				groups++

			}
		}
	}

	return (compressed && groups < 8) || (!compressed && groups == 8)
}

/*
RFC 3986 compliant URI struct.
See https://www.rfc-editor.org/info/rfc3986
*/
type RFC3986URI struct {
	scheme    string
	authority string
	userInfo  string
	host      string
	port      int
}

func ParseRFC3986URI(uri string) (RFC3986URI, error) {

	invalid := RFC3986URI{port: -1}
	parts := uriPattern.FindStringSubmatch(uri)

	if parts == nil ||
		!pathPattern.MatchString(parts[3]) ||
		!queryFragmentPattern.MatchString(parts[4]) ||
		!queryFragmentPattern.MatchString(parts[5]) {

		return invalid, fmt.Errorf("invalid URI %q: invalid syntax, path, query, or fragment", uri)
	}

	result := RFC3986URI{scheme: parts[1], authority: parts[2], port: -1}

	if !strings.HasPrefix(uri[len(parts[1])+1:], "//") {
		return result, nil
	}

	authority := authorityPattern.FindStringSubmatch(parts[2])

	if authority == nil {
		return invalid, fmt.Errorf("invalid URI %q: invalid authority", uri)
	}

	result.userInfo, result.host = authority[1], authority[2]

	if strings.HasPrefix(result.host, "[") {

		literal := result.host[1 : len(result.host)-1]

		if !ipvFuturePattern.MatchString(literal) && !validIPv6(literal) {
			return invalid, fmt.Errorf("invalid URI %q: invalid IP literal", uri)
		}

	}

	if authority[3] != "" {

		port, err := strconv.Atoi(authority[3])

		if err != nil {
			return invalid, fmt.Errorf("invalid URI %q: invalid port: %w", uri, err)
		}

		result.port = port
	}

	return result, nil
}
