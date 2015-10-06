package main

import (
	"log"
	"os"
	"os/exec"
)

type SubCommand struct {
	cmd      *exec.Cmd
	PipeRead *os.File
}

type Stream int

const (
	StreamStdout Stream = 1 << iota
	StreamStderr
)

func NewSubCommand(args []string, stream Stream) SubCommand {
	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}

	env := append(os.Environ(), "GODEBUG=gctrace=1")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin

	switch stream {
	case StreamStdout:
		cmd.Stdout = pipeWrite
		cmd.Stderr = os.Stderr
	case StreamStderr:
		cmd.Stdout = os.Stdout
		cmd.Stderr = pipeWrite
	}

	return SubCommand{
		cmd:      cmd,
		PipeRead: pipeRead,
	}
}

func (s *SubCommand) Run() {
	if err := s.cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
