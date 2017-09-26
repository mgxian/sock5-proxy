package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
)

var (
	// AuthenticationResponse 初始响应的认证数据
	AuthenticationResponse = []byte{5, 0}
)

func handleHandShake(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	reader.ReadByte()
	methodLen, _ := reader.ReadByte()
	method := make([]byte, methodLen)
	reader.Read(method)

	if n, err := conn.Write(AuthenticationResponse); err != nil {
		fmt.Println(err, n)
		return err
	}

	return nil
}

func getAddress(requestMata []byte) (string, string) {
	n := len(requestMata)
	remoteTye := requestMata[3]
	remoteAddr := requestMata[4:]
	remotePort := requestMata[n-2 : n]
	var host, port string
	switch remoteTye {
	case 0x01: // ipv4
		ipv4 := remoteAddr[:4]
		host = net.IPv4(ipv4[0], ipv4[1], ipv4[2], ipv4[3]).String()
	case 0x03: // domain
		domainLength := int(remoteAddr[0])
		host = string(remoteAddr[1 : domainLength+1])
	case 0x04: // ipv6
		ipv6 := remoteAddr[:16]
		host = net.IP{ipv6[0], ipv6[1], ipv6[2], ipv6[3], ipv6[4], ipv6[5], ipv6[6], ipv6[7], ipv6[8], ipv6[9], ipv6[10], ipv6[11], ipv6[12], ipv6[13], ipv6[14], ipv6[15]}.String()
	}

	port = strconv.Itoa(int(remotePort[0])<<8 | int(remotePort[1]))

	return host, port
}

func handle(conn net.Conn) {
	fmt.Println("sock5-proxy: got a client from ", conn.RemoteAddr().String())
	defer conn.Close()

	reader := bufio.NewReader(conn)

	if err := handleHandShake(conn); err != nil {
		return
	}

	requestMata := make([]byte, 1024)
	n, err := reader.Read(requestMata)
	if err != nil {
		fmt.Println(err)
		return
	}

	host, port := getAddress(requestMata[:n])
	fmt.Println(host, port)

	server, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer server.Close()

	conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})

	go io.Copy(server, conn)
	io.Copy(conn, server)
}

// Parameter 存储命令行参数的struct
type Parameter struct {
	host string
	port string
}

// NewParameter 创建存储命令行参数的struct
func NewParameter(host, port string) (*Parameter, error) {
	return &Parameter{
		host: host,
		port: port,
	}, nil
}

func handleParameter() (*Parameter, error) {
	host := flag.String("h", "", "-h listen address defualt 0.0.0.0")
	port := flag.String("p", "1025", "-p listen port default 1025")

	flag.Parse()

	return NewParameter(*host, *port)
}

func getListenAddress() string {
	p, _ := handleParameter()
	return p.host + ":" + p.port
}

func main() {
	listenAddress := getListenAddress()
	fmt.Println(listenAddress)
	listen, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handle(conn)
	}

}
