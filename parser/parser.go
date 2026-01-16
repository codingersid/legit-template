package parser

import (
	"fmt"
	"strings"

	"github.com/codingersid/legit-template/lexer"
)

// NodeType represents the type of AST node
type NodeType int

const (
	NODE_ROOT NodeType = iota
	NODE_TEXT
	NODE_ECHO_ESCAPED
	NODE_ECHO_RAW
	NODE_COMMENT
	NODE_DIRECTIVE
	NODE_IF
	NODE_ELSEIF
	NODE_ELSE
	NODE_UNLESS
	NODE_SWITCH
	NODE_CASE
	NODE_DEFAULT
	NODE_FOR
	NODE_FOREACH
	NODE_FORELSE
	NODE_WHILE
	NODE_SECTION
	NODE_YIELD
	NODE_EXTENDS
	NODE_INCLUDE
	NODE_EACH
	NODE_PUSH
	NODE_PREPEND
	NODE_STACK
	NODE_COMPONENT
	NODE_SLOT
	NODE_VERBATIM
	NODE_PHP
	NODE_BREAK
	NODE_CONTINUE
	NODE_EMPTY
	NODE_ISSET
	NODE_AUTH
	NODE_GUEST
	NODE_ENV
	NODE_PRODUCTION
	NODE_ERROR
	NODE_ONCE
	NODE_PARENT
)

// Node represents an AST node
type Node interface {
	Type() NodeType
	Position() lexer.Position
}

// BaseNode contains common node fields
type BaseNode struct {
	NodeType NodeType
	Pos      lexer.Position
}

func (n *BaseNode) Type() NodeType          { return n.NodeType }
func (n *BaseNode) Position() lexer.Position { return n.Pos }

// RootNode is the root of the AST
type RootNode struct {
	BaseNode
	Children []Node
}

// TextNode represents plain text
type TextNode struct {
	BaseNode
	Content string
}

// EchoNode represents {{ }} or {!! !!}
type EchoNode struct {
	BaseNode
	Expression string
	Escaped    bool
}

// CommentNode represents {{-- --}}
type CommentNode struct {
	BaseNode
	Content string
}

// DirectiveNode represents a simple directive without block
type DirectiveNode struct {
	BaseNode
	Name string
	Args string
}

// BlockNode represents a directive with children (if, foreach, etc.)
type BlockNode struct {
	BaseNode
	Name     string
	Args     string
	Children []Node
}

// IfNode represents @if...@elseif...@else...@endif
type IfNode struct {
	BaseNode
	Condition  string
	Children   []Node
	ElseIfs    []*ElseIfNode
	Else       *ElseNode
}

// ElseIfNode represents @elseif
type ElseIfNode struct {
	BaseNode
	Condition string
	Children  []Node
}

// ElseNode represents @else
type ElseNode struct {
	BaseNode
	Children []Node
}

// UnlessNode represents @unless...@endunless
type UnlessNode struct {
	BaseNode
	Condition string
	Children  []Node
}

// SwitchNode represents @switch...@endswitch
type SwitchNode struct {
	BaseNode
	Expression string
	Cases      []*CaseNode
	Default    *DefaultNode
}

// CaseNode represents @case
type CaseNode struct {
	BaseNode
	Value    string
	Children []Node
}

// DefaultNode represents @default in switch
type DefaultNode struct {
	BaseNode
	Children []Node
}

// ForNode represents @for...@endfor
type ForNode struct {
	BaseNode
	Init      string
	Condition string
	Post      string
	Children  []Node
}

// ForeachNode represents @foreach...@endforeach
type ForeachNode struct {
	BaseNode
	Items    string
	Key      string
	Value    string
	Children []Node
}

// ForelseNode represents @forelse...@empty...@endforelse
type ForelseNode struct {
	BaseNode
	Items     string
	Key       string
	Value     string
	Children  []Node
	Empty     []Node
}

// WhileNode represents @while...@endwhile
type WhileNode struct {
	BaseNode
	Condition string
	Children  []Node
}

