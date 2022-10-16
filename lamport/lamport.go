package lamport

type Lamport struct {
	time int64
	pid  int32
}

func (lamport *Lamport) Send() {
	lamport.time += 1
}

func (lamport *Lamport) Recv(other *Lamport) {
	if lamport.Compare(*other) > 0 {
		lamport.time = other.time + 1
	} else {
		lamport.time += 1
	}
}

func (lamport Lamport) Compare(other Lamport) int {
	if lamport.time < other.time || lamport.time == other.time && lamport.pid < other.pid {
		return -1
	} else if lamport.pid == other.pid {
		return 0
	} else {
		return 1
	}
}
