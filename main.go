package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type InputConfig struct {
	Repos     []string `json:"repos"`
	Since     string   `json:"since"`
	Until     string   `json:"until"`
	MaxAbs    int32    `json:"maxAbs"`
	SkipMerge bool     `json:"skipMerge"`
	Pattern   string   `json:"pattern"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var inputFile string
	if len(os.Args) > 1 {
		inputFile = os.Args[1]
	} else {
		inputFile = "input.json"
	}

	inputJSON, err := os.ReadFile(inputFile)
	check(err)
	fmt.Println(string(inputJSON))

	var input InputConfig
	err = json.Unmarshal(inputJSON, &input)
	check(err)

	fmt.Println(input)

	// time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	since, err := time.Parse(time.DateOnly, input.Since)
	check(err)
	until, err := time.Parse(time.DateOnly, input.Until)
	check(err)

	pattern, err := regexp.Compile(input.Pattern)
	check(err)

	byEmail := GitStatByEmail{}

	for _, repoPath := range input.Repos {
		repo, err := git.PlainOpen(repoPath)
		check(err)

		commitIter, err := repo.CommitObjects()
		check(err)

		limitOpt := object.LogLimitOptions{Since: &since, Until: &until}
		limitIter := object.NewCommitLimitIterFromIter(commitIter, limitOpt)

		limitIter.ForEach(func(commit *object.Commit) error {
			if input.SkipMerge && len(commit.ParentHashes) > 1 {
				return nil
			}

			byEmail.Append(commit, pattern)

			// stats, err := commit.Stats()

			// if err == nil {
			// 	byEmail.Append(commit, pattern)
			// 	fmt.Println(commit.Author.Email, "\t", commit.Author.When.Format(time.DateTime))

			// 	for _, stat := range stats {
			// 		if pattern.MatchString(stat.Name) {
			// 			fmt.Printf("+%d\t-%d\t%s\n", stat.Addition, stat.Deletion, stat.Name)
			// 		}
			// 	}

			// }

			return nil
		})
	}

	byEmail.Summary(int(input.MaxAbs))
}