// SectionNode represents @section...@endsection or @section...@show
type SectionNode struct {
	BaseNode
	Name     string
	Content  string   // For inline @section('name', 'content')
	Children []Node
	Show     bool     // If @show is used instead of @endsection
}

// YieldNode represents @yield
type YieldNode struct {
	BaseNode
	Name    string
	Default string
}

// ExtendsNode represents @extends
type ExtendsNode struct {
	BaseNode
	Template string
}

// IncludeNode represents @include, @includeIf, @includeWhen, @includeUnless, @includeFirst
type IncludeNode struct {
	BaseNode
	Variant   string // include, includeIf, includeWhen, includeUnless, includeFirst
	Template  string
	Data      string
	Condition string // For includeWhen/includeUnless
}

// EachNode represents @each
type EachNode struct {
	BaseNode
	Template  string
	Items     string
	ItemVar   string
	EmptyView string
}

// PushNode represents @push...@endpush
type PushNode struct {
	BaseNode
	Stack    string
	Children []Node
	Once     bool // For @pushOnce
}

// PrependNode represents @prepend...@endprepend
type PrependNode struct {
	BaseNode
	Stack    string
	Children []Node
}

// StackNode represents @stack
type StackNode struct {
	BaseNode
	Name string
}

// ComponentNode represents @component...@endcomponent
type ComponentNode struct {
	BaseNode
	Name     string
	Data     string
	Children []Node
	Slots    map[string]*SlotNode
}

// SlotNode represents @slot...@endslot
type SlotNode struct {
	BaseNode
	Name     string
	Children []Node
}

// VerbatimNode represents @verbatim...@endverbatim
type VerbatimNode struct {
	BaseNode
	Content string
}

// PhpNode represents @php...@endphp
type PhpNode struct {
	BaseNode
	Code string
}

// BreakNode represents @break
type BreakNode struct {
	BaseNode
	Condition string
}

// ContinueNode represents @continue
type ContinueNode struct {
	BaseNode
	Condition string
}

// IssetNode represents @isset...@endisset
type IssetNode struct {
	BaseNode
	Variable string
	Children []Node
}

// EmptyCheckNode represents @empty...@endempty
type EmptyCheckNode struct {
	BaseNode
	Variable string
	Children []Node
}

// AuthNode represents @auth...@endauth
type AuthNode struct {
	BaseNode
	Guard    string
	Children []Node
}

// GuestNode represents @guest...@endguest
type GuestNode struct {
	BaseNode
	Guard    string
	Children []Node
}

// EnvNode represents @env...@endenv
type EnvNode struct {
	BaseNode
	Environments []string
	Children     []Node
}

// ProductionNode represents @production...@endproduction
type ProductionNode struct {
	BaseNode
	Children []Node
}

// ErrorNode represents @error...@enderror
type ErrorNode struct {
	BaseNode
	Field    string
	Children []Node
}

// OnceNode represents @once...@endonce
type OnceNode struct {
	BaseNode
	Children []Node
}

// ParentNode represents @parent
type ParentNode struct {
	BaseNode
}

// Parser builds AST from tokens
type Parser struct {
	tokens  []lexer.Token
	pos     int
	current lexer.Token
}

// New creates a new Parser
func New(tokens []lexer.Token) *Parser {
	p := &Parser{
		tokens: tokens,
		pos:    0,
	}
	if len(tokens) > 0 {
		p.current = tokens[0]
	}
	return p
}

// Parse parses tokens into AST
func (p *Parser) Parse() (*RootNode, error) {
	root := &RootNode{
		BaseNode: BaseNode{NodeType: NODE_ROOT},
		Children: make([]Node, 0),
	}

	for !p.isAtEnd() {
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			root.Children = append(root.Children, node)
		}
	}

	return root, nil
}

