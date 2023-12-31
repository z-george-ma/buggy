package tcp

import (
	"errors"
	"strconv"
	"strings"
)

type HttpRequest struct {
	Method  string
	Url     string
	Version string
	Headers map[string]string
}

type HttpResponse struct {
	StatusCode int
	Reason     string
	Version    string
	Headers    map[string]string
}

var MaxHeadersSupported = 100
var ErrHttpMalformedHeader = errors.New("Malformed HTTP Header")
var ErrExceedingHeaderCount = errors.New("Exceeding max number of HTTP headers")

func parseHttp(r Reader, parseStartLine func(string) error) (headers map[string]string, err error) {
	var buf, b []byte

	startLine := true
	headers = map[string]string{}

	for {
		b, err = r.ReadSlice('\n', &buf)
		if err != nil {
			return
		}

		l := len(b)

		if b[l-2] != '\r' {
			err = ErrHttpMalformedHeader
			return
		}

		b = b[:l-2]

		bs := string(b)

		if startLine {
			if err = parseStartLine(bs); err != nil {
				return
			}

			startLine = false
			continue
		}

		if len(b) == 0 {
			// HTTP header parse complete
			break
		}

		s := strings.Split(bs, ": ")
		if len(s) != 2 {
			err = ErrHttpMalformedHeader
			return
		}
		headers[strings.ToLower(s[0])] = s[1]

		if len(headers) > MaxHeadersSupported {
			err = ErrExceedingHeaderCount
			return
		}
	}
	return
}

func ParseHttpRequest(r Reader) (ret HttpRequest, err error) {
	ret = HttpRequest{}

	ret.Headers, err = parseHttp(r, func(s string) error {
		ss := strings.Split(s, " ")
		if len(ss) != 3 {
			return ErrHttpMalformedHeader
		}

		ret.Method = ss[0]
		ret.Url = ss[1]
		ret.Version = ss[2]

		if len(ret.Method) == 0 || len(ret.Url) == 0 || !strings.HasPrefix(ret.Version, "HTTP/") {
			return ErrHttpMalformedHeader
		}
		return nil
	})

	return
}

func ParseHttpResponse(r Reader) (ret HttpResponse, err error) {
	ret = HttpResponse{}

	ret.Headers, err = parseHttp(r, func(s string) error {
		ss := strings.Split(s, " ")
		if len(ss) != 3 {
			return ErrHttpMalformedHeader
		}

		ret.Version = ss[0]
		ret.Reason = ss[2]

		if len(ss[1]) != 3 || !strings.HasPrefix(ret.Version, "HTTP/") {
			return ErrHttpMalformedHeader
		}

		ret.StatusCode, err = strconv.Atoi(ss[1])
		return err
	})

	return
}
