package main

type Config struct {
	ListenAddr string `env:"LISTEN_ADDR"`
	RootCA     string `env:"ROOT_CA"`
	ClientCert string `env:"CLIENT_CERT" default:"/etc/buggy/client.pem"`
	ClientKey  string `env:"CLIENT_KEY" default:"/etc/buggy/client.key"`
	RemoteUrl  string `env:"REMOTE_URL"`
}
