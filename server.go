package main

import (
	"fmt"
	"log"
	"net"

	"github.com/ChengYaoYan/distributedfilesystemgo/p2p"
)

type FileServerOpts struct {
	ListenAddr        net.Addr
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	TCPTransport      p2p.TCPTransport
}

type FileServer struct {
	FileServerOpts        FileServerOpts
	Store                 *Store
	quitch                chan struct{}
	bootstrapNetworkNodes []string
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{PathTransformFunc: opts.PathTransformFunc}

	return &FileServer{FileServerOpts: opts, Store: NewStore(storeOpts)}
}

func (fs *FileServer) Stop() {
	close(fs.quitch)
}

func (fs *FileServer) Loop() {
	defer func() {
		log.Printf("file server stopped due to user quit action")
		fs.FileServerOpts.TCPTransport.Close()
	}()

	for {
		select {
		case msg := <-fs.FileServerOpts.TCPTransport.Consume():
			fmt.Println(msg)
		case <-fs.quitch:
			return
		}
	}
}

func (fs *FileServer) BootstrapNetwork() error {
	for _, addr := range fs.bootstrapNetworkNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Println("attempting to connect with remote: ", addr)
			if err := fs.FileServerOpts.TCPTransport.Dial(addr); err != nil {
				log.Println("dial err: ", err)
			}

		}(addr)
	}

	return nil
}

func (fs *FileServer) Start() error {
	err := fs.FileServerOpts.TCPTransport.ListenAndAccept()

	if err != nil {
		return err
	}

	fs.Loop()

	return nil
}
