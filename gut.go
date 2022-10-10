package gut

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
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

func clientSendPing(conn net.Conn, token uint64) {
	//log.Println("====== start: send ping")

	buf := new(bytes.Buffer)

	var ping Ping
	ping.Type = 5
	ping.Flags = 0
	ping.Token = token

	buf.Reset()
	binary.Write(buf, endian, ping)
	_, err := conn.Write(buf.Bytes())
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	//log.Printf("SEND ping: %v", n)
}

func clientSendPong(conn net.Conn, token uint64) {
	//log.Println("====== start: send pong")
	buf := new(bytes.Buffer)

	var pong Pong
	pong.Type = 6
	pong.Flags = 0
	pong.Token = token

	buf.Reset()
	binary.Write(buf, endian, pong)
	_, err := conn.Write(buf.Bytes())
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	//log.Printf("SEND pong: %v", n)
}

func clientSendData(conn net.Conn, token uint64) {
	var n int
	var err error
	var data UdpData
	data.Type = 4
	data.Flags = 0
	data.Token = token
	data.Num = 10

	buf := new(bytes.Buffer)

	buf.Reset()
	err = binary.Write(buf, endian, data)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	log.Printf("SEND data: %v", buf.Bytes())
	n, err = conn.Write(buf.Bytes())
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	log.Printf("END data: %v", n)
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

func Client() {
	conn, err := net.Dial("udp", "127.0.0.1:9000")
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	defer conn.Close()

	var request ConnectRequest
	request.Type = 0
	request.Flags = 0
	request.Token = rand.Uint64()

	buf := new(bytes.Buffer)
	err = binary.Write(buf, endian, request)

	fmt.Println(buf.Bytes())
	n, err := conn.Write(buf.Bytes())
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	bs := make([]byte, 256)

	n, err = conn.Read(bs)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	//log.Printf("Received data: %s", string(recvBuf[:n]))
	log.Printf("Recv data: %v", bs[:n])

	var accept ConnectAccept
	buf = bytes.NewBuffer(bs[:n])
	binary.Read(buf, endian, &accept)

	//pp.Println(accept)

	token := accept.NewToken
	//token := accept.Token
	log.Printf("token: %v", token)

	// Ping
	//go func(conn net.Conn, token uint64) {
	//	for {
	//		sendPing(conn, token)
	//		// 500 ms 何もなければpingみたいな実装らしい
	//		// 条件がもうちょっと複雑だったのでもう一度しっかり見た方がいい
	//		time.Sleep(time.Millisecond * 450)
	//	}
	//}(conn, token)

	// Data
	go func(conn net.Conn, token uint64) {
		for {
			time.Sleep(time.Millisecond * 500)
			clientSendData(conn, token)
		}
	}(conn, token)

	for {
		bs := recvAny(conn)
		switch bs[0] {
		case 4: // Data
			{
				var data UdpData
				buf := bytes.NewBuffer(bs)
				log.Printf("reading: %v", bs)
				binary.Read(buf, endian, &data)
				log.Println("data start")
				//pp.Println(data)
				log.Println("data end")
			}
		case 5: // Ping
			clientSendPong(conn, token)
		case 6: // Pong
			log.Printf("PONG: %v", bs)
		default:
			log.Printf("RECV: %v", bs)
		}
	}

	// Recv Data
	/*
		var recv UdpData
		n, err = conn.Read(bs)
		log.Printf("Recd data: %v", bs[:n])
		buf = bytes.NewBuffer(bs[:n])
		binary.Read(buf, endian, &recv)

		pp.Println(recv)
	*/

}

/*
type Header struct {
	Type           byte
	Flags          byte
	SessionIdToken uint8
}

type ConnectRequest struct {
	UnityTransportHeader
}

type ConnectRequest struct {
	UnityTransportHeader
	NewSessionIdToken uint8
}
*/

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

func Server() {
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
