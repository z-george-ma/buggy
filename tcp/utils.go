package tcp

import "net/url"

type NetworkAddress struct {
	Scheme  string
	Host    string
	Port    string
	Address string
}

func UrlToAddress(addr string) (na NetworkAddress, err error) {
	u, err := url.Parse(addr)

	if err != nil {
		return
	}

	host := u.Hostname()
	port := u.Port()
	formattedAddr := u.Host

	if port == "" {
		switch u.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}

		formattedAddr += ":" + port
	}

	na = NetworkAddress{
		Scheme:  u.Scheme,
		Host:    host,
		Port:    port,
		Address: formattedAddr,
	}

	return
}
