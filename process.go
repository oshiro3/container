package main

import (
	"os"
	"os/exec"
)

type parentProcess interface{}

type filePair struct {
	parent *os.File
	child  *os.File
}

type setnsProcess struct {
	cmd             *exec.Cmd
	config          *initConfig
	messageSockPair filePair
	logFilePair     filePair
	process         *Process
}

type initProcess struct {
	cmd             *exec.Cmd
	config          *initConfig
	messageSockPair filePair
	logFilePair     filePair
	container       *MyContainer
	fds             []string
	process         *Process
}

func (sp *setnsProcess) start() error {
}
