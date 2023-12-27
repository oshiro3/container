package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/utils"
)

type MyContainer struct {
	id          string
	root        string
	config      *configs.Config
	initPath    string
	initArgs    []string
	initProcess parentProcess
	fifo        *os.File
}

func NewContainer(id string, config *configs.Config, initArgs []string) (*MyContainer, error) {
	abs, _ := filepath.Abs(".")
	return &MyContainer{
		id:       id,
		root:     abs + "/dev/rootfs",
		config:   config,
		initArgs: initArgs,
	}

}

func (c *MyContainer) Run(process *libcontainer.Process) error {
	if err := c.start(process); err != nil {
		return err
	}
	if process.Init {
		return c.exec()
	}
	return nil
}

func (c *MyContainer) start(process *libcontainer.Process) error {
	// if process.Init {
	// 	if err := c.createExecFifo(); err != nil {
	// 		return err
	// 	}
	// }

	parent, err := c.newParentProcess(process)
	if err != nil {
		return fmt.Errorf("unable to create new parent process: %w", err)
	}

	if err := parent.start(); err != nil {
		return fmt.Errorf("unable to start container process: %w", err)
	}

	// if process.Init {
	// 	c.fifo.Close()
	// }
	return nil
}

func (c *linuxContainer) newParentProcess(p *Process) (parentProcess, error) {
	parentInitPipe, childInitPipe, err := utils.NewSockPair("init")
	if err != nil {
		return nil, fmt.Errorf("unable to create init pipe: %w", err)
	}
	messageSockPair := filePair{parentInitPipe, childInitPipe}

	parentLogPipe, childLogPipe, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("unable to create log pipe: %w", err)
	}
	logFilePair := filePair{parentLogPipe, childLogPipe}

	cmd := c.commandTemplate(p, childInitPipe, childLogPipe)
	if !p.Init {
		return c.newSetnsProcess(p, cmd, messageSockPair, logFilePair)
	}

	return c.newInitProcess(p, cmd, messageSockPair, logFilePair)
}

func (c *linuxContainer) newSetnsProcess(p *Process, cmd *exec.Cmd, messageSockPair, logFilePair filePair) (*setnsProcess, error) {
	cmd.Env = append(cmd.Env, "_LIBCONTAINER_INITTYPE="+string(initSetns))
	// state, err := c.currentState()
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to get container state: %w", err)
	// }
	// for setns process, we don't have to set cloneflags as the process namespaces
	// will only be set via setns syscall
	// data, err := c.bootstrapData(0, state.NamespacePaths, initSetns)
	// if err != nil {
	// 	return nil, err
	// }
	return &setnsProcess{
		cmd:             cmd,
		messageSockPair: messageSockPair,
		logFilePair:     logFilePair,
		config:          c.newInitConfig(p),
		process:         p,
		// bootstrapData:   data,
		// initProcessPid: state.InitProcessPid,
	}
}

func (c *linuxContainer) newInitProcess(p *Process, cmd *exec.Cmd, messageSockPair, logFilePair filePair) (*initProcess, error) {
	cmd.Env = append(cmd.Env, "_LIBCONTAINER_INITTYPE="+string(initStandard))
	nsMaps := make(map[configs.NamespaceType]string)
	for _, ns := range c.config.Namespaces {
		if ns.Path != "" {
			nsMaps[ns.Type] = ns.Path
		}
	}
	_, sharePidns := nsMaps[configs.NEWPID]
	// data, err := c.bootstrapData(c.config.Namespaces.CloneFlags(), nsMaps, initStandard)
	// if err != nil {
	// 	return nil, err
	// }
	init := &initProcess{
		cmd:             cmd,
		messageSockPair: messageSockPair,
		logFilePair:     logFilePair,
		// manager:         c.cgroupManager,
		// intelRdtManager: c.intelRdtManager,
		config:    c.newInitConfig(p),
		container: c,
		process:   p,
		// bootstrapData: data,
		sharePidns: sharePidns,
	}
	c.initProcess = init
	return init, nil
}
