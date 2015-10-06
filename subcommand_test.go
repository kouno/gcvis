package main

import (
	"bufio"
	"testing"
	"time"
)

func TestSubCommandPipeRead(t *testing.T) {
	cmd := []string{"echo", "hello world"}
	subcommand := NewSubCommand(cmd, StreamStdout)
	done := make(chan bool, 1)

	go func() {
		subcommand.Run()

		info, err := subcommand.PipeRead.Stat()
		if err != nil {
			t.Errorf("Could not get stat from pipe: %v", err)
			done <- true
		}

		if info.Size() <= int64(0) {
			t.Fatalf("Expected command to write to pipe: %v byte written", info.Size())
		}

		buffer := make([]byte, info.Size())
		reader := bufio.NewReader(subcommand.PipeRead)
		_, err = reader.Read(buffer)
		if err != nil {
			t.Fatalf("Unexpected error while reading from Stdout: %v", err)
			done <- true
		}

		if string(buffer) != "hello world\n" {
			t.Errorf("line is not equal to 'hello world': '%v'", string(buffer))
			done <- true
		}

		done <- true
	}()

	select {
	case <-done:
		return
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}
