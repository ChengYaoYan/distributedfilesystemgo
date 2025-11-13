package main

import (
	"fmt"
	"log"

	"github.com/ChengYaoYan/distributedfilesystemgo/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":4000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcp := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := <-tcp.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()

	if err := tcp.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
