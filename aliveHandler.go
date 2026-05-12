package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	common "github.com/aadit-n3rdy/rainstorm_common"
)

func aliveHandler() {
	for true {
		time.Sleep(10 * time.Second)
		trackerList := GetTrackerIPs()
		for tr := range trackerList {
			destAddr := fmt.Sprintf("%v:%v", trackerList[tr], common.TRACKER_UDP_PORT)
			conn, err := net.Dial("udp", destAddr)
			if err != nil {
				appLogger.Error().Err(err).Str("tracker_addr", destAddr).Msg("could not open UDP connection with tracker")
				continue
			}
			defer conn.Close()
			buf, err := json.Marshal(GetTrackerFiles(trackerList[tr]))
			if err != nil {
				appLogger.Error().Err(err).Str("tracker_addr", destAddr).Msg("could not marshal alive file list")
				continue
			}
			_, err = conn.Write(buf)
			if err != nil {
				appLogger.Error().Err(err).Str("tracker_addr", destAddr).Msg("could not send UDP alive message")
				continue
			}
		}
	}
}