// parseNode parses a single node
func (p *Parser) parseNode() (Node, error) {
	token := p.current

	switch token.Type {
	case lexer.TOKEN_TEXT:
		p.advance()
		return &TextNode{
			BaseNode: BaseNode{NodeType: NODE_TEXT, Pos: token.Position},
			Content:  token.Value,
		}, nil

	case lexer.TOKEN_ECHO_ESCAPED:
		p.advance()
		return &EchoNode{
			BaseNode:   BaseNode{NodeType: NODE_ECHO_ESCAPED, Pos: token.Position},
			Expression: token.Value,
			Escaped:    true,
		}, nil

	case lexer.TOKEN_ECHO_RAW:
		p.advance()
		return &EchoNode{
			BaseNode:   BaseNode{NodeType: NODE_ECHO_RAW, Pos: token.Position},
			Expression: token.Value,
			Escaped:    false,
		}, nil

	case lexer.TOKEN_COMMENT:
		p.advance()
		return &CommentNode{
			BaseNode: BaseNode{NodeType: NODE_COMMENT, Pos: token.Position},
			Content:  token.Value,
		}, nil

	case lexer.TOKEN_DIRECTIVE, lexer.TOKEN_DIRECTIVE_ARGS:
		return p.parseDirective()

	case lexer.TOKEN_VERBATIM_START:
		return p.parseVerbatim()

	case lexer.TOKEN_EOF:
		return nil, nil

	default:
		p.advance()
		return nil, nil
	}
}

// parseDirective parses a directive
func (p *Parser) parseDirective() (Node, error) {
	token := p.current
	name := token.Value
	args := token.Args
	p.advance()

	switch name {
	case "if":
		return p.parseIf(token.Position, args)
	case "unless":
		return p.parseUnless(token.Position, args)
	case "switch":
		return p.parseSwitch(token.Position, args)
	case "for":
		return p.parseFor(token.Position, args)
	case "foreach":
		return p.parseForeach(token.Position, args)
	case "forelse":
		return p.parseForelse(token.Position, args)
	case "while":
		return p.parseWhile(token.Position, args)
	case "section":
		return p.parseSection(token.Position, args)
	case "yield":
		return p.parseYield(token.Position, args)
	case "extends":
		return &ExtendsNode{
			BaseNode: BaseNode{NodeType: NODE_EXTENDS, Pos: token.Position},
			Template: trimQuotes(args),
		}, nil
	case "include", "includeIf", "includeWhen", "includeUnless", "includeFirst":
		return p.parseInclude(token.Position, name, args)
	case "each":
		return p.parseEach(token.Position, args)
	case "push":
		return p.parsePush(token.Position, args, false)
	case "pushOnce":
		return p.parsePush(token.Position, args, true)
	case "prepend":
		return p.parsePrepend(token.Position, args)
	case "stack":
		return &StackNode{
			BaseNode: BaseNode{NodeType: NODE_STACK, Pos: token.Position},
			Name:     trimQuotes(args),
		}, nil
	case "component":
		return p.parseComponent(token.Position, args)
	case "php":
		return p.parsePhp(token.Position)
	case "isset":
		return p.parseIsset(token.Position, args)
	case "empty":
		return p.parseEmptyCheck(token.Position, args)
	case "auth":
		return p.parseAuth(token.Position, args)
	case "guest":
		return p.parseGuest(token.Position, args)
	case "env":
		return p.parseEnv(token.Position, args)
	case "production":
		return p.parseProduction(token.Position)
	case "error":
		return p.parseError(token.Position, args)
	case "once":
		return p.parseOnce(token.Position)
	case "break":
		return &BreakNode{
			BaseNode:  BaseNode{NodeType: NODE_BREAK, Pos: token.Position},
			Condition: args,
		}, nil
	case "continue":
		return &ContinueNode{
			BaseNode:  BaseNode{NodeType: NODE_CONTINUE, Pos: token.Position},
			Condition: args,
		}, nil
	case "parent":
		return &ParentNode{
			BaseNode: BaseNode{NodeType: NODE_PARENT, Pos: token.Position},
		}, nil
	case "csrf", "method", "json", "class", "style", "checked", "selected", "disabled", "readonly", "required", "old":
		return &DirectiveNode{
			BaseNode: BaseNode{NodeType: NODE_DIRECTIVE, Pos: token.Position},
			Name:     name,
			Args:     args,
		}, nil
	default:
		// Unknown directive - treat as simple directive
		return &DirectiveNode{
			BaseNode: BaseNode{NodeType: NODE_DIRECTIVE, Pos: token.Position},
			Name:     name,
			Args:     args,
		}, nil
	}
}

