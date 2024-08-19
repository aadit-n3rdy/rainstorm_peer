package main

import (
	"encoding/json"
	"fmt"
	"net"
	"rainstorm/common"
	"strings"
)

type PeerThreadType int;

const (
	PEER_THREAD_LISTENER PeerThreadType = iota;
	PEER_THREAD_SENDER = iota;
	PEER_THREAD_RECEIVER = iota;
)

type PeerThread struct {
	typ PeerThreadType;
};

func main() {
	done := false
	var s string;
	for true {
		fmt.Scanln(&s);
		tokens := strings.Fields(s)
		if len(tokens) == 0 {
			continue;
		}
	}
	return;
}
