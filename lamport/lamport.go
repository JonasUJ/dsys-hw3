package lamport

import (
	"github.com/JonasUJ/dsys-hw3/chittychat"
)

type Lamport interface {
	GetTime() uint64
	GetPid() uint32
}

// Create a new message with the time and pid of the passed lamport, and the passed content.
func MakeMessage(lamport Lamport, content string) *chittychat.Message {
	return &chittychat.Message{
		Time:    lamport.GetTime(),
		Pid:     uint32(lamport.GetPid()),
		Content: content,
	}
}

// Calculate new time on a send event.
func LamportSend(lamport Lamport) uint64 {
	return lamport.GetTime() + 1
}

// Calculate new time on a recv event, comparing the two lamports to determine which is greater.
func LamportRecv(lamport, other Lamport) uint64 {
	if Compare(lamport, other) > 0 {
		return lamport.GetTime() + 1
	} else {
		return other.GetTime() + 1
	}
}

// Compare two lamports (according to the spec ofc. ;)
func Compare(lamport, other Lamport) int {
	// First compare by time, then by pid
	if lamport.GetTime() < other.GetTime() ||
		lamport.GetTime() == other.GetTime() &&
			lamport.GetPid() < other.GetPid() {
		return -1
	} else if lamport.GetPid() == other.GetPid() {
		return 0
	} else {
		return 1
	}
}
