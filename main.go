package main

//GLOBAL
var config = readConfig() // config data
var mqttDataChannel = make(chan Message)
var tcpDataChannel = make(chan string)

func main(){
	// connect to mqtt
	mqttListen() // listen to mqtt broker targeted in config
	go processMqttData() // receive message via dataChannel and send out over tcp

	go receiveTCPData("127.0.0.1:8080") // read tcp messages
	processTcpData() // send tcp over mqtt

}
