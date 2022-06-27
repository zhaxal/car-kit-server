package main

import (
	"car-kit-server/app"
	"context"
	"fmt"
	"log"
	"net"

	"github.com/filipkroca/teltonikaparser"
)

func main() {

	server := app.Server{
		Protocol: "udp",
		IP:       []byte{0, 0, 0, 0},
		Port:     8080,
	}

	//create new server
	server.New(onUDPMessage)
	defer fmt.Println("server closed")

}

//onUDPMessage is invoked when packet arrive
func onUDPMessage(udpc *net.UDPConn, dataBs *[]byte, len int, addr *net.UDPAddr) {
	//conn := *udpc

	ctx := context.Background()
	client := app.GetFirestore()
	defer client.Close()

	parsedData, err := teltonikaparser.Decode(dataBs)
	if err != nil {
		log.Print("Unable to decode packet", err)
	}

	if err == nil {
		app.AddDevice(parsedData, ctx, client)
	}

	(*udpc).WriteToUDP(parsedData.Response, addr)

}
