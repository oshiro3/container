package main

import (
	"log"
	"os"
)

func main() {
	container := &Container{
		ID:      "1",
		Rootfs:  "dev/rootfs",
		Command: []string{"sh"},
	}
	if len(os.Args) > 1 && os.Args[1] == "init" {
		log.Println("init process")
		if err := container.Init(); err != nil {
			log.Fatal(err)
		}
		log.Println("init process end")
	} else {
		log.Println("run process")
		if err := container.Run(); err != nil {
			log.Fatal(err)
		}
	}
}
