package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/randyardiansyah25/wsbase-handler"
)

func main() {
	//host := "localhost:8881"
	// host := "192.169.253.156:8881"
	host := "ws.digitalfstech.com"

	client := wsbase.NewWSClient(host, "/connect/0001", true)
	client.SetLogHandler(func(errType int, val string) {
		log.Println("ini log custom", val)
	})

	client.SetReconnectPeriod(5 * time.Second)

	client.SetMessageHandler(func(m wsbase.Message) {
		j, _ := json.MarshalIndent(m, "", "    ")
		log.Println("receive message : ")
		fmt.Println(string(j))
	})
	er := client.Start()
	if er != nil {
		log.Println("cannot start ws client, error :", er)
	}
}
