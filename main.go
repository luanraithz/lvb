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

func getErrorCode(err error) int {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return exiterr.ExitCode()
		}
		return 1
	}
	return 1
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
	_, err := os.Stat(".git")
	cmd := exec.Command("git", "reflog", "--date=iso")

	r, err := cmd.StdoutPipe()
	if err != nil {
		println(bufio.NewScanner(r).Text())
		os.Exit(getErrorCode(err))
	}
	stderr, err := cmd.StderrPipe()
	err = cmd.Start()

	if err != nil {
		println(bufio.NewScanner(stderr).Text())
		os.Exit(getErrorCode(err))
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
			branchesInCommand := strings.Split(strings.Split(line, "checkout: moving from")[1], " to ")
			for _, b := range reverse(branchesInCommand) {
				trimmed := strings.TrimSpace(b)
				if trimmed != "" {
					v, exists := branchMap[trimmed]
					if !exists || v.Commit == "" {
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
					}
					if !exists {
						branches = append(branches, trimmed)
					}
				}
			}
		}
		counter++
	}

	if err := cmd.Wait(); err != nil {
		println(err.Error())
		if t := bufio.NewScanner(r).Text(); t != "" {
			println(t)
		}
		if t := bufio.NewScanner(stderr).Text(); t != "" {
			println(t)
		}
		os.Exit(getErrorCode(err))
	}
	if len(branches) == 1 {
		println("Didn't find any branch with `git reflog`")
		os.Exit(1)
	}
	qe := []*survey.Question{
		{
			Prompt: &survey.Select{
				Message: "Which branch do you want to go?",
				Options: branches,
				Filter: func(f string, opValue string, i int) bool {
					info, _ := branchMap[opValue]
					return strings.Contains(opValue, f) || strings.Contains(info.Commit, f)
				},
				Comment: func(opt string, index int) string {
					info, ex := branchMap[opt]
					if !ex || info.Commit == "" {
						return "Never commited"
					}
					return fmt.Sprintf("%s (%s)", info.Commit, info.Date.Format("Jan 02 Mon"))
				},
			},
			Name: "branch",
		},
	}
	ans := map[string]string{}
	err = survey.Ask(qe, &ans, survey.WithValidator(survey.Required))
	branch := ans["branch"]

	if branch == "" {
		os.Exit(0)
	}

	checkout := exec.Command("git", "checkout", branch)
	output, _ := checkout.CombinedOutput()
	println(string(output))
}
