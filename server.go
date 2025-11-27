package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/ChengYaoYan/distributedfilesystemgo/p2p"
)

type FileServerOpts struct {
	ListenAddr            string
	PathTransformFunc     PathTransformFunc
	Transport             p2p.TCPTransport
	bootstrapNetworkNodes []string
}

type FileServer struct {
	FileServerOpts FileServerOpts

	peerLock sync.Mutex
	peers    map[net.Addr]p2p.Peer

	Store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{PathTransformFunc: opts.PathTransformFunc}

	return &FileServer{
		FileServerOpts: opts,
		Store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[net.Addr]p2p.Peer),
	}
}

type Payload struct {
	key  string
	data []byte
}

func (fs *FileServer) broadcast(p Payload) error {
	peers := []io.Writer{}
	for _, peer := range fs.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(p)
}

func (fs *FileServer) StoreData(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)

	if err := fs.Store.Write(key, tee); err != nil {
		return err
	}

	p := Payload{
		key:  key,
		data: buf.Bytes(),
	}

	fmt.Println(fs.FileServerOpts.Transport.TCPTransportOpts.ListenAddr, buf.Bytes())

	return fs.broadcast(p)
}

func (fs *FileServer) Stop() {
	close(fs.quitch)
}

func (fs *FileServer) OnPeer(p p2p.Peer) error {
	fs.peerLock.Lock()
	defer fs.peerLock.Unlock()

	fs.peers[p.RemoteAddr()] = p

	log.Println("connect with remote: ", p.RemoteAddr())

	return nil
}

func (fs *FileServer) Loop() {
	defer func() {
		log.Printf("file server stopped due to user quit action")
		fs.FileServerOpts.Transport.Close()
	}()

	for {
		select {
		case rpc := <-fs.FileServerOpts.Transport.Consume():
			var p Payload
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&p); err != nil {
				if peer, ok := fs.peers[rpc.From]; !ok {
					panic("peer not fuond in peers")
				} else {
					fmt.Println("broadcast", peer, p.data)
				}

				log.Fatal(err)
			}
		case <-fs.quitch:
			return
		}
	}
}

func (fs *FileServer) BootstrapNetwork() error {
	for _, addr := range fs.FileServerOpts.bootstrapNetworkNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Println("attempting to connect with remote: ", addr)
			if err := fs.FileServerOpts.Transport.Dial(addr); err != nil {
				log.Println("dial err: ", err)
			}
		}(addr)
	}

	return nil
}

func (fs *FileServer) Start() error {
	if err := fs.FileServerOpts.Transport.ListenAndAccept(); err != nil {
		return err
	}

	if err := fs.BootstrapNetwork(); err != nil {
		return err
	}

	fs.Loop()

	return nil
}
