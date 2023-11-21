package main

//GLOBAL
var config = readConfig() // config data
var dataChannel = make(chan Message)

func main(){
	// connect to mqtt
	mqttListen() // listen to mqtt broker targeted in config
	go processData() // receive message via dataChannel and send out over tcp
	startTCPServer("127.0.0.1:8080") // read tcp messages
}
