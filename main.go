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

func reverse(arr []string) []string {
	if len(arr) == 0 {
		return arr
	}
	return append(reverse(arr[1:]), arr[0])

}

func main() {
	cmd := exec.Command("git", "reflog")

	r, err := cmd.StdoutPipe()
	must(err)
	stderr, err := cmd.StderrPipe()
	must(err)
	err = cmd.Start()

	if err != nil {
		println(bufio.NewScanner(r).Text())
		println(bufio.NewScanner(stderr).Text())
		os.Exit(1)
	}

	branchMap := map[string]bool{} // Wanted a set
	branches := []string{}
	scanner := bufio.NewScanner(r)
	counter := 0
	for scanner.Scan() && counter < 10000 {
		line := scanner.Text()
		if strings.Contains(line, "checkout: moving from") {
			branchesInCommmand := strings.Split(strings.Split(line, "checkout: moving from")[1], "to")
			for _, b := range reverse(branchesInCommmand) {
				trimmed := strings.TrimSpace(b)
				if trimmed != "" {
					exists := branchMap[trimmed]
					if !exists {
						branchMap[trimmed] = true
						branches = append(branches, trimmed)
					}
				}
			}
		}
		counter++
	}

	must(cmd.Wait())
	if len(branches) == 1 {
		println("Didn't find any branch with `git reflog`")
		os.Exit(1)
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
