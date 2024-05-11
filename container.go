package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

// Container はコンテナの基本情報を保持する
type Container struct {
	ID      string
	Rootfs  string
	Command []string
}

// Run はコンテナ内で指定されたコマンドを実行する
func (c *Container) Run() error {
	// initプロセスとして子プロセスを起動
	cmd := exec.Command("/proc/self/exe", append([]string{"init"}, c.Command...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	// initプロセスとして起動した子プロセスの終了を待つ
	return cmd.Wait()
}

// Initは実際にユーザが指定したコマンドを実行する
func (c *Container) Init() error {
	// rootfsにchrootするなどの処理を行う
	if err := syscall.Chroot(c.Rootfs); err != nil {
		return err
	}
	log.Println("chroot to rootfs")
	if err := os.Chdir("/"); err != nil {
		return err
	}
	log.Println("chdir to /")
	syscall.Mount("proc", "proc", "proc", 0, "")
	defer syscall.Unmount("proc", 0)

	// ユーザのコマンドを実行します
	log.Printf("exec command: %v", os.Args[2:])
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Println("command end")
	return nil
}
