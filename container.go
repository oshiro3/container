package main

import (
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
	if err := syscall.Chroot(c.Rootfs); err != nil {
		return err
	}

	if err := os.Chdir("/"); err != nil {
		return err
	}

	cmd := exec.Command(c.Command[0], c.Command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
