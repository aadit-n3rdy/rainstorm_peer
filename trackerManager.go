package main

import (
	"sync"
)

var trackerManager map[string]map[string]struct{} // Map of tracker IP to set of file IDs
var trackerMutex sync.Mutex

func TrackerManagerInit() {
	trackerManager = make(map[string]map[string]struct{})
}

func AddTracker (sf *StoredFile) {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	set, ok := trackerManager[sf.TrackerIP]
	if !ok {
		set = make(map[string]struct{})
		trackerManager[sf.TrackerIP] = set
	}
	_, ok = set[sf.FileID]
	if !ok {
		set[sf.FileID] = struct{}{}
	}
}

func RemoveTracker (sf *StoredFile) {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	set, ok := trackerManager[sf.TrackerIP]
	if !ok {
		return
	}
	delete(set, sf.FileID)
}

func GetTrackerIPs () []string {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	res := make([]string, len(trackerManager))
	i := 0
	for k := range(trackerManager) {
		res[i] = k
		i += 1
	}
	return res
}

func GetTrackerFiles (trackerIP string) []string {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	set, ok := trackerManager[trackerIP]
	if !ok {
		return make([]string, 0)
	}
	res := make([]string, len(set))
	i := 0
	for k := range(set) {
		res[i] = k
		i += 1
	}
	return res
}
