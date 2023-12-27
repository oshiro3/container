package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/opencontainers/runc/libcontainer"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/sirupsen/logrus"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
		factory, _ := libcontainer.New("")
		logrus.Info("init")
		if err := factory.StartInitialization(); err != nil {
			logrus.Fatal(err)
		}
		panic("--this line should have never been executed, congratulations--")
	}
}

func main() {
	abs, _ := filepath.Abs("./")

	spec, err := loadSpec("./dev/config.json")
	if err != nil {
		logrus.Fatal(err)
		return
	}

	id := "test-container"
	config, err := specconv.CreateLibcontainerConfig(&specconv.CreateOpts{
		CgroupName: id,
		Spec:       spec,
	})
	if err != nil {
		logrus.Fatal(err)
		return
	}
	// pp.Printf("%+v\n", config.Cgroups)

	// https://github.com/opencontainers/runc/blob/release-1.1/libcontainer/factory_linux.go#L76
	factory, err := libcontainer.New(abs+"/dev/rootfs", libcontainer.InitArgs(os.Args[0], "init"))
	if err != nil {
		logrus.Fatal(err)
	}

	container, err := factory.Create(id, config)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	process := &libcontainer.Process{
		Args:   []string{"/bin/ash"},
		Env:    []string{"PATH=/bin"},
		User:   "daemon",
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Init:   true,
	}

	if err := container.Run(process); err != nil {
		container.Destroy()
		logrus.Fatal(err)
		return
	}

	_, err = process.Wait()
	if err != nil {
		container.Destroy()
		logrus.Fatal(err)
	}
	container.Destroy()
}
