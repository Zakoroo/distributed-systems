package mr

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"
)

const (
	StatusSuccess        = "SUCCESS"
	StatusTaskQueueEmpty = "TASK_QUEUE_EMPTY"
	StatusSubmitRejected = "SUBMIT_REJECTED"
)

type Coordinator struct {
	mu        sync.Mutex
	taskQueue chan *Task
	pending   map[string]*Task
	done      bool
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	if files == nil || len(files) == 0 || nReduce <= 0 {
		panic("invalid coordinator arguments")
	}

	c := Coordinator{
		taskQueue: make(chan *Task, len(files)*nReduce),
		pending:   make(map[string]*Task),
	}

	go c.server(files, nReduce)
	return &c
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) FetchTask(args *Args, reply *Reply) error {
	task := <-c.taskQueue

	if task == nil {
		// Write fail status
		reply.parts["status"] = StatusTaskQueueEmpty
		reply.parts["description"] = "task queue is empty"

		return errors.New("task queue is closed")
	}

	// Write success status
	reply.parts["status"] = StatusSuccess

	// Write task id
	reply.parts["task_id"] = task.id

	// Write objective
	reply.parts["objective"] = task.objective

	// Write the input params
	for k, v := range task.input {
		reply.parts[k] = v
	}

	return nil
}

func (c *Coordinator) SubmitTask(args *Args, reply *Reply) error {
	id := args.parts["id"]

	if id == "" {
		reply.parts["status"] = StatusSubmitRejected
		reply.parts["description"] = "id is empty"
		return errors.New("invalid task id")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	task := c.pending[id]
	if task == nil {
		reply.parts["status"] = StatusSubmitRejected
		reply.parts["description"] = "this task id does not exist or already submitted"
		return errors.New("invalid task id")
	}

	delete(c.pending, id)
	reply.parts["status"] = StatusSuccess
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server(files []string, nReduce int) {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	// Serve incomming connections in the background
	go http.Serve(l, nil)

	/////////////////////////////// Map phase    ///////////////////////////////
	for _, file := range files {
		id, err := GenerateId()
		if err != nil {
			log.Fatal("unable to generate id")
		}

		task := &Task{
			id:        id,
			objective: ObjectiveMap,
			input:     make(map[string]string),
		}

		task.input["nReduce"] = strconv.Itoa(nReduce)
		task.input["input_file"] = file

		c.pending[task.id] = task
	}

	/////////////////////////////// Sort phase   ///////////////////////////////

	/////////////////////////////// Reduce phase ///////////////////////////////

	c.mu.Lock()
	c.done = true
	c.mu.Unlock()
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.

	return ret
}

func GenerateId() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", errors.New("failed to create id")
	}

	return hex.EncodeToString(bytes), nil
}
