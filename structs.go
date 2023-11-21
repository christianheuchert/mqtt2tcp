package main

// Needs to up to date with config.json
type Config struct {
    BrokerURL string
    BrokerPort string
    BrokerPassword string
    BrokerUsername string
	BrokerTopics []string
	TcpAddress string
}

// mqtt message format
type Message struct {
	Topic string
	Data  string
}