package main;

import (
	"testing"
	"fmt"
);

func TestFileChunk(t *testing.T) {
	var ck Chunker
	ck.init("chunkpath/")
	cf, err := ck.chunkFile("testfiles/file_100K.jpg")
	fmt.Println(cf, err)
	cf, err = ck.chunkFile("testfiles/file_100K.jpg")
	fmt.Println(cf, err)
}
