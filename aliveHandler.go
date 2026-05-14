package main

import (
	"encoding/json"
	"fmt"
	common "github.com/aadit-n3rdy/rainstorm_common"
	"net"
	"time"
)

func aliveHandler() {
	for true {
		time.Sleep(10 * time.Second)
		trackerList := GetTrackerIPs()
		for tr := range trackerList {
			destAddr := fmt.Sprintf("%v:%v", trackerList[tr], common.TRACKER_UDP_PORT)
			conn, err := net.Dial("udp", destAddr)
			if err != nil {
				appLogger.Warn().Err(err).Str("tracker", destAddr).Msg("could not open UDP connection to tracker")
				continue
			}
			defer conn.Close()
			buf, err := json.Marshal(GetTrackerFiles(trackerList[tr]))
			if err != nil {
				appLogger.Warn().Err(err).Str("tracker", destAddr).Msg("could not marshal alive file list")
				continue
			}
			_, err = conn.Write(buf)
			if err != nil {
				appLogger.Warn().Err(err).Str("tracker", destAddr).Msg("could not send UDP alive message")
				continue
			}
			//fmt.Printf("Sent alive %v to %v\n", string(buf), destAddr)
		}
	}
}
