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
	case "child":
		child()
	default:
		fmt.Println("invalid args")
	}
}

// run() は親プロセスでありホストと同じ状態を持つ
func run() {
	fmt.Printf("Running %v \n", os.Args[2:])

	// cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	 cmd := exec.Command(os.Args[2], os.Args[3:]...)
	 cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// NEW_PID -> pid: 1 で実行
		// NEW_UTS -> hostname を隔離する
		// NEW_IPC -> IPC を隔離する
		// Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	Cloneflags: syscall.CLONE_NEWNS,
  }
	check(cmd.Run())
}

// child() はCLONEされたプロセスであり親プロセスによって制御されている
func child() {
	fmt.Printf("Running child: %v \n", os.Args[2:])
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	// check(syscall.Sethostname([]byte("container")))
	check(syscall.Chroot("/home/yosuke/works/rootfs"))
	check(os.Chdir("/"))
	// check(syscall.Mount("/proc", "/proc", "proc", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, ""))
	// check(syscall.Mount("/dev", "/dev", "devtmpfs", syscall.MS_NOEXEC|syscall.MS_STRICTATIME, "mode=755"))
	//	check(syscall.Mount("/dev/mqueue", "/dev/mqueue", "mqueue", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, "mode=755"))
// 	check(syscall.Mount("dev/shm", "/dev/shm", "devtmpfs", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, "mode=1777,size=65536k"))
	// check(syscall.Mount("/sys", "/sys", "sysfs", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_RDONLY, ""))
	check(cmd.Run())
	// check(syscall.Unmount("/proc", 0))
	// check(syscall.Unmount("/dev", 0))
	// check(syscall.Unmount("/sys", 0))
	// check(syscall.Unmount("/dev/shm", 0))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
