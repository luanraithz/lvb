package main

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
func main() {
	cmd := exec.Command("git", "reflog")

	r, err := cmd.StdoutPipe()
	must(err)
	done := make(chan bool)
	scanner := bufio.NewScanner(r)
	branchMap := map[string]bool{} // Wanted a set
	go func() {
		counter := 0
		for scanner.Scan() && counter < 10000 {
			line := scanner.Text()
			if strings.Contains(line, "checkout: moving from") {
				branches := strings.Split(strings.Split(line, "checkout: moving from")[1], "to")
				for _, b := range branches {
					trimmed := strings.TrimSpace(b)
					if trimmed != "" {
						branchMap[trimmed] = true
					}
				}
			}
			counter++
		}
		done <- true

	}()
	stderr, err := cmd.StderrPipe()
	must(err)
	err = cmd.Start()
	if err != nil {
		println(bufio.NewScanner(r).Text())
		println(bufio.NewScanner(stderr).Text())
		os.Exit(1)
	}
	<-done
	must(cmd.Wait())
	branches := []string{}
	for k := range branchMap {
		branches = append(branches, k)
	}
	q := survey.Select{
		Message: "Which branch do you want to go?",
		Options: branches,
	}
	var branch string
	err = survey.AskOne(&q, &branch, survey.WithValidator(survey.Required))
	if err != nil {
		if err == terminal.InterruptErr {
			return
		}
		panic(err)
	}

	checkout := exec.Command("git", "checkout", branch)
	output, _ := checkout.CombinedOutput()
	println(string(output))
}