// parseIf parses @if...@elseif...@else...@endif
func (p *Parser) parseIf(pos lexer.Position, condition string) (*IfNode, error) {
	node := &IfNode{
		BaseNode:  BaseNode{NodeType: NODE_IF, Pos: pos},
		Condition: condition,
		Children:  make([]Node, 0),
		ElseIfs:   make([]*ElseIfNode, 0),
	}

	for !p.isAtEnd() {
		if p.isDirective("elseif") {
			elseifToken := p.current
			p.advance()
			elseifNode := &ElseIfNode{
				BaseNode:  BaseNode{NodeType: NODE_ELSEIF, Pos: elseifToken.Position},
				Condition: elseifToken.Args,
				Children:  make([]Node, 0),
			}

			for !p.isAtEnd() && !p.isDirective("elseif") && !p.isDirective("else") && !p.isDirective("endif") {
				child, err := p.parseNode()
				if err != nil {
					return nil, err
				}
				if child != nil {
					elseifNode.Children = append(elseifNode.Children, child)
				}
			}
			node.ElseIfs = append(node.ElseIfs, elseifNode)
			continue
		}

		if p.isDirective("else") {
			elseToken := p.current
			p.advance()
			node.Else = &ElseNode{
				BaseNode: BaseNode{NodeType: NODE_ELSE, Pos: elseToken.Position},
				Children: make([]Node, 0),
			}

			for !p.isAtEnd() && !p.isDirective("endif") {
				child, err := p.parseNode()
				if err != nil {
					return nil, err
				}
				if child != nil {
					node.Else.Children = append(node.Else.Children, child)
				}
			}
			continue
		}

		if p.isDirective("endif") {
			p.advance()
			break
		}

		// Before any elseif/else - add to main children
		if len(node.ElseIfs) == 0 && node.Else == nil {
			child, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			if child != nil {
				node.Children = append(node.Children, child)
			}
		}
	}

	return node, nil
}

// parseUnless parses @unless...@endunless
func (p *Parser) parseUnless(pos lexer.Position, condition string) (*UnlessNode, error) {
	node := &UnlessNode{
		BaseNode:  BaseNode{NodeType: NODE_UNLESS, Pos: pos},
		Condition: condition,
		Children:  make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("endunless") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endunless") {
		p.advance()
	}

	return node, nil
}

// parseSwitch parses @switch...@endswitch
func (p *Parser) parseSwitch(pos lexer.Position, expression string) (*SwitchNode, error) {
	node := &SwitchNode{
		BaseNode:   BaseNode{NodeType: NODE_SWITCH, Pos: pos},
		Expression: expression,
		Cases:      make([]*CaseNode, 0),
	}

	var currentCase *CaseNode

	for !p.isAtEnd() && !p.isDirective("endswitch") {
		if p.isDirective("case") {
			if currentCase != nil {
				node.Cases = append(node.Cases, currentCase)
			}
			caseToken := p.current
			p.advance()
			currentCase = &CaseNode{
				BaseNode: BaseNode{NodeType: NODE_CASE, Pos: caseToken.Position},
				Value:    caseToken.Args,
				Children: make([]Node, 0),
			}
			continue
		}

		if p.isDirective("default") {
			if currentCase != nil {
				node.Cases = append(node.Cases, currentCase)
				currentCase = nil
			}
			defaultToken := p.current
			p.advance()
			node.Default = &DefaultNode{
				BaseNode: BaseNode{NodeType: NODE_DEFAULT, Pos: defaultToken.Position},
				Children: make([]Node, 0),
			}
			continue
		}

		if p.isDirective("break") {
			p.advance()
			continue
		}

		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			if node.Default != nil {
				node.Default.Children = append(node.Default.Children, child)
			} else if currentCase != nil {
				currentCase.Children = append(currentCase.Children, child)
			}
		}
	}

	if currentCase != nil {
		node.Cases = append(node.Cases, currentCase)
	}

	if p.isDirective("endswitch") {
		p.advance()
	}

	return node, nil
}

