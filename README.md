# dice

[![CI](https://github.com/oldschoolgamers/dice/actions/workflows/ci.yml/badge.svg)](https://github.com/oldschoolgamers/dice/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/oldschoolgamers/dice.svg)](https://pkg.go.dev/github.com/oldschoolgamers/dice)
[![Go Report Card](https://goreportcard.com/badge/github.com/oldschoolgamers/dice)](https://goreportcard.com/report/github.com/oldschoolgamers/dice)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Tabletop RPG dice roller for Go: a library, a `roll` CLI, and a `dice-mcp`
server for use with Claude Code or any MCP client.

```
$ roll 4d6kh3
4d6kh3 = 14
  4d6kh3: 6 5 3 [1]

$ roll 'adv(1d20)+5 >= 15'
adv(1d20)+5 = 21 HIT
  adv(1d20): 16 [9]
```

## Install

```sh
task install         # builds and installs both binaries via go install
task install:roll    # roll CLI only
task install:mcp     # dice-mcp server only
task register-mcp    # install:mcp + register with Claude Code at user scope
```

Without `task`:

```sh
go install github.com/oldschoolgamers/dice/cmd/roll@latest
go install github.com/oldschoolgamers/dice/cmd/dice-mcp@latest
```

## Notation

| Form              | Meaning                                                       |
|-------------------|---------------------------------------------------------------|
| `XdY`, `dY`       | roll X dice with Y sides (count defaults to 1)                |
| `+`, `-`, `()`    | integer arithmetic and grouping                               |
| `khN`, `klN`, `k` | keep highest / lowest N (`k` = `kh`; N defaults to 1)         |
| `dhN`, `dlN`      | drop highest / lowest N                                       |
| `x[cmp]`, `xo`    | exploding dice; `xo` = once only                              |
| `r[cmp]`, `rr`    | reroll once / recursively (comparison required)               |
| `cs[cmp]`         | count successes (each match contributes +1, others 0)         |
| `cf[cmp]`         | count failures (each match contributes -1)                    |
| `adv(expr)`       | roll twice, take the higher                                   |
| `dis(expr)`       | roll twice, take the lower                                    |
| `>= N`, `<= N`, …trailing target; sets Hit without changing Total           |

`cmp` is a comparison: `>=N`, `<=N`, `>N`, `<N`, `=N`, or a bare `N`
(interpreted as `>=N` for `x` / `cs`, `<=N` for `cf`, `=N` for `r`/`rr`).
Recursion in `x` and `rr` is capped at 100 to bound runaway expressions
like `1d1x`.

Examples:

```
2d6+3                  basic damage
4d6kh3                 D&D 5e ability score
adv(1d20)+5 >= 15      attack with advantage against AC 15
5d10cs>=8              World of Darkness pool, count 8s and up
5d10cs>=8cf=1          successes minus botches
1d6x                   exploding d6
1d6r1                  reroll 1s once
1d6rr<3                reroll any 1 or 2, repeatedly
```

## Library

```go
import "github.com/oldschoolgamers/dice"

res, err := dice.Roll("2d6+3")
fmt.Println(res.Total)        // e.g. 9

// Pre-parse for repeated rolls (useful in servers).
expr, _ := dice.Parse("4d6kh3")
res, _ = expr.Roll(nil)       // nil Roller → crypto/rand

// Inject a deterministic Roller for tests.
res, _ = dice.RollWith("1d20", myFixedRoller)

// Plain-English description.
fmt.Println(expr.Explain())   // "roll 4 6-sided dice, keep the highest 3"
```

`Result.Groups[i].Dice[j].Status` distinguishes `Kept`, `Dropped`,
`Exploded`, `Rerolled`, `Success`, `Failure` — enough to render any
expression's outcome.

The default `Roller` uses `crypto/rand`; rejection sampling in
`crypto/rand.Int` makes all values uniform on `[1, sides]`.

## MCP server

`dice-mcp` exposes two tools over stdio:

- `roll(expression)` — rolls and returns a structured payload plus a
  short text summary. Per-die status fields included.
- `explain(expression)` — returns the prose description without rolling.

Register with Claude Code:

```sh
task register-mcp
# or, manually:
claude mcp add dice -s user -- /path/to/dice-mcp
```

## Development

```sh
task            # list tasks
task test       # go test -race -count=1 ./...
task check      # test + vet + fmt + lint + vulncheck
task build      # produce ./roll and ./dice-mcp without installing
task clean
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution guide and
[SECURITY.md](SECURITY.md) for vulnerability reporting.

## License

Apache-2.0 — see [LICENSE](LICENSE).
