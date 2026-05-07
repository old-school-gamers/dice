# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Use `task` for everything; see `Taskfile.yml` for the full list.

- `task test` — `go test -race -count=1 ./...`
- `task check` — test + vet + fmt + lint + vulncheck
- `task build` — produces `./roll` and `./dice-mcp`
- `task install:roll` / `task install:mcp` — install one binary at a time
- `task register-mcp` — install the MCP server and register it with Claude Code

Single test: `go test ./... -run TestExplodeChain` (subtests use `/` — `-run TestRoll/2d6`).

The library has zero runtime dependencies. The MCP server pulls in
`github.com/modelcontextprotocol/go-sdk` v1.5.0 — pinned to match the version
used by sibling MCP servers in this user's stack (memstore, herald). Bumping
it should be done in concert with the others to keep behavior consistent.

## Architecture

The library is a small parser/evaluator with a modifier pipeline:

```
input → lex → parse (recursive descent) → AST → eval → Result
                                          ↓
                                  diceNode runs its
                                  modifier slice in order
```

Three orthogonal mechanisms compose to express almost any tabletop RPG roll:

1. **AST nodes** (`parser.go`) — `numNode`, `binOpNode`, `diceNode`, `advNode`.
   Each implements `node`: `eval(Roller, *Result) (int, error)`, `String()`,
   `explain()`. `eval` returns the value contribution and may append `Group`s
   to the running `Result`.

2. **Modifiers** (`modifiers.go`, `explode.go`, `success.go`) — `keepMod`,
   `dropMod`, `explodeMod`, `rerollMod`, `successMod`. Each implements
   `modifier`: `apply([]Die, sides, Roller) ([]Die, error)`, `String()`,
   `explain()`. Modifiers run in source order and may append dice (explosions
   add new entries) or mark dice with a different `Status`.

3. **Status flags** (`dice.go`) — `Kept`/`Dropped`/`Exploded`/`Rerolled`/
   `Success`/`Failure`. Single-tag, not a bitfield: a die that exploded and
   then matched `cs>=4` is marked `Success` and loses the "exploded" tag.
   The `dieValue` function in `eval.go` is the single source of truth for
   how each Status contributes to `Subtotal`.

### Things that look weird but are intentional

- **Lexer is dumb; parser splits shared IDENT tokens.** `1d6xcs>=4` lexes
  as `IDENT("xcs")`. `parseModifier` matches the longest-known-keyword
  prefix in `modifierKeywords` (kh before k, xo before x, rr before r, etc.)
  and rewrites the token to drop the consumed prefix. Adding a new modifier
  requires updating that slice and `parseModifierByKeyword`.

- **Success counting flips Subtotal accounting.** When any `successMod` is
  present, unmatched `Kept` dice contribute 0 instead of their face value
  — successes/failures are the only thing that count. `dieValue` takes a
  `successCounting bool` flag, set by `diceNode.eval` after inspecting its
  modifier slice.

- **adv/dis use a scratch `*Result`.** The inner expression is evaluated
  twice, each pass into its own scratch Result, so the first pass's Groups
  don't pollute the second's. Winner and loser dice are coalesced into a
  single output Group with `Notation: "adv(...)"`. Don't change this to
  reuse `res` — the per-roll Group structure of compound expressions
  (e.g. `adv(2d6+1d4)`) gets entangled.

- **Notation is canonicalized at parse time.** `Group.Notation` shows the
  canonical form (e.g. `x4` → `x>=4`, `r3` → `r=3`), not the original input.
  Use `Expression.Source()` for the original.

- **Recursion is capped at 100** (`maxRecursiveRolls` in `explode.go`).
  `1d1x` would otherwise loop forever. The cap returns an error rather
  than truncating silently.

### Roller contract

`Roller.RollDie(sides) (int, error)` must return values in `[1, sides]`.
The default uses `crypto/rand.Int` (rejection-sampled — no modulo bias).
For determinism, tests inject a `fixedRoller`-style implementation via
`RollWith`. A `nil` Roller passed to `Expression.Roll` falls back to the
default.

### Output format split: CLI vs MCP

The CLI (`cmd/roll/render.go`) targets human eyes: ANSI colors on TTY,
plain glyphs (`[1]` for dropped, `*` for success, `!` for explode) when
piped. It's intentionally a separate render layer — the library doesn't
do display.

The MCP server (`cmd/dice-mcp/output.go`) emits a structured JSON
`RollOutput` mirroring `Result`, plus a terse text summary. The structured
payload carries every die's status string so an LLM client doesn't need
to parse the text.

## Repo conventions

- Library is at the module root; binaries live under `cmd/<name>/`.
- Tests use a black-box `package dice_test` and inject a deterministic
  `fixedRoller`. New features should add table-driven tests covering both
  the happy path and at least one parse-error case.
- Modifier `String()` output must round-trip through the parser. If you
  add a new modifier, write a test that parses `expr → e.String()` and
  confirms equivalence.
