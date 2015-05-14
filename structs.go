package main

import (
	"fmt"
	"strings"
)

type ghError struct {
	Message string `json:"message"`
}

type ghSearchRes struct {
	Items []ghSearchItem `json:"items"`
}

func (r ghSearchRes) String() string {
	var s string
	for _, m := range r.Items {
		s = s + fmt.Sprintf("%s/%s: %d matches\n", m.Repo.Name, m.trimPath(), len(m.Matches))
	}
	return s
}

type ghSearchItem struct {
	Repo    ghRepo      `json:"repository"`
	Path    string      `json:"path"`
	Matches []ghMatches `json:"text_matches"`
}

func (r ghSearchItem) trimPath() string {
	return strings.TrimLeft(r.Path, "/")
}

type ghRepo struct {
	Name  string `json:"full_name"`
	Desc  string `json:"description"`
	Size  int    `json:"size"`
	Langs ghLangs
}

func (r ghRepo) String() string {
	//return fmt.Sprintf("Repository info:\n  Name: %s\n  Description: %s\n  Size: %d\n  Languages: %s\n",
	//r.Name, r.Desc, r.Size, r.Langs)
	return fmt.Sprintf(`Repository info
  Name:        %s
  Description: %s
  Size:        %d LOC
  Languages:   %s
`,
		r.Name, r.Desc, r.Size, r.Langs)
}

type ghMatches struct {
	Nmatches int
	Fragment string      `json:"fragment"`
	Matches  []ghIndices `json:"matches"`
}

type ghIndices struct {
	Indices []int `json:"indices"`
}

type ghLangs map[string]int

func (l ghLangs) String() string {
	var (
		sz int
		s  string
	)

	for _, v := range l {
		sz += v
	}
	for k, v := range l {
		s = s + fmt.Sprintf("%s: %.2f%%, ", k, 100.0*float64(v)/float64(sz))
	}
	s = strings.TrimRight(s, ", ")
	return s
}

type ghContent struct {
	Content string `json:"content"`
}
