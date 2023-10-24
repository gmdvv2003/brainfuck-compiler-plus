package parser

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/gmdvv2003/brainfuck-compiler-plus/lexer"
)

type Production int

const (
	NODE_INTEGER_LITERAL Production = iota

	NODE_NEXT_CELL
	NODE_PREVIOUS_CELL

	NODE_INCREMENT_CELL
	NODE_DECREMENT_CELL

	NODE_OUTPUT_CELL
	NODE_INPUT_CELL

	NODE_WHILE
)

type AST struct {
	Nodes []interface{}
}

type Node struct {
	Repeat   int
	NodeType Production
}

type IntegerLiteral struct {
	Value int
	Node
}

type NextCell struct{ Node }
type PreviousCell struct{ Node }

type IncrementCell struct{ Node }
type DecrementCell struct{ Node }

type OutputCell struct{ Node }
type InputCell struct{ Node }

type NodeWhile struct {
	Tree *AST
	Node
}

func nodeIncrementer(nodeType reflect.Type) func(interface{}) {
	return func(node interface{}) {
		repeatValue := reflect.ValueOf(node).Elem().FieldByName("Node").FieldByName("Repeat")
		repeatValue.SetInt(repeatValue.Int() + 1)
	}
}

var (
	nodesTokensTypes = map[lexer.Token]reflect.Type{
		lexer.NEXT_CELL:     reflect.TypeOf(&NextCell{}),
		lexer.PREVIOUS_CELL: reflect.TypeOf(&PreviousCell{}),

		lexer.INCREMENT_CELL: reflect.TypeOf(&IncrementCell{}),
		lexer.DECREMENT_CELL: reflect.TypeOf(&DecrementCell{}),
	}

	repeatableTokens = map[reflect.Type]func(interface{}){
		reflect.TypeOf(&NextCell{}):     nodeIncrementer(reflect.TypeOf(&NextCell{})),
		reflect.TypeOf(&PreviousCell{}): nodeIncrementer(reflect.TypeOf(&PreviousCell{})),

		reflect.TypeOf(&IncrementCell{}): nodeIncrementer(reflect.TypeOf(&IncrementCell{})),
		reflect.TypeOf(&DecrementCell{}): nodeIncrementer(reflect.TypeOf(&DecrementCell{})),
	}
)

// Parse a sequence of tokens of the current block and produce an abstract tree
func Parse(sourceLexer *lexer.Lexer, blockType reflect.Type) (*AST, lexer.Token, error) {
	tree := &AST{}

	for {
		position, token, symbol := sourceLexer.Lex()
		if token == lexer.EOF {
			break
		}

		if *sourceLexer.Debug {
			fmt.Printf("token: %s, symbol: %s, position: %d:%d\n", token, symbol, position.Line+1, position.Column)
		}

		// Check wheter the last node is of the same type as the current token, if so, increment the repeat count if possible
		if len(tree.Nodes) > 0 {
			if tokenType, tokenTypeExists := nodesTokensTypes[token]; tokenTypeExists {
				if tokenIncrementer, tokenIncrementerExists := repeatableTokens[tokenType]; tokenIncrementerExists {
					// Make sure their reflection types is equal
					if reflect.TypeOf(tree.Nodes[len(tree.Nodes)-1]) == tokenType {
						tokenIncrementer(tree.Nodes[len(tree.Nodes)-1])
						continue
					}
				}
			}
		}

		switch token {
		case lexer.BRACKET_OPEN:
			// Parse the just entered block tokens and append it as a while node
			if subTree, lastSubTreeToken, ok := Parse(sourceLexer, reflect.TypeOf(&NodeWhile{})); ok == nil {
				// Assure that the last node was closed properly
				if lastSubTreeToken != lexer.BRACKET_CLOSE {
					return nil, -1, fmt.Errorf("unmatched [ at %d:%d", position.Line+1, position.Column)
				}

				tree.Nodes = append(tree.Nodes, &NodeWhile{Tree: subTree, Node: Node{NodeType: NODE_WHILE}})
			} else {
				return nil, -1, ok
			}
		case lexer.BRACKET_CLOSE:
			// Assure that the current block was opened properly
			if blockType != nil && blockType == reflect.TypeOf(&NodeWhile{}) {
				return tree, lexer.BRACKET_CLOSE, nil
			} else {
				return nil, -1, fmt.Errorf("encountered unmatched ] at %d:%d", position.Line+1, position.Column)
			}
		case lexer.NUMBER:
			// Attempts to parse the received symbol as an integer
			if number, ok := strconv.Atoi(symbol); ok == nil {
				tree.Nodes = append(tree.Nodes, &IntegerLiteral{Value: number, Node: Node{NodeType: NODE_INTEGER_LITERAL}})
			} else {
				return nil, -1, fmt.Errorf("encountered invalid number %s at %d:%d", symbol, position.Line+1, position.Column)
			}
		case lexer.NEXT_CELL:
			tree.Nodes = append(tree.Nodes, &NextCell{Node: Node{NodeType: NODE_NEXT_CELL}})
		case lexer.PREVIOUS_CELL:
			tree.Nodes = append(tree.Nodes, &PreviousCell{Node: Node{NodeType: NODE_PREVIOUS_CELL}})
		case lexer.INCREMENT_CELL:
			tree.Nodes = append(tree.Nodes, &IncrementCell{Node: Node{NodeType: NODE_INCREMENT_CELL}})
		case lexer.DECREMENT_CELL:
			tree.Nodes = append(tree.Nodes, &DecrementCell{Node: Node{NodeType: NODE_DECREMENT_CELL}})
		case lexer.INPUT_CELL:
			tree.Nodes = append(tree.Nodes, &InputCell{Node: Node{NodeType: NODE_INPUT_CELL}})
		case lexer.OUTPUT_CELL:
			tree.Nodes = append(tree.Nodes, &OutputCell{Node: Node{NodeType: NODE_OUTPUT_CELL}})
		default:
			return nil, -1, fmt.Errorf("encountered unhandled token %s at %d:%d", token, position.Line+1, position.Column)
		}
	}

	return tree, -1, nil
}
