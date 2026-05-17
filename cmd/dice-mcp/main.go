// Command dice-mcp is an MCP server that exposes the dice library as two
// tools: roll (evaluate a dice expression and return the result) and explain
// (translate a dice expression into plain English).
//
// It speaks MCP over stdio. Register with Claude Code via:
//
//	claude mcp add dice -s user -- /path/to/dice-mcp
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/old-school-gamers/dice"
)

const version = "0.1.0"

// RollInput is the input schema for the roll tool.
type RollInput struct {
	Expression string `json:"expression" jsonschema:"dice expression to roll, e.g. '2d6+3', '4d6kh3', 'adv(1d20)+5 >= 15'"`
}

// ExplainInput is the input schema for the explain tool.
type ExplainInput struct {
	Expression string `json:"expression" jsonschema:"dice expression to describe in plain English"`
}

const rollDescription = `Roll a tabletop RPG dice expression.

Notation:
  XdY              roll X dice with Y sides (e.g. 2d6, 1d20)
  +/-              integer modifiers (e.g. 1d20+5)
  khN, klN         keep highest/lowest N (e.g. 4d6kh3 — D&D 5e ability scores)
  dhN, dlN         drop highest/lowest N
  x[N], xo[N]      explode on max (or >=N); xo = once only
  r[N], rr[N]      reroll on =N (or comparison); rr = recursive
  cs[>=N], cf[<=N] count successes/failures (World of Darkness style)
  adv(expr)        roll twice, take higher (D&D 5e advantage)
  dis(expr)        roll twice, take lower
  >= N, <= N, etc. trailing target number; sets hit/miss without changing total

Examples:
  "2d6+3"               basic damage
  "4d6kh3"              ability score
  "adv(1d20)+5 >= 15"   attack roll with advantage against AC 15
  "5d10cs>=8"           World of Darkness pool, count 8s and up`

const explainDescription = `Translate a dice expression into a plain-English description.

Useful to confirm interpretation before rolling, or to render help text. Does not roll.`

// handlers carries the dependencies needed by tool handlers. A nil Roller
// causes the library to fall back to crypto/rand; tests inject a deterministic
// roller instead.
type handlers struct {
	roller dice.Roller
}

func main() {
	log.SetOutput(os.Stderr)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "dice-mcp",
		Version: version,
	}, nil)

	h := &handlers{}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "roll",
		Description: rollDescription,
	}, h.handleRoll)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "explain",
		Description: explainDescription,
	}, h.handleExplain)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("dice-mcp: %v", err)
	}
}

func (h *handlers) handleRoll(_ context.Context, _ *mcp.CallToolRequest, in RollInput) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(in.Expression) == "" {
		return textResult("Error: expression is required", true), nil, nil
	}
	res, err := dice.RollWith(in.Expression, h.roller)
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err), true), nil, nil
	}
	return textResult(formatRollText(res), false), buildRollOutput(res), nil
}

func (h *handlers) handleExplain(_ context.Context, _ *mcp.CallToolRequest, in ExplainInput) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(in.Expression) == "" {
		return textResult("Error: expression is required", true), nil, nil
	}
	e, err := dice.Parse(in.Expression)
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err), true), nil, nil
	}
	desc := e.Explain()
	return textResult(desc, false), ExplainOutput{Expression: in.Expression, Description: desc}, nil
}

func textResult(text string, isError bool) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
		IsError: isError,
	}
}
