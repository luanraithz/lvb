package main

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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
	must(cmd.Start())
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
	survey.AskOne(&q, &branch, survey.WithValidator(survey.Required))

	checkout := exec.Command("git", "checkout", branch)
	output, err := checkout.CombinedOutput()
	must(err)
	println(string(output))
}
