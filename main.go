package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		"cat":    cat,
	}
)

type command func(args ...string)

func init() {
	token = os.Getenv("GITHUB_API_TOKEN")
	if token == "" {
		fmt.Println("repo needs GITHUB_API_TOKEN to be set to a valid GitHub API token")
		os.Exit(1)
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
  cat        Print the contents of a file

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
		searchRepo(q, owner, repo)
		return
	}
	searchOwner(q, owner)
}

func searchOwner(q, owner string) {
	res := dosearch(fmt.Sprintf("https://api.github.com/search/code?q=\"%s\"+user:%s", q, owner))
	hits := make(map[string]int)
	for _, item := range res.Items {
		cnt := hits[item.Repo.Name]
		hits[item.Repo.Name] = cnt + len(item.Matches)
	}
	for r, m := range hits {
		fmt.Printf("%s: %d files with matches\n", r, m)
	}
}

type hit struct {
	file string
	line int
	frag string
}

func searchRepo(q, owner, repo string) {
	res := dosearch(fmt.Sprintf("https://api.github.com/search/code?q=\"%s\"+repo:%s/%s", q, owner, repo))
	hits := make(map[string]int)
	for _, item := range res.Items {
		hits[item.trimPath()] = 0
	}
	if len(hits) == 0 {
		return
	}
	re := regexp.MustCompile(q)
	for f, _ := range hits {
		fmt.Printf("%s:\n", f)
		content := getfile(owner, repo, f)
		lines := strings.Split(content, "\n")
		for i, l := range lines {
			if re.MatchString(l) {
				fmt.Printf("  %d %s\n", i+1, l)
			}
		}
	}
}

func dosearch(u string) ghSearchRes {
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

	return rr
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

func cat(args ...string) {
	pth := strings.Split(args[0], "/")
	o, repo := pth[0], pth[1]
	file := strings.Join(pth[2:], "/")

	fmt.Println(getfile(o, repo, file))
}

func getfile(owner, repo, file string) string {
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, file)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	b, _ := ioutil.ReadAll(resp.Body)

	var emsg ghError
	json.Unmarshal(b, &emsg)
	if emsg.Message != "" {
		fmt.Printf("Error: %s\n", emsg.Message)
		os.Exit(1)
	}

	var c ghContent
	if err = json.Unmarshal(b, &c); err != nil {
		panic(err)
	}

	txt, err := base64.StdEncoding.DecodeString(c.Content)
	if err != nil {
		log.Fatal("Error:", err)
	}

	return string(txt)
}
