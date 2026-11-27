package mr

import (
	"os"
	"strconv"
)

type FetchTaskArgs struct{}

type FetchTaskReply struct {
	Task Task
}

type SubmitTaskReply struct {
	Accepted bool
}

// coordinatorSock returns the shared RPC socket path for the current user,
// /var/tmp/5840-mr-<uid>. The coordinator listens there and workers dial it.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
