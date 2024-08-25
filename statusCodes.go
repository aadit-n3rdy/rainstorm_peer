package main

import "fmt"

type StatusCode int;

const (
	STATUS_OK StatusCode = iota
	STATUS_MISSING
	STATUS_CHUNK_MISSING
	STATUS_DONE
);

func (self StatusCode) String() string {
	switch self {
	case STATUS_OK:
		return "STATUS_OK"
	case STATUS_MISSING:
		return "STATUS_MISSING"
	case STATUS_CHUNK_MISSING:
		return "STATUS_CHUNK_MISSING"
	case STATUS_DONE:
		return "STATUS_DONE"
	default:
		return fmt.Sprintf("STATUS_INVALID_%v", int(self))
	}
}
