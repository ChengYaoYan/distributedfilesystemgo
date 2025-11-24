package main

import (
	"fmt"
	"log"
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
	peers    map[string]p2p.Peer

	Store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{PathTransformFunc: opts.PathTransformFunc}

	return &FileServer{
		FileServerOpts: opts,
		Store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (fs *FileServer) Stop() {
	close(fs.quitch)
}

func (fs *FileServer) OnPeer(p p2p.Peer) error {
	fs.peerLock.Lock()
	defer fs.peerLock.Unlock()

	fs.peers[p.RemoteAddr().String()] = p

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
		case msg := <-fs.FileServerOpts.Transport.Consume():
			fmt.Println(msg)
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
