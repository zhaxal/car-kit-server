package app

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	Protocol string
	IP       []byte
	Port     int
}

func (t *Server) New(callBack func(udpc *net.UDPConn, buf *[]byte, len int, addr *net.UDPAddr)) {

	udpc, err := net.ListenUDP(t.Protocol, &net.UDPAddr{IP: t.IP, Port: t.Port, Zone: ""})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer udpc.Close()

	fmt.Printf("Listening on %v\n", udpc.LocalAddr())

	for {

		buf := make([]byte, 4096)
		read, addr, err := udpc.ReadFromUDP(buf)
		if err != nil {
			log.Print("error when listening ", err)
			continue
		}

		sliced := buf[:read]

		fmt.Printf("New connection from %v , fired a new goroutine \n", addr)

		go callBack(udpc, &sliced, read, addr)

	

		fmt.Println("went offline")
	}
}
