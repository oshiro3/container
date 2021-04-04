package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "start":
		start()
	case "spec":
		spec()
	default:
		fmt.Println("invalid args")
	}
}

func spec() (*Config, error) {
	ex, _ := os.Executable()

	spec, e := loadSpec(filepath.Join(filepath.Dir(ex), "config.json"))
	fmt.Printf("%T\n", spec)
	must(e)
	config := &Config{
		Rootfs:     spec.Root.Path,
		Readonlyfs: spec.Root.Readonly,
		Hostname:   spec.Hostname,
		Sysctl:     spec.Linux.Sysctl,
	}

	// Add Process
	config.Process = &Process{
		spec.Process.Args,
		spec.Process.Env,
	}

	// Add Namespaces
	for _, ns := range spec.Linux.Namespaces {
		t, exists := namespaceMapping[ns.Type]
		if !exists {
			fmt.Errorf("namespace %q does not exist", ns)
			return config, nil
		}
		if config.Namespaces.Contains(t) {
			fmt.Errorf("malformed spec file: duplicated ns %q", ns)
			return config, nil
		}
		config.Namespaces.Add(t, ns.Path)
	}
	if config.Namespaces.Contains(NEWNET) && config.Namespaces.PathOf(NEWNET) == "" {
		config.Networks = []*Network{
			{
				Type: "loopback",
			},
		}
	}
	if config.Namespaces.Contains(NEWUSER) {
		if err := setupUserNamespace(spec, config); err != nil {
			return config, nil
		}
	}

	// UidMappings
	config.UidMappings = []IDMap{IDMap{
		int(spec.Process.User.UID),
		os.Getuid(),
		1,
	}}

	// GidMappings
	config.GidMappings = []IDMap{IDMap{
		int(spec.Process.User.GID),
		os.Getgid(),
		32000,
	}}

	// Mount
	// Capabilities
	// Networks
	// Routes
	return config, nil

}

// run() は親プロセスでありホストと同じ状態を持つ
func run() {
	conf, err := spec()
	must(err)

	ex, _ := os.Executable()
	cmd := exec.Command(ex, append([]string{"start"}, conf.Process.Args[1:]...)...)

	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	procAttr := &syscall.SysProcAttr{}

	// procAttr.UidMappings = []syscall.SysProcIDMap{
	// 	syscall.SysProcIDMap{
	// 		conf.UidMappings[0].ContainerID,
	// 		conf.UidMappings[0].HostID,
	// 		conf.UidMappings[0].Size,
	// 	},
	// }
	procAttr.UidMappings = []syscall.SysProcIDMap{syscall.SysProcIDMap{0, 1000, 1}}
	fmt.Printf("ContainerID: %v\n", conf.UidMappings[0].ContainerID)
	fmt.Printf("HostID: %v\n", conf.UidMappings[0].HostID)
	// procAttr.GidMappings = []syscall.SysProcIDMap{
	// 	syscall.SysProcIDMap{
	// 		conf.GidMappings[0].ContainerID,
	// 		conf.GidMappings[0].HostID,
	// 		conf.GidMappings[0].Size,
	// 	},
	// }

	for _, ns := range conf.Namespaces {
		switch ns.Type {
		case NEWPID:
			procAttr.Cloneflags = procAttr.Cloneflags | syscall.CLONE_NEWPID
		case NEWIPC:
			procAttr.Cloneflags = procAttr.Cloneflags | syscall.CLONE_NEWIPC
		case NEWUTS:
			procAttr.Cloneflags = procAttr.Cloneflags | syscall.CLONE_NEWUTS
		case NEWNS:
			procAttr.Cloneflags = procAttr.Cloneflags | syscall.CLONE_NEWNS
		}
	}
	procAttr.Cloneflags = procAttr.Cloneflags | syscall.CLONE_NEWUSER
	cmd.SysProcAttr = procAttr
	must(cmd.Run())
}

// start() はCLONEされたプロセスであり親プロセスによって制御されている
func start() {
	conf, _ := spec()
	fmt.Printf("Running start: %v \n", conf.Process.Args)
	cmd := exec.Command(conf.Process.Args[0])
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	must(syscall.Sethostname([]byte(conf.Hostname)))
	must(syscall.Chroot(conf.Rootfs))
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
