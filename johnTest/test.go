package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {

	// cmd := exec.Command("/usr/local/git/bin/git", "clone", "https://github.com/Makoz/SampleRepo.git")
	cmd := exec.Command("git", "clone", "https://github.com/Makoz/SampleRepo.git")
	err := cmd.Run()
	if err != nil {
		fmt.Println("hey")
		cmd = exec.Command("/usr/local/git/bin/git", "-C", "SampleRepo/", "pull")
	}
	fmt.Println("hey")

	// cmd = exec.Command("/usr/local/go/bin/go", "run", "SampleRepo/test2.go", "1")
	cmd = exec.Command("go", "run", "SampleRepo/test2.go", "1")
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("in all caps: %q\n", out.String())

}
