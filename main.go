package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
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

func extractDateFromLine(line string) *time.Time {
	r, _ := regexp.Compile("\\{(\\d|\\-)*")
	date := strings.TrimLeftFunc(r.FindString(line), func(r rune) bool { return r == '{' })

	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return &t
}

func main() {
	cmd := exec.Command("git", "reflog", "--date=iso")

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

	currentCommit := ""
	var currentDate *time.Time
	type BranchInfo struct {
		Commit string
		Date   time.Time
		Name   string
	}
	branchMap := map[string]BranchInfo{} // Wanted a set
	branches := []string{}
	scanner := bufio.NewScanner(r)
	counter := 0
	for scanner.Scan() && counter < 10000 {
		line := scanner.Text()
		if strings.Contains(line, ": commit:") {
			if currentCommit == "" {
				currentCommit = strings.Split(line, ": commit: ")[1]
				currentDate = extractDateFromLine(line)
			}
		}
		if strings.Contains(line, "checkout: moving from") {
			branchesInCommmand := strings.Split(strings.Split(line, "checkout: moving from")[1], " to ")
			for _, b := range reverse(branchesInCommmand) {
				trimmed := strings.TrimSpace(b)
				if trimmed != "" {
					_, exists := branchMap[trimmed]
					if !exists {
						fmt.Printf("%v", currentDate)
						if currentDate == nil {
							currentDate = extractDateFromLine(line)
						}
						branchMap[trimmed] = BranchInfo{
							Commit: currentCommit,
							Name:   trimmed,
							Date:   *currentDate,
						}
						currentCommit = ""
						currentDate = nil
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
	qe := []*survey.Question{
		{
			Prompt: &survey.Select{
				Message: "Which branch do you want to go?",
				Options: branches,
				Comment: func(opt string, index int) string {
					info, ex := branchMap[opt]
					if !ex || info.Commit == "" {
						return "Never commited"
					}
					return fmt.Sprintf("%s (%s)", info.Commit, info.Date.Format("Jan 01 Mon"))
				},
			},
			Name: "branch",
		},
	}
	ans := map[string]string{}
	err = survey.Ask(qe, &ans, survey.WithValidator(survey.Required))
	branch := ans["branch"]

	checkout := exec.Command("git", "checkout", branch)
	output, _ := checkout.CombinedOutput()
	println(string(output))
}
