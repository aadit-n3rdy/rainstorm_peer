package main;

type PeerThreadType int;

const (
	PEER_THREAD_LISTENER PeerThreadType = iota;
	PEER_THREAD_SENDER = iota;
	PEER_THREAD_RECEIVER = iota;
)

type PeerChannelMsg int;

const (
	CHANNEL_EXIT PeerChannelMsg = iota; 	// Forces channel to exit
	CHANNEL_STATUS							// Request for channel status
)

type PeerThread struct {
	typ PeerThreadType;
	channel chan any;
};
