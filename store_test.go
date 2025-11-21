package main

import (
	"bytes"
	"io"
	"testing"
)

func TestPathTransfromFunc(t *testing.T) {
	key := "password"
	pathKey := GASPathTransformFunc(key)
	expectedPathName := "5baa6/1e4c9/b93f3/f0682/250b6/cf833/1b7ee/68fd8/"
	expectedFileName := "5baa61e4c9b93f3f0682250b6cf8331b7ee68fd8"

	if pathKey.PathName != expectedPathName {
		t.Errorf("want: %s, have: %s", expectedPathName, pathKey.PathName)
	}
	if pathKey.FileName != expectedFileName {
		t.Errorf("want: %s, have: %s", expectedFileName, pathKey.FileName)
	}
}

func TestStore(t *testing.T) {
	key := "password"
	data := []byte("some data blablalba")
	opts := StoreOpts{PathTransformFunc: GASPathTransformFunc}
	store := NewStore(opts)

	// test writing
	if err := store.WriteStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	// test reading
	r, err := store.Read(key)
	if err != nil {
		t.Error(err)
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		t.Error(t)
	}
	if string(buf) != string(data[:]) {
		t.Errorf("want %s, have %s", data, buf)
	}

	// test deleting
	store.Delete(key)
}
