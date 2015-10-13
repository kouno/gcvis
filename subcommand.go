package main

import (
	"log"
	"os"
	"os/exec"
)

type SubCommand struct {
	cmd       *exec.Cmd
	PipeRead  *os.File
	pipeWrite *os.File
	Err       error
}

func NewSubCommand(args []string) SubCommand {
	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}

	env := append(os.Environ(), "GODEBUG=gctrace=1")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = pipeWrite

	return SubCommand{
		cmd:       cmd,
		PipeRead:  pipeRead,
		pipeWrite: pipeWrite,
	}
}

func (s *SubCommand) Run() {
	if err := s.cmd.Run(); err != nil {
		s.Err = err
	}
	s.pipeWrite.Close()
}
