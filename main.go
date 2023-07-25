package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type InputConfig struct {
	Repos   []string `json:"repos"`
	Authors []string `json:"authors"`
	Since   string   `json:"since"`
	Until   string   `json:"until"`
	MaxAbs  int32    `json:"maxAbs"`
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

	check(err)
	for _, repoPath := range input.Repos {
		repo, err := git.PlainOpen(repoPath)
		check(err)

		commitIter, err := repo.CommitObjects()
		check(err)

		limitOpt := object.LogLimitOptions{Since: &since, Until: &until}
		limitIter := object.NewCommitLimitIterFromIter(commitIter, limitOpt)

		limitIter.ForEach(func(commit *object.Commit) error {
			stats, err := commit.Stats()

			if err == nil {
				fmt.Println(commit.Message)

				for _, stat := range stats {
					fmt.Printf("%d\t%d\t%s\n", stat.Addition, stat.Deletion, stat.Name)
				}
			}

			return nil
		})
	}

}