// parseFor parses @for...@endfor
func (p *Parser) parseFor(pos lexer.Position, args string) (*ForNode, error) {
	// Parse for(init; condition; post) format
	parts := strings.SplitN(args, ";", 3)
	node := &ForNode{
		BaseNode: BaseNode{NodeType: NODE_FOR, Pos: pos},
		Children: make([]Node, 0),
	}

	if len(parts) >= 1 {
		node.Init = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		node.Condition = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 {
		node.Post = strings.TrimSpace(parts[2])
	}

	for !p.isAtEnd() && !p.isDirective("endfor") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endfor") {
		p.advance()
	}

	return node, nil
}

// parseForeach parses @foreach...@endforeach
func (p *Parser) parseForeach(pos lexer.Position, args string) (*ForeachNode, error) {
	node := &ForeachNode{
		BaseNode: BaseNode{NodeType: NODE_FOREACH, Pos: pos},
		Children: make([]Node, 0),
	}

	// Parse $items as $key => $value or $items as $value
	p.parseForeachArgs(args, node)

	for !p.isAtEnd() && !p.isDirective("endforeach") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endforeach") {
		p.advance()
	}

	return node, nil
}

// parseForeachArgs parses foreach arguments
func (p *Parser) parseForeachArgs(args string, node *ForeachNode) {
	// $items as $key => $value
	// $items as $value
	parts := strings.SplitN(args, " as ", 2)
	if len(parts) >= 1 {
		node.Items = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		valuePart := strings.TrimSpace(parts[1])
		if strings.Contains(valuePart, "=>") {
			kvParts := strings.SplitN(valuePart, "=>", 2)
			node.Key = strings.TrimSpace(kvParts[0])
			node.Value = strings.TrimSpace(kvParts[1])
		} else {
			node.Value = valuePart
		}
	}
}

// parseForelse parses @forelse...@empty...@endforelse
func (p *Parser) parseForelse(pos lexer.Position, args string) (*ForelseNode, error) {
	node := &ForelseNode{
		BaseNode: BaseNode{NodeType: NODE_FORELSE, Pos: pos},
		Children: make([]Node, 0),
		Empty:    make([]Node, 0),
	}

	// Parse same as foreach
	parts := strings.SplitN(args, " as ", 2)
	if len(parts) >= 1 {
		node.Items = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		valuePart := strings.TrimSpace(parts[1])
		if strings.Contains(valuePart, "=>") {
			kvParts := strings.SplitN(valuePart, "=>", 2)
			node.Key = strings.TrimSpace(kvParts[0])
			node.Value = strings.TrimSpace(kvParts[1])
		} else {
			node.Value = valuePart
		}
	}

	inEmpty := false
	for !p.isAtEnd() && !p.isDirective("endforelse") {
		if p.isDirective("empty") {
			p.advance()
			inEmpty = true
			continue
		}

		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			if inEmpty {
				node.Empty = append(node.Empty, child)
			} else {
				node.Children = append(node.Children, child)
			}
		}
	}

	if p.isDirective("endforelse") {
		p.advance()
	}

	return node, nil
}

// parseWhile parses @while...@endwhile
func (p *Parser) parseWhile(pos lexer.Position, condition string) (*WhileNode, error) {
	node := &WhileNode{
		BaseNode:  BaseNode{NodeType: NODE_WHILE, Pos: pos},
		Condition: condition,
		Children:  make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("endwhile") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endwhile") {
		p.advance()
	}

	return node, nil
}

