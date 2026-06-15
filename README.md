# spacedevs

Space launch and astronaut data from The Space Devs Launch Library 2.

`spacedevs` is a single pure-Go binary. It reads public data from
[ll.thespacedevs.com](https://ll.thespacedevs.com) over plain HTTPS, shapes it
into clean records, and prints output that pipes into the rest of your tools. No
API key required (anonymous access, rate-limited to 15 req/hour).

The same package is also a [resource-URI driver](#use-it-as-a-resource-uri-driver),
so a host program like [ant](https://github.com/tamnd/ant) can address
spacedevs data as `spacedevs://` URIs.

## Install

```bash
go install github.com/tamnd/spacedevs-cli/cmd/spacedevs@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/spacedevs-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/spacedevs:latest --help
```

## Usage

```bash
spacedevs upcoming                         # upcoming rocket launches
spacedevs upcoming --limit 20             # more results
spacedevs upcoming --search SpaceX        # filter by name
spacedevs launches --search Falcon        # all/historical launches
spacedevs astronauts --search Armstrong   # search astronauts
spacedevs agencies --type Government      # space agencies by type
spacedevs spacecraft --search Dragon      # spacecraft catalog
spacedevs --help                          # the whole command tree
```

Every command shares one output contract: `-o table|json|jsonl|csv|tsv|url|raw`,
`--fields` to pick columns, `--template` for a custom line, and `-n` to limit.
The default adapts to where output goes (a table on a terminal, JSONL in a
pipe), so the same command reads well by hand and parses cleanly downstream.

```bash
spacedevs upcoming -o json | jq '.[].name'
spacedevs astronauts --search Gagarin --fields name,nationality,dob
spacedevs agencies --type Government -o csv
```

## Serve it

The same operations are available over HTTP and as an MCP tool set for agents,
with no extra code:

```bash
spacedevs serve --addr :7777    # GET /v1/upcoming  returns NDJSON
spacedevs mcp                   # speak MCP over stdio
```

## Use it as a resource-URI driver

`spacedevs` registers a `spacedevs` domain the way a program registers a
database driver with `database/sql`. A host enables it with one blank import:

```go
import _ "github.com/tamnd/spacedevs-cli/spacedevs"
```

Then [ant](https://github.com/tamnd/ant) (or any program that links the package)
dereferences `spacedevs://` URIs without knowing anything about the API.

## Development

```
cmd/spacedevs/   thin main: hands cli.NewApp to kit.Run
cli/             assembles the kit App from the spacedevs domain
spacedevs/       the library: HTTP client, data models, and domain.go (the driver)
docs/            tago documentation site
```

```bash
make build      # ./bin/spacedevs
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.0
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so the first
release works with no extra secrets.

## License

Apache-2.0. See [LICENSE](LICENSE).
