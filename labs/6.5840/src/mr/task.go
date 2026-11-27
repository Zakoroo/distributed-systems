package mr

import (
	"fmt"
	"path/filepath"
)

type Objective string

const (
	ObjectiveMap    Objective = "map"
	ObjectiveReduce Objective = "reduce"
	ObjectiveWait   Objective = "wait"
	ObjectiveExit   Objective = "exit"
)

// Assignment identifies a task and the staging directory for this attempt.
// Workers return it unchanged when they submit their output.
type Assignment struct {
	Objective Objective
	ID        int
	Dir       string
}

type Task struct {
	Assignment        Assignment
	InputFile         string
	NReduce           int
	IntermediateFiles []string
}

// partitionFile returns the path dir/partition-N.json for partition N.
// Mappers pass their attempt directory to write it. The coordinator passes
// the published map directory to build a reducer's input paths.
func partitionFile(dir string, partition int) string {
	return filepath.Join(dir, fmt.Sprintf("partition-%d.json", partition))
}

// reduceOutput returns dir/output, the path where a reducer stages its result.
// SubmitTask uses this path as the source when publishing mr-out-N.
func reduceOutput(dir string) string {
	return filepath.Join(dir, "output")
}
