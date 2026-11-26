package mr

const (
	ObjectiveMap    = "map"
	ObjectiveSort   = "sort"
	ObjectiveReduce = "reduce"
	ObjectiveExit   = "exit"
)

type Task struct {
	id        string
	objective string
	input     map[string]string
}
