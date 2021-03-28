package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

var rootfs_dir = "/home/yosuke/works/rootfs"

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

	cmd := exec.Command("/proc/self/exe", append([]string{"start"}, os.Args[2:]...)...) // 1
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// NEW_PID -> pid: 1 で実行
		// NEWNS -> Namespace を作成
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID, //2
	}
	must(cmd.Run())
}

// start() はCLONEされたプロセスであり親プロセスによって制御されている
func start() {
	fmt.Printf("Running start: %v \n", os.Args[2:])
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	must(syscall.Chroot(rootfs_dir)) // 3
	must(os.Chdir("/"))
	must(syscall.Mount("/proc", "/proc", "proc", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, "")) // 4
	must(cmd.Run())
	must(syscall.Unmount("/proc", 0))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
