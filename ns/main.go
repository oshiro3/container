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

	ex, _ := os.Executable()
	cmd := exec.Command(ex, append([]string{"start"}, os.Args[2:]...)...)

	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// NEW_PID -> pid: 1 で実行
		// NEW_UTS -> hostname を隔離する
		// NEW_IPC -> IPC を隔離する
		Cloneflags:  syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{syscall.SysProcIDMap{0, 1000, 1}},
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
	must(syscall.Mount("/proc", "/proc", "proc", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, ""))
	// must(syscall.Mount("/dev", "/dev", "devtmpfs", syscall.MS_NOEXEC|syscall.MS_STRICTATIME, "mode=755"))
	//	must(syscall.Mount("/dev/mqueue", "/dev/mqueue", "mqueue", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, "mode=755"))
	// 	must(syscall.Mount("dev/shm", "/dev/shm", "devtmpfs", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, "mode=1777,size=65536k"))
	// must(syscall.Mount("/sys", "/sys", "sysfs", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_RDONLY, ""))
	must(cmd.Run())
	must(syscall.Unmount("/proc", 0))
	// must(syscall.Unmount("/dev", 0))
	// must(syscall.Unmount("/sys", 0))
	// must(syscall.Unmount("/dev/shm", 0))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
