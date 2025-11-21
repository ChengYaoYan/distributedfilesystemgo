package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)

type PathTransformFunc func(string) PathKey

func GASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	var pathKey PathKey
	pathKey.FileName = hashStr

	blockSize := 5
	blockNum := len(hashStr) / 5

	for i := range blockNum {
		from, to := i*blockSize, (i+1)*blockSize
		pathKey.PathName += hashStr[from:to] + "/"
	}

	return pathKey
}

type PathKey struct {
	PathName string
	FileName string
}

func (pk *PathKey) FullName() string {
	return pk.PathName + pk.FileName
}

func (pk *PathKey) FirstName() string {
	return strings.Split(pk.PathName, "/")[0]
}

type StoreOpts struct {
	PathTransformFunc
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	return &Store{opts}
}

func (s *Store) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)
	_, err := os.Stat(pathKey.FullName())

	return os.IsExist(err)
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)

	defer func() {
		log.Printf("Deleted %s from disk", pathKey.FileName)
	}()

	return os.RemoveAll(pathKey.FirstName())
}

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.ReadStream(key)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Store) ReadStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	return os.Open(pathKey.FullName())
}

func (s *Store) Write(key string, r io.Reader) error {
	return s.WriteStream(key, r)
}

func (s *Store) WriteStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)
	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, r)

	pathAndFilename := pathKey.FullName()

	f, err := os.Create(pathAndFilename)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, buf)
	if err != nil {
		return err
	}

	log.Printf("written (%d) bytes to disk: %s", n, pathAndFilename)

	return nil
}
