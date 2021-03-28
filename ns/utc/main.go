package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "start":
		start()
	default:
		fmt.Println("invalid args")
	}
}

// run() は親プロセスでありホストと同じ状態を持つ
func run() {
	fmt.Printf("Running %v \n", os.Args[2:])

	cmd := exec.Command("/proc/self/exe", append([]string{"start"}, os.Args[2:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// NEW_UTS -> hostname を隔離する
		Cloneflags: syscall.CLONE_NEWUTS,
	}
	must(cmd.Run())
}

// start() はCLONEされたプロセスであり親プロセスによって制御されている
func start() {
	fmt.Printf("Running start: %v \n", os.Args[2:])
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("/home/yosuke/works/rootfs"))
	must(os.Chdir("/"))
	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
