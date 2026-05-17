# Contributing

Open an issue before sending a PR for substantial changes — easier to
agree on the approach up front than to rework it in review.

## Local development

```sh
task            # list tasks
task test       # go test -race -count=1 ./...
task check      # test + vet + fmt + lint + vulncheck
```

`task check` must pass before a PR will be merged. Requires Go 1.25.10
or newer.

## Conventions

- Run `task fmt:fix` before committing.
- The library (module root) has zero runtime dependencies. Don't add
  any. Only `cmd/dice-mcp` is permitted to pull in the MCP SDK.
- Tests live in `package dice_test` (black-box) and inject a
  deterministic `fixedRoller` via `dice.RollWith`. Use table-driven
  tests, with at least one parse-error case for new features.
- Modifier `String()` output must round-trip through the parser. Add a
  test for it when you add a new modifier.

## Adding a modifier

The lexer doesn't know about modifier keywords. `parseModifier` in
`parser.go` matches the longest known keyword prefix from the
`modifierKeywords` slice and rewrites the IDENT token to drop the
consumed prefix.

Steps:

1. Implement the `modifier` interface (`apply`, `String`, `explain`).
2. Add the keyword(s) to `modifierKeywords` — longer keywords go before
   shorter prefixes (`kh` before `k`, `xo` before `x`, `rr` before `r`).
3. Add a case in `parseModifierByKeyword`.
4. Write parse, eval, and round-trip tests.

## Reporting issues

Bugs and feature requests: open a [GitHub issue][issues].

Security vulnerabilities: see [SECURITY.md](SECURITY.md). Don't open
public issues for security problems.

[issues]: https://github.com/oldschoolgamers/dice/issues
