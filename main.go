package main

import (
	"log"
)

func main() {
	container := &Container{
		ID:      "1",
		Rootfs:  "dev/rootfs",
		Command: []string{"sh"},
	}

	// コンテナでコマンドを実行する
	if err := container.Run(); err != nil {
		log.Fatalf("Error running container: %s\n", err)
	}
}
