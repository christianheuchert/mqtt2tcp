package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func mqttListen() {
	client := mqttConnect("mqttRelay", config.BrokerURL, config.BrokerPort)
	for _, topic := range config.BrokerTopics {
		go subscribeAndSend(client, topic)
	}

	// Wait for some time to allow Goroutines to subscribe
	time.Sleep(2 * time.Second)
}
func subscribeAndSend(client mqtt.Client, topic string) {
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		// Send the data along with the topic information to the channel
		dataChannel <- Message{Topic: topic, Data: string(msg.Payload())}
	})
}
func processData() {
	for {
		select {
		case message := <-dataChannel:
			// Process the data based on the topic
			// fmt.Printf("Received data from topic %s: %s\n", message.Topic, message.Data)
			err := sendDataOverTCP(message)
			if err !=nil{
				println("tcp error: ", err)
			}

		}
	}
}

func mqttConnect(clientId string, host string, port string) mqtt.Client {
	opts := mqttCreateClientOptions(clientId, host, port)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}
func mqttCreateClientOptions(clientId string, host string, port string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ws://%s:%s", host, port))
	opts.SetClientID(clientId)
	opts.SetUsername(config.BrokerUsername)
	opts.SetPassword(config.BrokerPassword)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	return opts
}
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("MQTT Connected")
}
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

func readConfig() Config {
	var config Config

	// Read the JSON file.
    jsonFile, err := os.ReadFile("config.json")
    if err != nil {
        fmt.Println(err)
        return config
    }

    // Decode the JSON file into the config struct.
    err = json.Unmarshal(jsonFile, &config)
    if err != nil {
        fmt.Println(err)
        return config
    }

	return config
}

func sendDataOverTCP(message Message) error {
	// Connect to the TCP server
	conn, err := net.Dial("tcp", config.TcpAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send data to the server
	fmt.Fprintf(conn, "%s\n", message.Data)

	return nil
}

// Listen to Tcp port with helper fxn handleConnection
func startTCPServer(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error starting TCP server: %v", err)
	}

	defer listener.Close()
	fmt.Printf("TCP server listening on %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}
		fmt.Printf("Received message: %s", message)
	}
}

