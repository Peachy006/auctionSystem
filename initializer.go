// Run from the project root: go run ./initializer
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	ids := "1,2,3"

	for _, id := range []string{"1", "2", "3"} {
		logFile, _ := os.Create("node-" + id + ".log")

		cmd := exec.Command("go", "run", "./node", "-id", id, "-nodes", ids)
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		cmd.Start()

		fmt.Println("started node", id)
	}

	fmt.Println("start a client with: go run ./client anna", ids)
}
