package src

type HTTPMethod string

const (
	GET     HTTPMethod = "GET"
	HEAD    HTTPMethod = "HEAD"
	POST    HTTPMethod = "POST"
	PUT     HTTPMethod = "PUT"
	DELETE  HTTPMethod = "DELETE"
	CONNECT HTTPMethod = "CONNECT"
	OPTIONS HTTPMethod = "OPTIONS"
	PATCH   HTTPMethod = "PATCH" // Introduced in RFC 5789
)
