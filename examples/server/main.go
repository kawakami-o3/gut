package main

import (
	"github/com/kawakami-o3/gut"
	"net"
)

func main() {
	server := gut.Server{
		udpAddr: &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 9000,
		},
	}

	server.Run()
}
