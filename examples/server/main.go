package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/k0kubun/pp"
)

var endian = binary.LittleEndian

type UdpHeader struct {
	Type  byte
	Flags byte
	Token uint64
}

type ConnectRequest struct {
	UdpHeader
}

type ConnectAccept struct {
	UdpHeader
	NewToken uint64
}

type UdpData struct {
	UdpHeader
	Num uint32
}

type Ping struct {
	UdpHeader
}

type Pong struct {
	UdpHeader
}

func recvAny(conn net.Conn) []byte {
	//log.Println("====== start: recv ANY")
	bs := make([]byte, 256)
	n, err := conn.Read(bs)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	//log.Printf("Recd data: %v", bs[:n])
	return bs[:n]
}

func recvPing(conn net.Conn) {
	log.Println("====== start: recv ping")
	bs := make([]byte, 16)
	var ping Ping
	n, err := conn.Read(bs)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	//log.Printf("Recd data: %v", bs[:n])
	buf := bytes.NewBuffer(bs[:n])
	binary.Read(buf, endian, &ping)
	//log.Println("RECV pong: %v %v", n, ping)
}

func recvPong(conn net.Conn) {
	log.Println("====== start: recv pong")
	bs := make([]byte, 16)
	var pong Pong
	n, err := conn.Read(bs)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	//log.Printf("Recd data: %v", bs[:n])
	buf := bytes.NewBuffer(bs[:n])
	binary.Read(buf, endian, &pong)
	//log.Println("RECV pong: %v %v", n, pong)
}


const (
	ConnectionRequestType = iota
	ConnectionRejectType
	ConnectionAcceptType
	DisconnectType
	DataType
	PingType
	PongType
)

func sendAccept(conn *net.UDPConn, addr *net.UDPAddr, bs []byte) {
	var request ConnectRequest
	binary.Read(bytes.NewBuffer(bs), endian, &request)

	var accept ConnectAccept
	accept.Type = 2
	accept.Flags = 1

	RECV_TOKEN = request.Token
	SEND_TOKEN = rand.Uint64()

	accept.Token = RECV_TOKEN
	accept.NewToken = SEND_TOKEN
	buf := new(bytes.Buffer)
	binary.Write(buf, endian, accept)

	bout := buf.Bytes()
	n, err := conn.WriteTo(bout, addr)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	log.Printf("send size: %v", n)

	guestAddr = addr
}

func sendData(conn *net.UDPConn, addr *net.UDPAddr, bs []byte) {
	log.Printf("recv bytes: %v", bs)

	var udpData UdpData
	buf := bytes.NewBuffer(bs)
	binary.Read(buf, endian, &udpData)

	var response UdpData
	response.Type = 4
	response.Flags = 0
	response.Token = RECV_TOKEN

	response.Num = udpData.Num + 1000

	buf = new(bytes.Buffer)
	binary.Write(buf, endian, response)

	bout := buf.Bytes()
	log.Printf("send bytes: %v", bout)
	n, err := conn.WriteTo(bout, addr)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	log.Printf("send size: %v", n)

	guestAddr = addr
}

func sendPing(conn *net.UDPConn, addr *net.UDPAddr) {
	log.Println("=== start ping")
	//conn, err := net.DialUDP("udp", nil, addr)
	//if err != nil {
	//	log.Fatalln(err)
	//	os.Exit(1)
	//}

	var ping Ping
	ping.Type = 5
	ping.Flags = 0
	ping.Token = RECV_TOKEN

	buf := new(bytes.Buffer)
	binary.Write(buf, endian, ping)
	n, err := conn.WriteTo(buf.Bytes(), addr)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	log.Printf("=== end ping %v", n)
}

func sendPong(conn *net.UDPConn, addr *net.UDPAddr) {
	log.Println("=== start pong")

	var ping Ping
	ping.Type = 6
	ping.Flags = 0
	ping.Token = RECV_TOKEN

	buf := new(bytes.Buffer)
	binary.Write(buf, endian, ping)
	n, err := conn.WriteTo(buf.Bytes(), addr)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	log.Printf("=== end pong %v", n)
}

var RECV_TOKEN uint64
var SEND_TOKEN uint64
var guestAddr *net.UDPAddr

func server() {
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 9000,
	}
	conn, err := net.ListenUDP("udp", udpAddr) // UDPConn

	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	buf := make([]byte, 1024)
	log.Println("Starting udp server...")

	/*
		go func() {
			for {
				time.Sleep(time.Millisecond * 500)
				if guestAddr == nil {
					continue
				}

				sendPing(guestAddr)
			}
		}()
	*/

	//endian := binary.BigEndian
	//endian := binary.LittleEndian
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}

		switch buf[0] {
		case ConnectionRequestType:
			sendAccept(conn, addr, buf[:n])
		case DataType:
			sendData(conn, addr, buf[:n])
		case PingType:
			sendPong(conn, addr)
		default:
			//log.Printf("Reciving data: %s from %s", string(buf[:n]), addr.String())
			log.Printf("default Recv data: from %s", addr.String())
			for _, i := range buf[:n] {
				log.Printf("data: %v", i)
			}

			//log.Printf("Sending data..")
			//conn.WriteTo([]byte("Pong"), addr)
			//log.Printf("Complete Sending data..")
		}
	}
}

func main() {
	server()
}
