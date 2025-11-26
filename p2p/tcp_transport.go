package p2p

import (
	"fmt"
	"log"
	"net"
)

type TCPPeer struct {
	net.Conn
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
	}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{TCPTransportOpts: opts, rpcch: make(chan RPC)}
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)

	if err != nil {
		return err
	}

	log.Println("TCP transport listening on: ", t.ListenAddr)

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()

		if err != nil {
			log.Fatal(err)
		}

		log.Println("new incoming connection ", conn)

		go t.handleConn(conn, true)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn, outbood bool) error {
	var err error

	defer func() {
		log.Println("dropping peer connection: ", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbood)

	if err = t.HandshakeFunc(peer); err != nil {
		log.Printf("TCP handshake error: %s\n", err)
		return err
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return err
		}
	}

	rpc := RPC{}
	for {
		if err := t.Decoder.Decode(conn, &rpc); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}

		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
