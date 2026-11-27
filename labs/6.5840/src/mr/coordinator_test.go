package mr

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestReduceRetryRejectsStaleSubmission verifies that only the current attempt
// can publish reduce output. It checks cleanup both during and after the job.
func TestReduceRetryRejectsStaleSubmission(t *testing.T) {
	root := t.TempDir()
	c := &Coordinator{workDir: filepath.Join(root, "job"), reduces: make([]taskState, 2)}
	if err := os.Mkdir(c.workDir, 0700); err != nil {
		t.Fatal(err)
	}

	original := fetchForTest(t, c).Assignment
	// Expire the assignment without waiting ten seconds in the test.
	c.reduces[0].deadline = time.Now().Add(-time.Second)
	retry := fetchForTest(t, c).Assignment
	if retry.Objective != ObjectiveReduce || retry.ID != original.ID || retry.Dir == original.Dir {
		t.Fatalf("expected a new attempt for the same reduce task, got %+v then %+v", original, retry)
	}
	if err := ioutil.WriteFile(reduceOutput(original.Dir), []byte("stale"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(reduceOutput(retry.Dir), []byte("current"), 0600); err != nil {
		t.Fatal(err)
	}

	var rejected, accepted SubmitTaskReply
	if err := c.SubmitTask(&original, &rejected); err != nil || rejected.Accepted || c.Done() {
		t.Fatalf("expected stale submission to leave the job active, got accepted=%v err=%v done=%v", rejected.Accepted, err, c.Done())
	}
	outputPath := filepath.Join(root, "mr-out-0")
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("stale submission published output, stat error = %v", err)
	}
	if err := c.SubmitTask(&retry, &accepted); err != nil || !accepted.Accepted || c.Done() {
		t.Fatalf("expected accepted output with the job still active, got accepted=%v err=%v done=%v", accepted.Accepted, err, c.Done())
	}
	output, err := ioutil.ReadFile(outputPath)
	if err != nil || string(output) != "current" {
		t.Fatalf("expected current output, got %q, err = %v", output, err)
	}
	if _, err := os.Stat(retry.Dir); !os.IsNotExist(err) {
		t.Fatalf("published attempt directory should be gone before job completion, stat error = %v", err)
	}
	if _, err := os.Stat(original.Dir); err != nil {
		t.Fatalf("stale attempt directory should still exist before job completion: %v", err)
	}

	last := fetchForTest(t, c).Assignment
	if err := ioutil.WriteFile(reduceOutput(last.Dir), nil, 0600); err != nil {
		t.Fatal(err)
	}
	var finished SubmitTaskReply
	if err := c.SubmitTask(&last, &finished); err != nil || !finished.Accepted || !c.Done() {
		t.Fatalf("expected final submission to finish the job, got accepted=%v err=%v done=%v", finished.Accepted, err, c.Done())
	}
	if _, err := os.Stat(c.workDir); !os.IsNotExist(err) {
		t.Fatalf("job staging directory should be gone, stat error = %v", err)
	}
}

// TestReduceWaitsForMapPublication verifies that a failed map publish keeps
// reducers waiting until a retry publishes the partition they need.
func TestReduceWaitsForMapPublication(t *testing.T) {
	c := &Coordinator{
		workDir: t.TempDir(),
		files:   []string{"input.txt"},
		maps:    make([]taskState, 1),
		reduces: make([]taskState, 1),
	}
	assignment := fetchForTest(t, c).Assignment
	if err := os.RemoveAll(assignment.Dir); err != nil {
		t.Fatal(err)
	}
	var failed SubmitTaskReply
	if err := c.SubmitTask(&assignment, &failed); err == nil || failed.Accepted {
		t.Fatalf("expected missing map output to fail publication, got accepted=%v err=%v", failed.Accepted, err)
	}
	if task := fetchForTest(t, c); task.Assignment.Objective != ObjectiveWait {
		t.Fatalf("reduce should wait after failed publication, got %+v", task)
	}

	c.maps[0].deadline = time.Now().Add(-time.Second)
	retry := fetchForTest(t, c).Assignment
	if err := ioutil.WriteFile(partitionFile(retry.Dir, 0), []byte("map output"), 0600); err != nil {
		t.Fatal(err)
	}
	var accepted SubmitTaskReply
	if err := c.SubmitTask(&retry, &accepted); err != nil || !accepted.Accepted {
		t.Fatalf("expected map publication to succeed, got accepted=%v err=%v", accepted.Accepted, err)
	}
	task := fetchForTest(t, c)
	if task.Assignment.Objective != ObjectiveReduce || len(task.IntermediateFiles) != 1 {
		t.Fatalf("expected a reduce task with one published map partition, got %+v", task)
	}
	contents, err := ioutil.ReadFile(task.IntermediateFiles[0])
	if err != nil || string(contents) != "map output" {
		t.Fatalf("expected published map output, got %q, err = %v", contents, err)
	}
}

// fetchForTest returns the next coordinator reply and fails the test on an RPC handler error.
func fetchForTest(t *testing.T, c *Coordinator) Task {
	t.Helper()
	var reply FetchTaskReply
	if err := c.FetchTask(&FetchTaskArgs{}, &reply); err != nil {
		t.Fatal(err)
	}
	return reply.Task
}
