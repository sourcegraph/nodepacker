package filter

import (
	"fmt"
	"regexp"
	"strings"

	"nodepacker/types"
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer/ebnf"
)

const lexerSpec = `
Whitespace = " " | "\t" | "\n" | "\r" .
Natural = digit { digit } .
NumericOp = ("<" ["="]) | (">" ["="]) .
StringOp = "~=" .
Op = ("!" "=") | "=" .
NumericField = "mem" | "cpu" .
StringField = "name" .
Punct = "(" | ")" | "&" | "|" .
QuotedString = "'" { "\u0000"…"\uffff"-"'" } "'"  .

alpha = "a"…"z" | "A"…"Z" .
digit = "0"…"9" .
`

type stringComparison struct {
	Field string `@StringField`
	Op    string `(@Op | @StringOp)`
	Value string `@QuotedString`
}

func (cf *stringComparison) Pass(r types.Resource) bool {
	switch cf.Op {
	case "=":
		return r.Name == cf.Value
	case "!=":
		return r.Name != cf.Value
	case "~=":
		rgx, err := regexp.Compile(cf.Value)
		if err != nil {
			return false
		}
		return rgx.MatchString(r.Name)
	}
	return true
}

type numericComparison struct {
	Field   string `@NumericField`
	Op      string `(@NumericOp | @Op)`
	Natural int64  `@Natural`
}

func (cf *numericComparison) Pass(r types.Resource) bool {
	if cf.Field != "mem" && cf.Field != "cpu" {
		fmt.Println("failed to parse machine filter: expected mem or cpu")
		return false
	}
	v := r.CPU
	if cf.Field == "mem" {
		v = r.Memory
	}

	switch cf.Op {
	case "=":
		return v == cf.Natural*1000
	case "!=":
		return v != cf.Natural*1000
	case ">":
		return v > cf.Natural*1000
	case ">=":
		return v >= cf.Natural*1000
	case "<":
		return v < cf.Natural*1000
	case "<=":
		return v <= cf.Natural*1000
	}
	fmt.Println("failed to parse machine filter: unknown operator", cf.Op)
	return false
}

type comparison struct {
	NumericComp *numericComparison `@@`
	StringComp  *stringComparison  `| @@`
}

func (cf *comparison) Pass(r types.Resource) bool {
	if cf.NumericComp != nil {
		return cf.NumericComp.Pass(r)
	}
	return cf.StringComp.Pass(r)
}

type leaf struct {
	Comp          *comparison `@@`
	Subexpression *expression `| "(" @@ ")"`
}

func (cf *leaf) Pass(r types.Resource) bool {
	if cf.Comp != nil {
		return cf.Comp.Pass(r)
	}
	return cf.Subexpression.Pass(r)
}

type andFactor struct {
	Factor *leaf `"&" @@`
}

func (cf *andFactor) Pass(r types.Resource) bool {
	return cf.Factor.Pass(r)
}

type term struct {
	Left  *leaf        `@@`
	Right []*andFactor `{ @@ }`
}

func (cf *term) Pass(r types.Resource) bool {
	b := cf.Left.Pass(r)
	if !b {
		return false
	}
	for _, af := range cf.Right {
		ab := af.Pass(r)
		if !ab {
			return false
		}
	}
	return true
}

type orTerm struct {
	Term *term `"|" @@`
}

func (cf *orTerm) Pass(r types.Resource) bool {
	return cf.Term.Pass(r)
}

type expression struct {
	Left  *term     `@@`
	Right []*orTerm `{ @@ }`
}

func (cf *expression) Pass(r types.Resource) bool {
	b := cf.Left.Pass(r)
	if b {
		return true
	}
	for _, of := range cf.Right {
		ob := of.Pass(r)
		if ob {
			return true
		}
	}
	return false
}

func Create(args []string) (types.ResourceFilter, error) {
	if len(args) == 0 {
		return &types.NoopResourceFilter{}, nil
	}

	expr := strings.Join(args, " ")

	return parseMachineFilter(expr)
}

func parseMachineFilter(expr string) (*expression, error) {
	filterLexer, err := ebnf.New(lexerSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse machine filter (lexer): %v", err)
	}
	parser, err := participle.Build(&expression{},
		participle.Lexer(filterLexer), participle.Elide("Whitespace"),
		participle.Unquote("QuotedString"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse machine filter (parser builder): %v", err)
	}

	var rf expression

	err = parser.ParseString(expr, &rf)
	if err != nil {

		return nil, fmt.Errorf("failed to parse machine filter (parse): %v", err)
	}

	return &rf, nil
}
