package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	token string
	owner *string

	cmds = map[string]command{
		"info":   info,
		"search": search,
		"list":   list,
	}
)

type command func(args ...string)

func init() {
	token = os.Getenv("GH_TOKEN")
	if token == "" {
		log.Fatal("Error: No GH_TOKEN found")
	}
}

func main() {
	if len(os.Args) == 1 {
		printUsage()
		return
	}
	cmd, ok := cmds[os.Args[1]]
	if !ok {
		fmt.Println("unknown command")
		printUsage()
		return
	}
	cmd(os.Args[2:]...)
}

func printUsage() {
	fmt.Println(`Repo is a tool for querying GitHub repositories from the command line

usage: repo <command> [<args>]

Commands:

  info       Print a brief summary about a repository
  search     Search the file contents in a repository
  list       List all repositories pertaining to a user or organization

Examples:

  Print a short summary of repository 'myrepo', owned by kitty
  repos info kitty/myrepo

  Search through user kitty's repository 'myrepo' for occurrences of the string
  "http.StatusBadRequest".
  repos search http\.StatusBadRequest kitty/myrepo

  repos search "package main" kitty
`)
}

func info(args ...string) {
	repo := args[0]
	u := fmt.Sprintf("https://api.github.com/repos/%s", repo)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	var rr ghRepo
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&rr); err != nil {
		panic(err)
	}

	rr.Langs = langs(repo)
	fmt.Println(rr)
}

func langs(repo string) ghLangs {
	u := fmt.Sprintf("https://api.github.com/repos/%s/languages", repo)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	var lngs ghLangs
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&lngs); err != nil {
		panic(err)
	}
	return lngs
}

// args[0] == "search"
func search(args ...string) {
	q := args[0]
	owner := args[1]
	if ur := strings.Split(args[1], "/"); len(ur) > 1 {
		owner = ur[0]
		repo := ur[1]
		dosearch(fmt.Sprintf("https://api.github.com/search/code?q=\"%s\"+repo:%s/%s", q, owner, repo))
		return
	}
	dosearch(fmt.Sprintf("https://api.github.com/search/code?q=\"%s\"+user:%s", q, owner))
}

func dosearch(u string) {
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Add("Accept", "application/vnd.github.v3.text-match+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	var rr ghSearchRes
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&rr); err != nil {
		panic(err)
	}

	fmt.Println(rr)
}

func list(args ...string) {
	owner := args[0]
	utempl := "https://api.github.com/users/%s/repos?page=%d"
	if usertype(owner) == "Organization" {
		utempl = "https://api.github.com/orgs/%s/repos?page=%d"
	}

	var repos []string
	for i := 1; true; i++ {
		u := fmt.Sprintf(utempl, owner, i)
		req, _ := http.NewRequest("GET", u, nil)
		req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}

		var rr []ghRepo
		dec := json.NewDecoder(resp.Body)
		if err = dec.Decode(&rr); err != nil {
			panic(err)
		}
		for _, r := range rr {
			repos = append(repos, r.Name)
		}

		// If link then more pages
		link := resp.Header.Get("Link")
		match, err := regexp.MatchString("rel=\"next\"", link)
		if err != nil {
			panic(err)
		}
		if !match {
			break
		}
	}

	fmt.Printf("Repositories: %d\n", len(repos))
	for _, r := range repos {
		fmt.Printf("  %s\n", r)
	}
}

func usertype(owner string) string {
	u := fmt.Sprintf("https://api.github.com/users/%s", owner)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	var rr map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&rr); err != nil {
		panic(err)
	}
	return rr["type"].(string)
}
