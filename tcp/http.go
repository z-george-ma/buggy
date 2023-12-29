package tcp

import (
	"errors"
	"strings"
)

type HttpRequestHeader struct {
	Method  string
	Url     string
	Version string
	Headers map[string]string
}

var ErrHttpMalformedHeader = errors.New("Malformed HTTP Header")

func ParseHttpHeader(conn Conn) (ret HttpRequestHeader, err error) {
	var buf, b []byte
	ret = HttpRequestHeader{
		Headers: map[string]string{},
	}

	for {
		b, err = conn.ReadSlice('\n', &buf)
		if err != nil {
			err = ErrHttpMalformedHeader
			return
		}

		l := len(b)

		if b[l-2] != '\r' {
			err = ErrHttpMalformedHeader
			return
		}

		b = b[:l-2]

		bs := string(b)

		if ret.Method == "" {
			ss := strings.Split(bs, " ")
			if len(ss) != 3 {
				err = ErrHttpMalformedHeader
				return
			}

			ret.Method = ss[0]
			ret.Url = ss[1]
			ret.Version = ss[2]

			if len(ret.Method) == 0 || len(ret.Url) == 0 || len(ret.Version) == 0 || !strings.HasPrefix(ret.Version, "HTTP/") {
				err = ErrHttpMalformedHeader
				return
			}

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
		ret.Headers[strings.ToLower(s[0])] = s[1]
	}

	return
}
