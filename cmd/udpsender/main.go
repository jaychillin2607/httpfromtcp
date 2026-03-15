package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	address := "localhost:42069"

	udpAddress, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatal(err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		str, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatal(err)
		}
		_, err = udpConn.Write(str)
		if err != nil {
			log.Fatal(err)
		}

	}
}
