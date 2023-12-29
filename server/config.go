package main

type Config struct {
	ListenAddr   string `env:"LISTEN_ADDR"`
	ClientRootCA string `env:"CLIENT_ROOT_CA"`
	ServerCert   string `env:"SERVER_CERT" default:"/etc/buggy/server.pem"`
	ServerKey    string `env:"SERVER_KEY" default:"/etc/buggy/server.key"`
}