// parseSection parses @section
func (p *Parser) parseSection(pos lexer.Position, args string) (*SectionNode, error) {
	node := &SectionNode{
		BaseNode: BaseNode{NodeType: NODE_SECTION, Pos: pos},
		Children: make([]Node, 0),
	}

	// Check for inline: @section('name', 'content')
	parts := splitArgs(args)
	if len(parts) >= 1 {
		node.Name = trimQuotes(parts[0])
	}
	if len(parts) >= 2 {
		// Inline content
		node.Content = trimQuotes(parts[1])
		return node, nil
	}

	// Block section
	for !p.isAtEnd() && !p.isDirective("endsection") && !p.isDirective("show") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("show") {
		p.advance()
		node.Show = true
	} else if p.isDirective("endsection") {
		p.advance()
	}

	return node, nil
}

// parseYield parses @yield
func (p *Parser) parseYield(pos lexer.Position, args string) (*YieldNode, error) {
	node := &YieldNode{
		BaseNode: BaseNode{NodeType: NODE_YIELD, Pos: pos},
	}

	parts := splitArgs(args)
	if len(parts) >= 1 {
		node.Name = trimQuotes(parts[0])
	}
	if len(parts) >= 2 {
		node.Default = trimQuotes(parts[1])
	}

	return node, nil
}

// parseInclude parses @include variants
func (p *Parser) parseInclude(pos lexer.Position, variant, args string) (*IncludeNode, error) {
	node := &IncludeNode{
		BaseNode: BaseNode{NodeType: NODE_INCLUDE, Pos: pos},
		Variant:  variant,
	}

	parts := splitArgs(args)
	switch variant {
	case "include", "includeIf":
		if len(parts) >= 1 {
			node.Template = trimQuotes(parts[0])
		}
		if len(parts) >= 2 {
			node.Data = parts[1]
		}
	case "includeWhen", "includeUnless":
		if len(parts) >= 1 {
			node.Condition = parts[0]
		}
		if len(parts) >= 2 {
			node.Template = trimQuotes(parts[1])
		}
		if len(parts) >= 3 {
			node.Data = parts[2]
		}
	case "includeFirst":
		if len(parts) >= 1 {
			node.Template = parts[0] // Array of templates
		}
		if len(parts) >= 2 {
			node.Data = parts[1]
		}
	}

	return node, nil
}

// parseEach parses @each
func (p *Parser) parseEach(pos lexer.Position, args string) (*EachNode, error) {
	node := &EachNode{
		BaseNode: BaseNode{NodeType: NODE_EACH, Pos: pos},
	}

	parts := splitArgs(args)
	if len(parts) >= 1 {
		node.Template = trimQuotes(parts[0])
	}
	if len(parts) >= 2 {
		node.Items = parts[1]
	}
	if len(parts) >= 3 {
		node.ItemVar = trimQuotes(parts[2])
	}
	if len(parts) >= 4 {
		node.EmptyView = trimQuotes(parts[3])
	}

	return node, nil
}

