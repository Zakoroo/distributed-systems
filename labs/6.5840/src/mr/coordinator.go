package mr

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"sync"
	"time"

	"6.5840/observability"
)

const taskTimeout = 10 * time.Second

type taskState struct {
	attemptDir string
	deadline   time.Time
	done       bool
}

type Coordinator struct {
	mu       sync.Mutex
	files    []string
	workDir  string
	maps     []taskState
	reduces  []taskState
	nReduce  int
	reporter *observability.Reporter
}

// MakeCoordinator creates one map task per input file and nReduce reduce tasks.
// It creates a private .mr-job-* directory in the working directory and starts
// the RPC server. It panics if nReduce is zero or negative.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	if nReduce <= 0 {
		panic("nReduce must be positive")
	}

	// Private to this job and on the same filesystem as the final outputs.
	workDir, err := ioutil.TempDir(".", ".mr-job-")
	if err != nil {
		log.Fatalf("create job directory: %v", err)
	}
	stderr, statErr := os.Stderr.Stat()
	interactive := statErr == nil && stderr.Mode()&os.ModeCharDevice != 0 && os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb"
	c := &Coordinator{
		files:    append([]string(nil), files...),
		workDir:  workDir,
		maps:     make([]taskState, len(files)),
		reduces:  make([]taskState, nReduce),
		reporter: observability.New(os.Stderr, interactive),
		nReduce:  nReduce,
	}
	c.server()
	c.reporter.Logf("MapReduce started with %d map tasks and %d reduce tasks", len(c.maps), len(c.reduces))
	return c
}

// allDone reports whether every task has published its output.
// An empty task list is complete.
func allDone(tasks []taskState) bool {
	for _, task := range tasks {
		if !task.done {
			return false
		}
	}
	return true
}

// activeTasks returns the map tasks while any map is unfinished, then the
// reduce tasks. Once all tasks are done, it returns ObjectiveExit and no tasks.
// The caller must hold c.mu.
func (c *Coordinator) activeTasks() (Objective, []taskState) {
	if !allDone(c.maps) {
		return ObjectiveMap, c.maps
	}
	if !allDone(c.reduces) {
		return ObjectiveReduce, c.reduces
	}
	return ObjectiveExit, nil
}

// FetchTask assigns an available task in the current phase. It creates a fresh
// staging directory and sets a ten-second deadline for reassignment.
// Reduce assignments list their partition files from the published map directories.
// The reply tells the worker to wait if no task is available or exit if the job is done.
func (c *Coordinator) FetchTask(_ *FetchTaskArgs, reply *FetchTaskReply) error {
	var event string
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
		c.reportEvent(event)
	}()

	objective, tasks := c.activeTasks()
	if objective == ObjectiveExit {
		reply.Task.Assignment.Objective = ObjectiveExit
		return nil
	}

	now := time.Now()
	for id := range tasks {
		state := &tasks[id]
		if state.done || now.Before(state.deadline) {
			continue
		}
		dir, err := ioutil.TempDir(c.workDir, "attempt-")
		if err != nil {
			err := fmt.Errorf("create staging directory for %s task %d: %w", objective, id, err)
			event = err.Error()
			return err
		}
		if state.attemptDir != "" {
			event = fmt.Sprintf("reassigned %s task %d after its deadline", objective, id)
		}
		state.attemptDir = dir
		state.deadline = time.Now().Add(taskTimeout)

		task := Task{Assignment: Assignment{Objective: objective, ID: id, Dir: dir}}
		if objective == ObjectiveMap {
			task.InputFile = c.files[id]
			task.NReduce = c.nReduce
		} else {
			for mapID := range c.maps {
				task.IntermediateFiles = append(task.IntermediateFiles, partitionFile(c.mapOutput(mapID), id))
			}
		}
		reply.Task = task
		return nil
	}

	reply.Task.Assignment.Objective = ObjectiveWait
	return nil
}

// SubmitTask publishes output only for the current assignment. A stale or
// duplicate submission leaves Accepted false.
// For maps, it renames the staging directory to map-N inside the job directory.
// For reduces, it renames the staged output file to mr-out-N beside the job directory.
// A successful reduce removes its attempt directory. The last reduce removes the
// entire job directory. A failed rename leaves the task unfinished.
func (c *Coordinator) SubmitTask(args *Assignment, reply *SubmitTaskReply) error {
	var event string
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
		c.reportEvent(event)
	}()

	objective, tasks := c.activeTasks()
	if args.Objective != objective || args.ID < 0 || args.ID >= len(tasks) || args.Dir == "" {
		return nil
	}
	state := &tasks[args.ID]
	if state.done || state.attemptDir != args.Dir {
		return nil
	}

	// Keep the attempt check, rename, and completion under the same lock.
	staged, published := state.attemptDir, c.mapOutput(args.ID)
	if objective == ObjectiveReduce {
		staged = reduceOutput(staged)
		published = filepath.Join(filepath.Dir(c.workDir), fmt.Sprintf("mr-out-%d", args.ID))
	}
	if err := os.Rename(staged, published); err != nil {
		err := fmt.Errorf("publish %s task %d: %w", objective, args.ID, err)
		event = err.Error()
		return err
	}
	state.done = true
	reply.Accepted = true

	if objective == ObjectiveReduce {
		cleanupDir := state.attemptDir
		if allDone(c.reduces) {
			cleanupDir = c.workDir
		}
		if err := os.RemoveAll(cleanupDir); err != nil {
			event = fmt.Sprintf("remove staging directory %q: %v", cleanupDir, err)
		}
	}
	return nil
}

// reportEvent logs a scheduling or file error outside the coordinator lock.
func (c *Coordinator) reportEvent(event string) {
	if c.reporter != nil && event != "" {
		c.reporter.Logf("%s", event)
	}
}

// mapOutput returns the path to the published map directory, such as
// .mr-job-123/map-2. SubmitTask renames the attempt directory to this path,
// and FetchTask uses it to locate the map's partition files for reducers.
func (c *Coordinator) mapOutput(id int) string {
	return filepath.Join(c.workDir, fmt.Sprintf("map-%d", id))
}

// server registers the coordinator's RPC methods, replaces the Unix socket
// path, and starts serving worker requests in a goroutine.
func (c *Coordinator) server() {
	server := rpc.NewServer()
	if err := server.Register(c); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle(rpc.DefaultRPCPath, server)

	sockname := coordinatorSock()
	os.Remove(sockname)
	listener, err := net.Listen("unix", sockname)
	if err != nil {
		log.Fatal("listen error: ", err)
	}
	go http.Serve(listener, mux)
}

// Done reports whether every reduce output has been published and displays
// the current progress when called by the coordinator's polling loop.
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	maps, reduces := countDone(c.maps), countDone(c.reduces)
	mapTotal, reduceTotal := len(c.maps), len(c.reduces)
	c.mu.Unlock()

	if c.reporter != nil {
		label := fmt.Sprintf("map %d/%d | reduce %d/%d", maps, mapTotal, reduces, reduceTotal)
		c.reporter.Progress(label, maps+reduces, mapTotal+reduceTotal)
	}
	return maps == mapTotal && reduces == reduceTotal
}

// countDone counts published tasks. The caller must hold the coordinator lock.
func countDone(tasks []taskState) int {
	count := 0
	for _, task := range tasks {
		if task.done {
			count++
		}
	}
	return count
}
