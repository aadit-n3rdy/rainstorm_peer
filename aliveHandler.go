package main

import (
	"time"
	"encoding/json"
	"net"
	"fmt"
	"rainstorm/common"
)

func aliveHandler() {
	for true {
		time.Sleep(10*time.Second)
		trackerList := GetTrackerIPs()
		for tr := range(trackerList) {
			destAddr := fmt.Sprintf("%v:%v", trackerList[tr], common.TRACKER_UDP_PORT)
			conn, err := net.Dial("udp", destAddr)
			if err != nil {
				fmt.Printf("Couldn't open UDP with tracker %v: %v\n", destAddr, err)
				continue
			}
			defer conn.Close()
			buf, err := json.Marshal(GetTrackerFiles(trackerList[tr]))
			if err != nil {
				fmt.Printf("Couldn't marshal alive file list %v: %v\n", destAddr, err)
				continue
			}
			_, err = conn.Write(buf)
			if err != nil {
				fmt.Printf("Couldn't send UDP alive to %v: %v\n", destAddr, err)
				continue
			}
			fmt.Printf("Sent alive %v to %v\n", string(buf), destAddr)
		}
	}
}