// parsePush parses @push...@endpush or @pushOnce...@endPushOnce
func (p *Parser) parsePush(pos lexer.Position, args string, once bool) (*PushNode, error) {
	node := &PushNode{
		BaseNode: BaseNode{NodeType: NODE_PUSH, Pos: pos},
		Stack:    trimQuotes(args),
		Children: make([]Node, 0),
		Once:     once,
	}

	endDirective := "endpush"
	if once {
		endDirective = "endPushOnce"
	}

	for !p.isAtEnd() && !p.isDirective(endDirective) {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective(endDirective) {
		p.advance()
	}

	return node, nil
}

// parsePrepend parses @prepend...@endprepend
func (p *Parser) parsePrepend(pos lexer.Position, args string) (*PrependNode, error) {
	node := &PrependNode{
		BaseNode: BaseNode{NodeType: NODE_PREPEND, Pos: pos},
		Stack:    trimQuotes(args),
		Children: make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("endprepend") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endprepend") {
		p.advance()
	}

	return node, nil
}

// parseComponent parses @component...@endcomponent
func (p *Parser) parseComponent(pos lexer.Position, args string) (*ComponentNode, error) {
	parts := splitArgs(args)
	node := &ComponentNode{
		BaseNode: BaseNode{NodeType: NODE_COMPONENT, Pos: pos},
		Children: make([]Node, 0),
		Slots:    make(map[string]*SlotNode),
	}

	if len(parts) >= 1 {
		node.Name = trimQuotes(parts[0])
	}
	if len(parts) >= 2 {
		node.Data = parts[1]
	}

	var currentSlot *SlotNode

	for !p.isAtEnd() && !p.isDirective("endcomponent") {
		if p.isDirective("slot") {
			if currentSlot != nil {
				node.Slots[currentSlot.Name] = currentSlot
			}
			slotToken := p.current
			p.advance()
			currentSlot = &SlotNode{
				BaseNode: BaseNode{NodeType: NODE_SLOT, Pos: slotToken.Position},
				Name:     trimQuotes(slotToken.Args),
				Children: make([]Node, 0),
			}
			continue
		}

		if p.isDirective("endslot") {
			if currentSlot != nil {
				node.Slots[currentSlot.Name] = currentSlot
				currentSlot = nil
			}
			p.advance()
			continue
		}

		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			if currentSlot != nil {
				currentSlot.Children = append(currentSlot.Children, child)
			} else {
				node.Children = append(node.Children, child)
			}
		}
	}

	if currentSlot != nil {
		node.Slots[currentSlot.Name] = currentSlot
	}

	if p.isDirective("endcomponent") {
		p.advance()
	}

	return node, nil
}

// parseVerbatim parses @verbatim...@endverbatim
func (p *Parser) parseVerbatim() (*VerbatimNode, error) {
	pos := p.current.Position
	p.advance()

	var content strings.Builder

	for !p.isAtEnd() {
		if p.current.Type == lexer.TOKEN_VERBATIM_END {
			p.advance()
			break
		}
		if p.current.Type == lexer.TOKEN_TEXT {
			content.WriteString(p.current.Value)
		}
		p.advance()
	}

	return &VerbatimNode{
		BaseNode: BaseNode{NodeType: NODE_VERBATIM, Pos: pos},
		Content:  content.String(),
	}, nil
}

// parsePhp parses @php...@endphp
func (p *Parser) parsePhp(pos lexer.Position) (*PhpNode, error) {
	var code strings.Builder

	for !p.isAtEnd() && !p.isDirective("endphp") {
		if p.current.Type == lexer.TOKEN_TEXT {
			code.WriteString(p.current.Value)
		}
		p.advance()
	}

	if p.isDirective("endphp") {
		p.advance()
	}

	return &PhpNode{
		BaseNode: BaseNode{NodeType: NODE_PHP, Pos: pos},
		Code:     strings.TrimSpace(code.String()),
	}, nil
}

// parseIsset parses @isset...@endisset
func (p *Parser) parseIsset(pos lexer.Position, variable string) (*IssetNode, error) {
	node := &IssetNode{
		BaseNode: BaseNode{NodeType: NODE_ISSET, Pos: pos},
		Variable: variable,
		Children: make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("endisset") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endisset") {
		p.advance()
	}

	return node, nil
}

// parseEmptyCheck parses @empty...@endempty
func (p *Parser) parseEmptyCheck(pos lexer.Position, variable string) (*EmptyCheckNode, error) {
	node := &EmptyCheckNode{
		BaseNode: BaseNode{NodeType: NODE_EMPTY, Pos: pos},
		Variable: variable,
		Children: make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("endempty") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endempty") {
		p.advance()
	}

	return node, nil
}

