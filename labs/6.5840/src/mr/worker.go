package mr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// ByKey sorts intermediate records by key.
type ByKey []KeyValue

// Len returns the number of records to sort.
func (a ByKey) Len() int { return len(a) }

// Swap exchanges two records in place.
func (a ByKey) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less reports whether record i's key sorts before record j's key.
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// ihash returns a nonnegative FNV-1a hash of key. Mappers take it modulo
// the number of reducers so the same key always goes to the same partition.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// Worker polls the coordinator, runs assigned tasks, and submits each successful
// assignment after its files are closed. It returns on an exit reply or a fetch
// transport failure.
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {
	const pollInterval = 200 * time.Millisecond
	for {
		var reply FetchTaskReply
		if err := call("Coordinator.FetchTask", &FetchTaskArgs{}, &reply); err != nil {
			log.Printf("fetch task: %v", err)
			if _, serverErr := err.(rpc.ServerError); serverErr {
				time.Sleep(pollInterval)
				continue
			}
			return
		}
		task := reply.Task
		assignment := task.Assignment
		var err error
		switch assignment.Objective {
		case ObjectiveWait:
			time.Sleep(pollInterval)
			continue
		case ObjectiveExit:
			return
		case ObjectiveMap:
			err = runMap(task, mapf)
		case ObjectiveReduce:
			err = runReduce(task, reducef)
		}
		if err != nil {
			log.Printf("run %s task %d: %v", assignment.Objective, assignment.ID, err)
			continue
		}

		var submitted SubmitTaskReply
		if err := call("Coordinator.SubmitTask", &assignment, &submitted); err != nil {
			log.Printf("submit %s task %d: %v", assignment.Objective, assignment.ID, err)
		}
	}
}

// runMap reads the assigned input file, calls mapf, and writes one JSON file
// per reduce partition into the attempt directory, including empty partitions.
func runMap(task Task, mapf func(string, string) []KeyValue) error {
	content, err := ioutil.ReadFile(task.InputFile)
	if err != nil {
		return err
	}
	partitions := make([][]KeyValue, task.NReduce)
	for _, kv := range mapf(task.InputFile, string(content)) {
		partition := ihash(kv.Key) % task.NReduce
		partitions[partition] = append(partitions[partition], kv)
	}
	for partition, records := range partitions {
		name := partitionFile(task.Assignment.Dir, partition)
		if err := writePartition(name, records); err != nil {
			return err
		}
	}
	return nil
}

// writePartition creates or truncates name, writes one JSON object per record,
// and closes the file before reporting success.
func writePartition(name string, records []KeyValue) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	for _, kv := range records {
		if err := encoder.Encode(kv); err != nil {
			file.Close()
			return err
		}
	}
	return file.Close()
}

// runReduce reads the assigned partition files, groups records by key, and
// calls reducef once per key. It flushes and closes the staged output file
// in the attempt directory before reporting success.
func runReduce(task Task, reducef func(string, []string) string) error {
	var records []KeyValue
	for _, name := range task.IntermediateFiles {
		var err error
		records, err = readPartition(name, records)
		if err != nil {
			return err
		}
	}
	sort.Sort(ByKey(records))

	file, err := os.Create(reduceOutput(task.Assignment.Dir))
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	for i := 0; i < len(records); {
		key := records[i].Key
		var values []string
		for i < len(records) && records[i].Key == key {
			values = append(values, records[i].Value)
			i++
		}
		if _, err := fmt.Fprintf(writer, "%s %s\n", key, reducef(key, values)); err != nil {
			file.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

// readPartition opens name, appends its JSON records to records, and closes
// the file before returning. An empty file leaves records unchanged.
func readPartition(name string, records []KeyValue) ([]KeyValue, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	for {
		var kv KeyValue
		err := decoder.Decode(&kv)
		if err == io.EOF {
			return records, nil
		}
		if err != nil {
			return nil, err
		}
		records = append(records, kv)
	}
}

// call opens a Unix socket connection for one coordinator RPC, fills reply
// on success, and closes the connection before returning.
func call(method string, args, reply interface{}) error {
	client, err := rpc.DialHTTP("unix", coordinatorSock())
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Call(method, args, reply)
}
