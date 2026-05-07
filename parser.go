package dice

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// node is the internal AST. eval contributes a value to the running Total
// and may append Groups to res for rolled dice. String returns canonical
// notation; explain returns prose description for Expression.Explain.
type node interface {
	eval(r Roller, res *Result) (int, error)
	String() string
	explain() string
}

type numNode struct{ value int }

func (n *numNode) String() string { return strconv.Itoa(n.value) }

type binOpNode struct {
	op    rune // '+' or '-'
	left  node
	right node
}

func (n *binOpNode) String() string {
	// Suppress the synthetic "0-x" form produced by unary minus.
	if l, ok := n.left.(*numNode); ok && l.value == 0 && n.op == '-' {
		return "-" + n.right.String()
	}
	return fmt.Sprintf("%s%c%s", n.left.String(), n.op, n.right.String())
}

type diceNode struct {
	count    int
	sides    int
	notation string // canonicalized source for the Group label
	mods     []modifier
}

func (n *diceNode) String() string { return n.notation }

type advNode struct {
	inner        node
	disadvantage bool
}

func (n *advNode) String() string {
	keyword := "adv"
	if n.disadvantage {
		keyword = "dis"
	}
	return fmt.Sprintf("%s(%s)", keyword, n.inner.String())
}

// --- Lexer ---

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokNum
	tokIdent // run of letters; meaning is parser-context-dependent
	tokPlus
	tokMinus
	tokLParen
	tokRParen
	tokGE
	tokLE
	tokGT
	tokLT
	tokEQ
)

type token struct {
	kind tokenKind
	text string
	pos  int
}

func lex(src string) ([]token, error) {
	src = strings.ToLower(src)
	var toks []token
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == ' ' || c == '\t':
			i++
		case c >= '0' && c <= '9':
			start := i
			for i < len(src) && src[i] >= '0' && src[i] <= '9' {
				i++
			}
			toks = append(toks, token{tokNum, src[start:i], start})
		case unicode.IsLetter(rune(c)):
			start := i
			for i < len(src) && unicode.IsLetter(rune(src[i])) {
				i++
			}
			toks = append(toks, token{tokIdent, src[start:i], start})
		case c == '+':
			toks = append(toks, token{tokPlus, "+", i})
			i++
		case c == '-':
			toks = append(toks, token{tokMinus, "-", i})
			i++
		case c == '(':
			toks = append(toks, token{tokLParen, "(", i})
			i++
		case c == ')':
			toks = append(toks, token{tokRParen, ")", i})
			i++
		case c == '>':
			if i+1 < len(src) && src[i+1] == '=' {
				toks = append(toks, token{tokGE, ">=", i})
				i += 2
			} else {
				toks = append(toks, token{tokGT, ">", i})
				i++
			}
		case c == '<':
			if i+1 < len(src) && src[i+1] == '=' {
				toks = append(toks, token{tokLE, "<=", i})
				i += 2
			} else {
				toks = append(toks, token{tokLT, "<", i})
				i++
			}
		case c == '=':
			toks = append(toks, token{tokEQ, "=", i})
			i++
		default:
			return nil, fmt.Errorf("dice: unexpected character %q at position %d", c, i)
		}
	}
	return toks, nil
}

// --- Parser ---

type parser struct {
	toks []token
	i    int
	src  string
}

func parse(src string) (node, *Comparison, error) {
	if strings.TrimSpace(src) == "" {
		return nil, nil, fmt.Errorf("dice: empty expression")
	}
	toks, err := lex(src)
	if err != nil {
		return nil, nil, err
	}
	p := &parser{toks: toks, src: src}
	root, err := p.parseExpr()
	if err != nil {
		return nil, nil, err
	}
	// Optional trailing target number: any comparison operator followed by an integer.
	var target *Comparison
	if isCompareTok(p.peek().kind) {
		c, ok, err := p.parseComparison(OpGE, false)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			target = &c
		}
	}
	if p.i < len(p.toks) {
		t := p.toks[p.i]
		return nil, nil, fmt.Errorf("dice: unexpected %q at position %d", t.text, t.pos)
	}
	return root, target, nil
}

