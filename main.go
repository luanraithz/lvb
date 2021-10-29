package main

import (
	"bytes"
	"os/exec"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	cmd := exec.Command("git", "reflog")

	var out bytes.Buffer
	cmd.Stdout = &out
	must(cmd.Run())

	println(out.String())
}