// parseAuth parses @auth...@endauth
func (p *Parser) parseAuth(pos lexer.Position, guard string) (*AuthNode, error) {
	node := &AuthNode{
		BaseNode: BaseNode{NodeType: NODE_AUTH, Pos: pos},
		Guard:    trimQuotes(guard),
		Children: make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("endauth") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endauth") {
		p.advance()
	}

	return node, nil
}

// parseGuest parses @guest...@endguest
func (p *Parser) parseGuest(pos lexer.Position, guard string) (*GuestNode, error) {
	node := &GuestNode{
		BaseNode: BaseNode{NodeType: NODE_GUEST, Pos: pos},
		Guard:    trimQuotes(guard),
		Children: make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("endguest") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endguest") {
		p.advance()
	}

	return node, nil
}

// parseEnv parses @env...@endenv
func (p *Parser) parseEnv(pos lexer.Position, args string) (*EnvNode, error) {
	node := &EnvNode{
		BaseNode:     BaseNode{NodeType: NODE_ENV, Pos: pos},
		Environments: parseEnvList(args),
		Children:     make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("endenv") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endenv") {
		p.advance()
	}

	return node, nil
}

// parseProduction parses @production...@endproduction
func (p *Parser) parseProduction(pos lexer.Position) (*ProductionNode, error) {
	node := &ProductionNode{
		BaseNode: BaseNode{NodeType: NODE_PRODUCTION, Pos: pos},
		Children: make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("endproduction") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endproduction") {
		p.advance()
	}

	return node, nil
}

// parseError parses @error...@enderror
func (p *Parser) parseError(pos lexer.Position, field string) (*ErrorNode, error) {
	node := &ErrorNode{
		BaseNode: BaseNode{NodeType: NODE_ERROR, Pos: pos},
		Field:    trimQuotes(field),
		Children: make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("enderror") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("enderror") {
		p.advance()
	}

	return node, nil
}

// parseOnce parses @once...@endonce
func (p *Parser) parseOnce(pos lexer.Position) (*OnceNode, error) {
	node := &OnceNode{
		BaseNode: BaseNode{NodeType: NODE_ONCE, Pos: pos},
		Children: make([]Node, 0),
	}

	for !p.isAtEnd() && !p.isDirective("endonce") {
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	if p.isDirective("endonce") {
		p.advance()
	}

	return node, nil
}

// Helper methods

func (p *Parser) advance() {
	p.pos++
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
	}
}

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.current.Type == lexer.TOKEN_EOF
}

func (p *Parser) isDirective(name string) bool {
	return (p.current.Type == lexer.TOKEN_DIRECTIVE || p.current.Type == lexer.TOKEN_DIRECTIVE_ARGS) && p.current.Value == name
}

// trimQuotes removes surrounding quotes from a string
func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// splitArgs splits comma-separated arguments respecting strings and brackets
func splitArgs(args string) []string {
	var result []string
	var current strings.Builder
	depth := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(args); i++ {
		ch := args[i]

		if (ch == '"' || ch == '\'') && (i == 0 || args[i-1] != '\\') {
			if !inString {
				inString = true
				stringChar = ch
			} else if ch == stringChar {
				inString = false
			}
		}

		if !inString {
			if ch == '(' || ch == '[' || ch == '{' {
				depth++
			} else if ch == ')' || ch == ']' || ch == '}' {
				depth--
			} else if ch == ',' && depth == 0 {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
				continue
			}
		}

		current.WriteByte(ch)
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

// parseEnvList parses environment list from @env argument
func parseEnvList(args string) []string {
	args = strings.TrimSpace(args)

	// Check if it's an array ['local', 'staging']
	if strings.HasPrefix(args, "[") && strings.HasSuffix(args, "]") {
		args = args[1 : len(args)-1]
		parts := splitArgs(args)
		for i, p := range parts {
			parts[i] = trimQuotes(p)
		}
		return parts
	}

	// Single environment
	return []string{trimQuotes(args)}
}

// ParserError represents a parser error
type ParserError struct {
	Message  string
	Position lexer.Position
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("%s at line %d, column %d", e.Message, e.Position.Line, e.Position.Column)
}
