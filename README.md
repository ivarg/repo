# repo

repo is a utility to extract and present data about GitHub repositories.

## Actions

`repos info <repository>`

Present short summary about a repository, such as language composition, lines
of code, contributors, and last update.

`repos search <query> <repository> `

Search the repository with the given query and present a list of hits, with
file, line number, and fragment.

`repos search <query> <user>/<org>

Search all repositories owned by the given user or organization for content
matching the provided query.

`repos list <user>/<org>`

List all repositories pertaining to the given user/org, with a short summary
and ordered by date updated.