func isCompareTok(k tokenKind) bool {
	switch k {
	case tokGE, tokLE, tokGT, tokLT, tokEQ:
		return true
	}
	return false
}

func (p *parser) peek() token {
	if p.i >= len(p.toks) {
		return token{kind: tokEOF, pos: len(p.src)}
	}
	return p.toks[p.i]
}

func (p *parser) advance() token {
	t := p.peek()
	if t.kind != tokEOF {
		p.i++
	}
	return t
}

// expr := term (('+'|'-') term)*
func (p *parser) parseExpr() (node, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		if t.kind != tokPlus && t.kind != tokMinus {
			return left, nil
		}
		p.advance()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		op := '+'
		if t.kind == tokMinus {
			op = '-'
		}
		left = &binOpNode{op: op, left: left, right: right}
	}
}

// term := number | dice | '-' term | '(' expr ')' | 'd' number ...
func (p *parser) parseTerm() (node, error) {
	t := p.peek()
	switch t.kind {
	case tokLParen:
		p.advance()
		n, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().kind != tokRParen {
			return nil, fmt.Errorf("dice: expected ')' at position %d", p.peek().pos)
		}
		p.advance()
		return n, nil
	case tokMinus:
		p.advance()
		inner, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		return &binOpNode{op: '-', left: &numNode{0}, right: inner}, nil
	case tokNum:
		return p.parseNumOrDice()
	case tokIdent:
		// 'd' alone means implicit count of 1 (e.g. "d20").
		if t.text == "d" {
			return p.parseDice(1)
		}
		// adv(expr) / dis(expr): roll twice, take higher / lower.
		if t.text == "adv" || t.text == "dis" {
			return p.parseAdvOrDis(t.text == "dis")
		}
		return nil, fmt.Errorf("dice: unexpected identifier %q at position %d", t.text, t.pos)
	default:
		return nil, fmt.Errorf("dice: expected expression at position %d", t.pos)
	}
}

func (p *parser) parseNumOrDice() (node, error) {
	numTok := p.advance()
	n, err := strconv.Atoi(numTok.text)
	if err != nil {
		return nil, fmt.Errorf("dice: invalid number %q at position %d", numTok.text, numTok.pos)
	}
	if p.peek().kind == tokIdent && p.peek().text == "d" {
		return p.parseDice(n)
	}
	return &numNode{value: n}, nil
}

func (p *parser) parseAdvOrDis(disadvantage bool) (node, error) {
	keyword := p.advance().text
	if p.peek().kind != tokLParen {
		return nil, fmt.Errorf("dice: expected '(' after %s at position %d", keyword, p.peek().pos)
	}
	p.advance()
	inner, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.peek().kind != tokRParen {
		return nil, fmt.Errorf("dice: expected ')' to close %s at position %d", keyword, p.peek().pos)
	}
	p.advance()
	return &advNode{inner: inner, disadvantage: disadvantage}, nil
}

// parseDice consumes the 'd' and the sides number, then any trailing modifiers.
func (p *parser) parseDice(count int) (node, error) {
	if count < 0 {
		return nil, fmt.Errorf("dice: count must be >= 0, got %d", count)
	}
	p.advance() // consume 'd'
	if p.peek().kind != tokNum {
		return nil, fmt.Errorf("dice: expected die size after 'd' at position %d", p.peek().pos)
	}
	sidesTok := p.advance()
	sides, err := strconv.Atoi(sidesTok.text)
	if err != nil {
		return nil, fmt.Errorf("dice: invalid die size %q at position %d", sidesTok.text, sidesTok.pos)
	}
	if sides < 1 {
		return nil, fmt.Errorf("dice: die size must be >= 1, got %d", sides)
	}
	d := &diceNode{count: count, sides: sides}
	for {
		mod, ok, err := p.parseModifier()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		d.mods = append(d.mods, mod)
	}
	d.notation = fmt.Sprintf("%dd%d", count, sides)
	for _, m := range d.mods {
		d.notation += m.String()
	}
	return d, nil
}
