package main;

import (
	"testing"
	"fmt"
);

func TestFileChunk(t *testing.T) {
	var ck Chunker
	ck.init("chunkpath/")
	cf, err := ck.addChunkedFile("testfiles/file_100K.jpg")
	fmt.Println(cf, err)
	uu, err := ck.getChunks(cf)
	fmt.Println(uu, err)
}
