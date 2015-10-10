package main

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"
	"time"
)

var (
	gcChan      = make(chan *gctrace, 1)
	scvgChan    = make(chan *scvgtrace, 1)
	noMatchChan = make(chan string, 1)
)

func runParserWith(line string, re *regexp.Regexp) {
	reader := bytes.NewReader([]byte(line))

	parser := Parser{
		reader:      reader,
		GcChan:      gcChan,
		ScvgChan:    scvgChan,
		NoMatchChan: noMatchChan,
	}

	parser.gcRegexp = re

	parser.Run()
}

func TestParserWithMatchingInputGo15(t *testing.T) {
	line := "gc 47 @1.101s 13%: 0.027+6.1+0.001+0.29+1.0 ms clock, 0.11+6.1+0+6.0/0.015/0.021+4.3 ms cpu, 6->7->5 MB, 7 MB goal, 4 P"

	go runParserWith(line, gcrego15)

	expectedGCTrace := &gctrace{
		Heap1: 7,
	}

	select {
	case gctrace := <-gcChan:
		if !reflect.DeepEqual(gctrace, expectedGCTrace) {
			t.Errorf("Expected gctrace to equal %+v. Got %+v instead.", expectedGCTrace, gctrace)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserWithMatchingInputGo14(t *testing.T) {
	line := "gc76(1): 2+1+1390+1 us, 1 -> 3 MB, 16397 (1015746-999349) objects, 1436/1/0 sweeps, 0(0) handoff, 0(0) steal, 0/0/0 yields\n"

	go runParserWith(line, gcrego14)

	expectedGCTrace := &gctrace{
		Heap1: 3,
	}

	select {
	case gctrace := <-gcChan:
		if !reflect.DeepEqual(gctrace, expectedGCTrace) {
			t.Errorf("Expected gctrace to equal %+v. Got %+v instead.", expectedGCTrace, gctrace)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserGoRoutinesInputGo14(t *testing.T) {
	line := "gc76(1): 2+1+1390+1 us, 1 -> 3 MB, 16397 (1015746-999349) objects, 12 goroutines, 1436/1/0 sweeps, 0(0) handoff, 0(0) steal, 0/0/0 yields\n"

	go runParserWith(line, gcrego14)

	expectedGCTrace := &gctrace{
		Heap1: 3,
	}

	select {
	case gctrace := <-gcChan:
		if !reflect.DeepEqual(gctrace, expectedGCTrace) {
			t.Errorf("Expected gctrace to equal %+v. Got %+v instead.", expectedGCTrace, gctrace)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserWithScvgLine(t *testing.T) {
	line := "scvg1: inuse: 12, idle: 13, sys: 14, released: 15, consumed: 16 (MB)"

	go runParserWith(line, nil)

	expectedScvgTrace := &scvgtrace{
		inuse:    12,
		idle:     13,
		sys:      14,
		released: 15,
		consumed: 16,
	}

	select {
	case scvgTrace := <-scvgChan:
		if !reflect.DeepEqual(scvgTrace, expectedScvgTrace) {
			t.Errorf("Expected scvgTrace to equal %+v. Got %+v instead.", expectedScvgTrace, scvgTrace)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}

func TestParserNonMatchingInput(t *testing.T) {
	line := "INFO: test"

	go runParserWith(line, nil)

	select {
	case <-gcChan:
		t.Fatalf("Unexpected trace result. This input should not trigger gcChan.")
	case <-scvgChan:
		t.Fatalf("Unexpected trace result. This input should not trigger scvgChan.")
	case <-noMatchChan:
		return
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Execution timed out.")
	}
}
