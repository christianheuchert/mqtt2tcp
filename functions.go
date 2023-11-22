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

// MQTT MQTT MQTT MQTT MQTT ------------------------------------------

// listen to mqtt topics and send to data channel
func mqttListen() {
	client := mqttConnect("mqttRelay", config.BrokerURL, config.BrokerPort)
	for _, topic := range config.BrokerTopics {
		go subscribeMqttSendToChannel(client, topic)
	}

	// Wait for some time to allow Goroutines to subscribe
	time.Sleep(2 * time.Second)
}
func subscribeMqttSendToChannel(client mqtt.Client, topic string) {
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		// Send the data along with the topic information to the channel
		mqttDataChannel <- Message{Topic: topic, Data: string(msg.Payload())}
	})
}

// publish
func mqttPublish(publishClient mqtt.Client, topic string, data string){
	publishClient.Publish(topic, 0, false, data)
}

// Handle Received mqtt data
func processMqttData() {
	for {
		select {
		case message := <-mqttDataChannel:
			// Process the data based on the topic
			// fmt.Printf("Received data from topic %s: %s\n", message.Topic, message.Data)
			err := sendTCPData(message)
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

// TCP TCP TCP TCP TCP ------------------------------------------

func sendTCPData(message Message) error {
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
func receiveTCPData(address string) {
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
		// handle tcp message
		//fmt.Printf("Received tcp message: %s", message)
		tcpDataChannel <- message
	}
}

// Handle Received tcp data
func processTcpData() {
	publishClient := mqttConnect("mqttPublish", config.BrokerURL, config.BrokerPort)
	for {
		select {
		case message := <-tcpDataChannel:
			// Process the data based on the topic
			fmt.Printf("Received data from tcp connection: %s\n", message)
			publishClient.Publish("mqtt2tcp2mqtt", 0, false, message)
		}
	}
}

// MISC MISC MISC MISC MISC ------------------------------------------
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