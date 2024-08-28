package main

import (
	"fmt"
	"testing"
)

func TestReceiver(t *testing.T) {

	chunker := &Chunker{}
	chunker.init("./receiver_chunk_path")

	fmt.Println("Chunker ready");

	var trackerIP string

	fmt.Print("Enter tracker IP address: ")
	fmt.Scanf("%s", &trackerIP)

	//trackerIP := "127.0.0.1"
	fdd, err := fetchFDD("somefileid", trackerIP)
	if err != nil {
		fmt.Printf("Error fetching FDD: %v\n", err)
		return
	}

	fmt.Println(fdd)

	err = fileReceiver(
		fdd,
		"testfile.jpeg",
		chunker,
	)

	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Test done!")
}
