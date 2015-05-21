# repo

repo is a utility to extract and present data about GitHub repositories.
As with the GitHub search, queries are interpreted as regular expressions.

It requires the GITHUB_API_TOKEN environment variable to contain a valid
GitHub API token.

## Installation

`go get github.com/ivarg/repo`

## Commands

`repo info <repository>`

Present short summary about a repository, such as language composition, lines
of code, contributors, and last update.

`repo search <query> <repository>`

Search the repository with the given query and present a list of hits, with
file, line number, and fragment.

`repo search <query> {<user>|<org>}`

Search all repositories owned by the given user or organization for content
matching the provided query.

`repo list <user>/<org>`

List all repositories pertaining to the given user/org, with a short summary
and ordered by date updated.

`repo cat <repository>/<path>`

Print the contents of a file to stdout.

## Examples

Print a short summary of repository 'myrepo', owned by kitty:

`$ repo info kitty/myrepo`

Search through user kitty's repository 'myrepo' for occurrences of the string
"http.StatusBadRequest":

`$ repo search http\\.StatusBadRequest kitty/myrepo`

Enclose multi-word search terms in quotes:

`$ repo search "package main" kitty`

