package fp

import (
	"github.com/kr/pretty"
	"github.com/si3nloong/go-rsql/lex"
)

// Lexer types
const (
	TypeName = (1 + lex.TypeEOF) + iota // continue where lex left off
	TypeOperator
	TypeValuesStart
	TypeValuesEnd
	TypeValue
	TypeAnd
	TypeOr
	TypeGroupStart
	TypeGroupEnd
)

var (
	SINGLE_QUOTE = `'`
	DOUBLE_QUOTE = `"`
	RESERVED     = `"'();,=!~<>`
	RESERVED_VAL = RESERVED + " "
)

func Parse(input string) (*Node, error) {
	lexer := lex.Lex("rsql", input, lexName)

	rootNode := newNode(NodeGroup)
	var comp *Comparison
	var node *Node

	node = rootNode // init

outer:
	for {
		// println("> next token")
		tok := lexer.NextToken()
		pretty.Println(tok)

		switch tok.Type {
		case TypeName:
			comp = newComparison(tok.Value)
			// fmt.Println("name", tok.Value)
		case TypeOperator:
			comp.Operator = operators[tok.Value]
			// fmt.Println("op", tok.Value)
		case TypeValue:
			comp.Arguments = append(comp.Arguments, tok.Value)
			// fmt.Println("value", tok.Value)

		case TypeValuesEnd:
			node.Nodes = append(node.Nodes, newCompNode(comp))

		case TypeOr:
			node.Nodes = append(node.Nodes, newLogicNode(Or))

		case TypeAnd:
			node.Nodes = append(node.Nodes, newLogicNode(And))

		case lex.TypeError, lex.TypeEOF:
			break outer
		}
	}

	return rootNode, nil
}

func lexName(l *lex.Lexer) lex.StateFn {
	if l.Peek() == lex.EOF {
		return nil
	}
	l.AcceptButRun("=<>!")
	l.Emit(TypeName)
	return lexOperator
}

func lexOperator(l *lex.Lexer) lex.StateFn {
	if l.Accept("=") {
		// handle =..= operators
		for {
			r := l.Next()
			if r == '=' {
				l.Emit(TypeOperator)
				return lexValues
			}
		}
	}

	// handle mathematic operators
	// > >= < <= !=
	l.Accept("><!")
	l.Accept("=")

	l.Emit(TypeOperator)
	return lexValues
}

// lexValues looks for an array of values grouped with parens
// e.g., (a,'b')
func lexValues(l *lex.Lexer) lex.StateFn {
	if l.Accept("(") {
		l.Ignore()
		l.Emit(TypeValuesStart)
		for {
			lexValue(l)
			if l.Accept(",") {
				l.Ignore()
				continue
			} else if l.Accept(")") {
				break
			}
		}

		// closed value group
		l.Ignore()
		l.Emit(TypeValuesEnd)
		return lexLogic
	}

	// single value
	l.Emit(TypeValuesStart)
	lexValue(l)
	l.Emit(TypeValuesEnd)
	return lexLogic
}

// lexValue reads a single value, quoted or unquoted
func lexValue(l *lex.Lexer) {
	if lexQuotedValue(l, SINGLE_QUOTE) || lexQuotedValue(l, DOUBLE_QUOTE) {
		return
	}
	// consume bare value
	l.AcceptButRun(RESERVED_VAL)
	l.Emit(TypeValue)
}

// lexQuotedValue and return true if processed. False if nothing was done.
func lexQuotedValue(l *lex.Lexer, quoteChar string) bool {
	if l.Accept(quoteChar) {
		l.Ignore()
		for {
			l.AcceptButRun(quoteChar)
			l.Dec(1)
			if l.Accept(`\`) {
				l.Next() // consume the " also
			} else {
				// no escape char, consume whatever was there
				l.Next()
				l.Emit(TypeValue)
				l.Accept(quoteChar) // consume the "
				l.Ignore()
				break
			}
		}
		return true
	}
	return false
}

func eatSpaces(l *lex.Lexer) {
	l.AcceptRun(" ")
	l.Ignore()
}

func lexLogic(l *lex.Lexer) lex.StateFn {
	if l.Peek() == lex.EOF {
		return nil
	}

	eatSpaces(l)
	if l.Consume("and") || l.Consume(";") {
		l.Emit(TypeAnd)
		eatSpaces(l)
		return lexName

	} else if l.Consume(",") || l.Consume("or") {
		l.Emit(TypeOr)
		eatSpaces(l)
		return lexName
	}
	return nil
}

func isValueRune(r rune) bool {
	return lex.IsAlphaNumeric(r)
}
