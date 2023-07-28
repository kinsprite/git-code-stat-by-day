package main

import (
	// "fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/olekukonko/tablewriter"
)

func NumberInAbs(value int, maxAbs int) int {
	if value > maxAbs {
		return maxAbs
	}

	if value < -maxAbs {
		return -maxAbs
	}

	return value
}

type GitStatSum struct {
	Addition     int
	Deletion     int
	Modification int
	CommitCount  int
	DayCount     int
}

func (sum *GitStatSum) Append(fileStat *object.FileStat) {
	sum.Addition += fileStat.Addition
	sum.Deletion += fileStat.Deletion
	sum.Modification += fileStat.Addition - fileStat.Deletion
}

func (sum *GitStatSum) PlusCommit(_ *object.Commit) {
	sum.CommitCount++
}

func (sum *GitStatSum) Add(other *GitStatSum) {
	sum.Addition += other.Addition
	sum.Deletion += other.Deletion
	sum.Modification += other.Modification
	sum.CommitCount += other.CommitCount
	sum.DayCount += other.DayCount
}

type GitStatByDate struct {
	// key: Date string
	stats map[string]*GitStatSum
}

func (byDate *GitStatByDate) Append(commit *object.Commit, pattern *regexp.Regexp) {
	// make the map exist
	if byDate.stats == nil {
		byDate.stats = make(map[string]*GitStatSum)
	}

	// make the date sum exist
	dateKey := commit.Author.When.Format(time.DateOnly)
	sum, ok := byDate.stats[dateKey]

	if !ok {
		sum = &GitStatSum{}
		byDate.stats[dateKey] = sum
	}

	sum.PlusCommit(commit)

	// for loop file changes
	fileStats, err := commit.Stats()

	if err == nil {
		for _, fileStat := range fileStats {
			if pattern.MatchString(fileStat.Name) {
				sum.Append(&fileStat)
			}
		}
	}
}

func (byDate *GitStatByDate) Add(other *GitStatByDate) {
	// make the map exist
	if byDate.stats == nil {
		byDate.stats = make(map[string]*GitStatSum)
	}

	for dateKey := range other.stats {
		// make the date sum exist
		sum, ok := byDate.stats[dateKey]

		if !ok {
			sum = &GitStatSum{}
			byDate.stats[dateKey] = sum
		}

		sum.Add(other.stats[dateKey])
	}
}

func (byDate *GitStatByDate) Summary(maxAbsPerDate int) GitStatSum {
	sum := GitStatSum{}

	if byDate.stats != nil {
		for _, stat := range byDate.stats {
			sum.Addition += NumberInAbs(stat.Addition, maxAbsPerDate)
			sum.Deletion += NumberInAbs(stat.Deletion, maxAbsPerDate)
			sum.Modification += NumberInAbs(stat.Modification, maxAbsPerDate)
			sum.CommitCount += stat.CommitCount
			sum.DayCount++
		}
	}

	return sum
}

type GitStatByEmail struct {
	// key: Email string
	stats map[string]*GitStatByDate
}

type GitStatByEmailSummary struct {
}

func (byEmail *GitStatByEmail) Append(commit *object.Commit, pattern *regexp.Regexp) {
	// make the map exist
	if byEmail.stats == nil {
		byEmail.stats = make(map[string]*GitStatByDate)
	}

	// make the email sum exist
	emailKey := commit.Author.Email
	byDate, ok := byEmail.stats[emailKey]

	if !ok {
		byDate = &GitStatByDate{}
		byEmail.stats[emailKey] = byDate
	}

	// append to the email's by date
	byDate.Append(commit, pattern)
}

func (byEmail *GitStatByEmail) Add(other *GitStatByEmail) {
	// make the map exist
	if byEmail.stats == nil {
		byEmail.stats = make(map[string]*GitStatByDate)
	}

	for emailKey := range other.stats {
		// make the email sum exist
		byDate, ok := byEmail.stats[emailKey]

		if !ok {
			byDate = &GitStatByDate{}
			byEmail.stats[emailKey] = byDate
		}

		byDate.Add(other.stats[emailKey])
	}
}

func (byEmail *GitStatByEmail) Summary(maxAbsPerDate int) {
	if byEmail.stats == nil {
		return
	}

	keys := make([]string, 0, len(byEmail.stats))
	for k := range byEmail.stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Email", "Add", "Delete", "Modify", "Commits", "Days", "A/D", "M/D"})

	for _, email := range keys {
		byDate, ok := byEmail.stats[email]

		if ok {
			sum := byDate.Summary(maxAbsPerDate)
			table.Append([]string{
				email,
				strconv.Itoa(sum.Addition),
				strconv.Itoa(sum.Deletion),
				strconv.Itoa(sum.Modification),
				strconv.Itoa(sum.CommitCount),
				strconv.Itoa(sum.DayCount),
				strconv.Itoa(sum.Addition / sum.DayCount),
				strconv.Itoa(sum.Modification / sum.DayCount),
			})
			// fmt.Printf("%s +%d -%d %d \n", email, sum.Addition, sum.Deletion, sum.Modification)
		}
	}

	table.Render()
}
