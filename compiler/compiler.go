package compiler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/codingersid/legit-template/parser"
)

// Compiler compiles AST to Go template string
type Compiler struct {
	// Template inheritance
	extends     string
	sections    map[string]string
	parentCalls map[string]bool

	// Stacks
	pushes   map[string][]string
	prepends map[string][]string

	// State
	loopDepth int
	onceKeys  map[string]bool
}

// New creates a new Compiler
func New() *Compiler {
	return &Compiler{
		sections:    make(map[string]string),
		parentCalls: make(map[string]bool),
		pushes:      make(map[string][]string),
		prepends:    make(map[string][]string),
		onceKeys:    make(map[string]bool),
	}
}

// Compile compiles AST to Go template string
func (c *Compiler) Compile(root *parser.RootNode) (string, error) {
	var result strings.Builder

	for _, node := range root.Children {
		compiled, err := c.compileNode(node)
		if err != nil {
			return "", err
		}
		result.WriteString(compiled)
	}

	return result.String(), nil
}

// GetExtends returns the parent template name if @extends was used
func (c *Compiler) GetExtends() string {
	return c.extends
}

// GetSections returns all defined sections
func (c *Compiler) GetSections() map[string]string {
	return c.sections
}

// GetStacks returns push content for a stack
func (c *Compiler) GetPushes(name string) []string {
	return c.pushes[name]
}

// GetPrepends returns prepend content for a stack
func (c *Compiler) GetPrepends(name string) []string {
	return c.prepends[name]
}

// HasParentCall checks if a section has @parent
func (c *Compiler) HasParentCall(section string) bool {
	return c.parentCalls[section]
}

// compileNode compiles a single node
func (c *Compiler) compileNode(node parser.Node) (string, error) {
	switch n := node.(type) {
	case *parser.TextNode:
		return n.Content, nil

	case *parser.EchoNode:
		return c.compileEcho(n), nil

	case *parser.CommentNode:
		return "", nil // Comments are not rendered

	case *parser.DirectiveNode:
		return c.compileDirective(n), nil

	case *parser.IfNode:
		return c.compileIf(n)

	case *parser.UnlessNode:
		return c.compileUnless(n)

	case *parser.SwitchNode:
		return c.compileSwitch(n)

	case *parser.ForNode:
		return c.compileFor(n)

	case *parser.ForeachNode:
		return c.compileForeach(n)

	case *parser.ForelseNode:
		return c.compileForelse(n)

	case *parser.WhileNode:
		return c.compileWhile(n)

	case *parser.SectionNode:
		return c.compileSection(n)

	case *parser.YieldNode:
		return c.compileYield(n), nil

	case *parser.ExtendsNode:
		c.extends = n.Template
		return "", nil

	case *parser.IncludeNode:
		return c.compileInclude(n), nil

	case *parser.EachNode:
		return c.compileEach(n), nil

	case *parser.PushNode:
		return c.compilePush(n)

	case *parser.PrependNode:
		return c.compilePrepend(n)

	case *parser.StackNode:
		return c.compileStack(n), nil

	case *parser.ComponentNode:
		return c.compileComponent(n)

	case *parser.VerbatimNode:
		return n.Content, nil

	case *parser.PhpNode:
		return c.compilePhp(n), nil

	case *parser.IssetNode:
		return c.compileIsset(n)

	case *parser.EmptyCheckNode:
		return c.compileEmptyCheck(n)

	case *parser.AuthNode:
		return c.compileAuth(n)

	case *parser.GuestNode:
		return c.compileGuest(n)

	case *parser.EnvNode:
		return c.compileEnv(n)

	case *parser.ProductionNode:
		return c.compileProduction(n)

	case *parser.ErrorNode:
		return c.compileError(n)

	case *parser.OnceNode:
		return c.compileOnce(n)

	case *parser.BreakNode:
		return c.compileBreak(n), nil

	case *parser.ContinueNode:
		return c.compileContinue(n), nil

	case *parser.ParentNode:
		return "{{__PARENT__}}", nil

	default:
		return "", nil
	}
}

// compileChildren compiles children nodes
func (c *Compiler) compileChildren(children []parser.Node) (string, error) {
	var result strings.Builder
	for _, child := range children {
		compiled, err := c.compileNode(child)
		if err != nil {
			return "", err
		}
		result.WriteString(compiled)
	}
	return result.String(), nil
}

// compileEcho compiles {{ }} and {!! !!}
func (c *Compiler) compileEcho(n *parser.EchoNode) string {
	expr := c.transformExpression(n.Expression)
	if n.Escaped {
		return fmt.Sprintf("{{ html %s }}", expr)
	}
	return fmt.Sprintf("{{ %s }}", expr)
}

// compileDirective compiles simple directives
func (c *Compiler) compileDirective(n *parser.DirectiveNode) string {
	switch n.Name {
	case "csrf":
		return `<input type="hidden" name="_token" value="{{ .csrf_token }}">`
	case "method":
		method := strings.Trim(n.Args, "'\"")
		return fmt.Sprintf(`<input type="hidden" name="_method" value="%s">`, method)
	case "json":
		expr := c.transformExpression(n.Args)
		return fmt.Sprintf("{{ json %s }}", expr)
	case "class":
		return c.compileClass(n.Args)
	case "style":
		return c.compileStyle(n.Args)
	case "checked":
		expr := c.transformExpression(n.Args)
		return fmt.Sprintf(`{{ if %s }}checked{{ end }}`, expr)
	case "selected":
		expr := c.transformExpression(n.Args)
		return fmt.Sprintf(`{{ if %s }}selected{{ end }}`, expr)
	case "disabled":
		expr := c.transformExpression(n.Args)
		return fmt.Sprintf(`{{ if %s }}disabled{{ end }}`, expr)
	case "readonly":
		expr := c.transformExpression(n.Args)
		return fmt.Sprintf(`{{ if %s }}readonly{{ end }}`, expr)
	case "required":
		expr := c.transformExpression(n.Args)
		return fmt.Sprintf(`{{ if %s }}required{{ end }}`, expr)
	case "old":
		field := strings.Trim(n.Args, "'\"")
		return fmt.Sprintf(`{{ index .old "%s" }}`, field)
	default:
		// Custom directive - call as function
		if n.Args != "" {
			return fmt.Sprintf("{{ %s %s }}", n.Name, c.transformExpression(n.Args))
		}
		return fmt.Sprintf("{{ %s }}", n.Name)
	}
}

// compileClass compiles @class directive
func (c *Compiler) compileClass(args string) string {
	// @class(['p-4', 'font-bold' => $isActive])
	// TODO: Implement proper parsing of class array
	return fmt.Sprintf(`class="{{ classArray %s }}"`, args)
}

// compileStyle compiles @style directive
func (c *Compiler) compileStyle(args string) string {
	// @style(['color: red' => $hasError])
	return fmt.Sprintf(`style="{{ styleArray %s }}"`, args)
}

// compileIf compiles @if...@endif
func (c *Compiler) compileIf(n *parser.IfNode) (string, error) {
	var result strings.Builder

	condition := c.transformExpression(n.Condition)
	result.WriteString(fmt.Sprintf("{{ if %s }}", condition))

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)

	for _, elseif := range n.ElseIfs {
		elseifCond := c.transformExpression(elseif.Condition)
		result.WriteString(fmt.Sprintf("{{ else if %s }}", elseifCond))

		elseifChildren, err := c.compileChildren(elseif.Children)
		if err != nil {
			return "", err
		}
		result.WriteString(elseifChildren)
	}

	if n.Else != nil {
		result.WriteString("{{ else }}")
		elseChildren, err := c.compileChildren(n.Else.Children)
		if err != nil {
			return "", err
		}
		result.WriteString(elseChildren)
	}

	result.WriteString("{{ end }}")
	return result.String(), nil
}

// compileUnless compiles @unless...@endunless
func (c *Compiler) compileUnless(n *parser.UnlessNode) (string, error) {
	var result strings.Builder

	condition := c.transformExpression(n.Condition)
	result.WriteString(fmt.Sprintf("{{ if not %s }}", condition))

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileSwitch compiles @switch...@endswitch
func (c *Compiler) compileSwitch(n *parser.SwitchNode) (string, error) {
	var result strings.Builder

	expr := c.transformExpression(n.Expression)

	for i, caseNode := range n.Cases {
		caseVal := c.transformExpression(caseNode.Value)
		if i == 0 {
			result.WriteString(fmt.Sprintf("{{ if eq %s %s }}", expr, caseVal))
		} else {
			result.WriteString(fmt.Sprintf("{{ else if eq %s %s }}", expr, caseVal))
		}

		caseChildren, err := c.compileChildren(caseNode.Children)
		if err != nil {
			return "", err
		}
		result.WriteString(caseChildren)
	}

	if n.Default != nil {
		result.WriteString("{{ else }}")
		defaultChildren, err := c.compileChildren(n.Default.Children)
		if err != nil {
			return "", err
		}
		result.WriteString(defaultChildren)
	}

	if len(n.Cases) > 0 || n.Default != nil {
		result.WriteString("{{ end }}")
	}

	return result.String(), nil
}

// compileFor compiles @for...@endfor
func (c *Compiler) compileFor(n *parser.ForNode) (string, error) {
	c.loopDepth++
	defer func() { c.loopDepth-- }()

	var result strings.Builder

	// Convert PHP-style for to Go range
	// @for($i = 0; $i < 10; $i++) -> {{ range $i := seq 0 10 }}
	// This is a simplified conversion - real implementation needs expression parsing
	result.WriteString(fmt.Sprintf("{{ $__loop%d := newLoop -1 %d }}", c.loopDepth, c.loopDepth))
	result.WriteString(fmt.Sprintf("{{ range $__idx%d := seq %s }}", c.loopDepth, c.extractForRange(n)))

	result.WriteString(fmt.Sprintf("{{ $loop := $__loop%d.Update $__idx%d }}", c.loopDepth, c.loopDepth))

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// extractForRange extracts range parameters from for loop
func (c *Compiler) extractForRange(n *parser.ForNode) string {
	// Simple extraction: $i = 0; $i < 10 -> 0 10
	// This is simplified - real implementation needs proper parsing
	init := strings.TrimPrefix(n.Init, "$")
	if idx := strings.Index(init, "="); idx != -1 {
		init = strings.TrimSpace(init[idx+1:])
	}

	cond := n.Condition
	// Extract end value from $i < 10 or $i <= 9
	re := regexp.MustCompile(`<\s*=?\s*(\d+)`)
	matches := re.FindStringSubmatch(cond)
	end := "10"
	if len(matches) > 1 {
		end = matches[1]
	}

	return fmt.Sprintf("%s %s", init, end)
}

// compileForeach compiles @foreach...@endforeach
func (c *Compiler) compileForeach(n *parser.ForeachNode) (string, error) {
	c.loopDepth++
	defer func() { c.loopDepth-- }()

	var result strings.Builder

	items := c.transformExpression(n.Items)
	key := n.Key
	value := n.Value

	if key == "" {
		key = "_"
	}
	key = strings.TrimPrefix(key, "$")
	value = strings.TrimPrefix(value, "$")

	// Initialize loop variable
	result.WriteString(fmt.Sprintf("{{ $__loop%d := newLoop (len %s) %d }}", c.loopDepth, items, c.loopDepth))

	if key == "_" {
		result.WriteString(fmt.Sprintf("{{ range $__idx%d, $%s := %s }}", c.loopDepth, value, items))
	} else {
		result.WriteString(fmt.Sprintf("{{ range $%s, $%s := %s }}", key, value, items))
	}

	// Update loop on each iteration
	if key == "_" {
		result.WriteString(fmt.Sprintf("{{ $loop := $__loop%d.Update $__idx%d }}", c.loopDepth, c.loopDepth))
	} else {
		result.WriteString(fmt.Sprintf("{{ $loop := $__loop%d.Update $%s }}", c.loopDepth, key))
	}

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileForelse compiles @forelse...@empty...@endforelse
func (c *Compiler) compileForelse(n *parser.ForelseNode) (string, error) {
	c.loopDepth++
	defer func() { c.loopDepth-- }()

	var result strings.Builder

	items := c.transformExpression(n.Items)
	key := n.Key
	value := n.Value

	if key == "" {
		key = "_"
	}
	key = strings.TrimPrefix(key, "$")
	value = strings.TrimPrefix(value, "$")

	// Check if items is not empty
	result.WriteString(fmt.Sprintf("{{ if %s }}", items))
	result.WriteString(fmt.Sprintf("{{ $__loop%d := newLoop (len %s) %d }}", c.loopDepth, items, c.loopDepth))

	if key == "_" {
		result.WriteString(fmt.Sprintf("{{ range $__idx%d, $%s := %s }}", c.loopDepth, value, items))
	} else {
		result.WriteString(fmt.Sprintf("{{ range $%s, $%s := %s }}", key, value, items))
	}

	if key == "_" {
		result.WriteString(fmt.Sprintf("{{ $loop := $__loop%d.Update $__idx%d }}", c.loopDepth, c.loopDepth))
	} else {
		result.WriteString(fmt.Sprintf("{{ $loop := $__loop%d.Update $%s }}", c.loopDepth, key))
	}

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	// Empty block
	result.WriteString("{{ else }}")
	empty, err := c.compileChildren(n.Empty)
	if err != nil {
		return "", err
	}
	result.WriteString(empty)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileWhile compiles @while...@endwhile
func (c *Compiler) compileWhile(n *parser.WhileNode) (string, error) {
	c.loopDepth++
	defer func() { c.loopDepth-- }()

	var result strings.Builder

	// Go templates don't have while loops, so we use a workaround with range and break
	// This is a simplified implementation
	condition := c.transformExpression(n.Condition)
	result.WriteString(fmt.Sprintf("{{ $__loop%d := newLoop -1 %d }}", c.loopDepth, c.loopDepth))
	result.WriteString(fmt.Sprintf("{{ range $__idx%d := until 1000 }}", c.loopDepth))
	result.WriteString(fmt.Sprintf("{{ if not %s }}{{ break }}{{ end }}", condition))
	result.WriteString(fmt.Sprintf("{{ $loop := $__loop%d.Update $__idx%d }}", c.loopDepth, c.loopDepth))

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileSection compiles @section
func (c *Compiler) compileSection(n *parser.SectionNode) (string, error) {
	if n.Content != "" {
		// Inline section
		c.sections[n.Name] = n.Content
		return "", nil
	}

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}

	// Check for @parent
	if strings.Contains(children, "{{__PARENT__}}") {
		c.parentCalls[n.Name] = true
	}

	c.sections[n.Name] = children

	if n.Show {
		// @show outputs immediately
		return fmt.Sprintf("{{ block \"%s\" . }}%s{{ end }}", n.Name, children), nil
	}

	return "", nil
}

// compileYield compiles @yield
func (c *Compiler) compileYield(n *parser.YieldNode) string {
	if n.Default != "" {
		return fmt.Sprintf("{{ block \"%s\" . }}%s{{ end }}", n.Name, n.Default)
	}
	return fmt.Sprintf("{{ block \"%s\" . }}{{ end }}", n.Name)
}

// compileInclude compiles @include variants
func (c *Compiler) compileInclude(n *parser.IncludeNode) string {
	switch n.Variant {
	case "include":
		if n.Data != "" {
			return fmt.Sprintf("{{ template \"%s\" (merge . %s) }}", n.Template, n.Data)
		}
		return fmt.Sprintf("{{ template \"%s\" . }}", n.Template)
	case "includeIf":
		if n.Data != "" {
			return fmt.Sprintf("{{ if templateExists \"%s\" }}{{ template \"%s\" (merge . %s) }}{{ end }}", n.Template, n.Template, n.Data)
		}
		return fmt.Sprintf("{{ if templateExists \"%s\" }}{{ template \"%s\" . }}{{ end }}", n.Template, n.Template)
	case "includeWhen":
		cond := c.transformExpression(n.Condition)
		if n.Data != "" {
			return fmt.Sprintf("{{ if %s }}{{ template \"%s\" (merge . %s) }}{{ end }}", cond, n.Template, n.Data)
		}
		return fmt.Sprintf("{{ if %s }}{{ template \"%s\" . }}{{ end }}", cond, n.Template)
	case "includeUnless":
		cond := c.transformExpression(n.Condition)
		if n.Data != "" {
			return fmt.Sprintf("{{ if not %s }}{{ template \"%s\" (merge . %s) }}{{ end }}", cond, n.Template, n.Data)
		}
		return fmt.Sprintf("{{ if not %s }}{{ template \"%s\" . }}{{ end }}", cond, n.Template)
	case "includeFirst":
		return fmt.Sprintf("{{ includeFirst %s . }}", n.Template)
	}
	return ""
}

// compileEach compiles @each
func (c *Compiler) compileEach(n *parser.EachNode) string {
	items := c.transformExpression(n.Items)
	if n.EmptyView != "" {
		return fmt.Sprintf("{{ each \"%s\" %s \"%s\" \"%s\" }}", n.Template, items, n.ItemVar, n.EmptyView)
	}
	return fmt.Sprintf("{{ each \"%s\" %s \"%s\" \"\" }}", n.Template, items, n.ItemVar)
}

// compilePush compiles @push...@endpush
func (c *Compiler) compilePush(n *parser.PushNode) (string, error) {
	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}

	if n.Once {
		key := fmt.Sprintf("push_%s_%s", n.Stack, children)
		if c.onceKeys[key] {
			return "", nil
		}
		c.onceKeys[key] = true
	}

	c.pushes[n.Stack] = append(c.pushes[n.Stack], children)
	return "", nil
}

// compilePrepend compiles @prepend...@endprepend
func (c *Compiler) compilePrepend(n *parser.PrependNode) (string, error) {
	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}

	c.prepends[n.Stack] = append([]string{children}, c.prepends[n.Stack]...)
	return "", nil
}

// compileStack compiles @stack
func (c *Compiler) compileStack(n *parser.StackNode) string {
	return fmt.Sprintf("{{ stack \"%s\" }}", n.Name)
}

// compileComponent compiles @component...@endcomponent
func (c *Compiler) compileComponent(n *parser.ComponentNode) (string, error) {
	var result strings.Builder

	// Compile default slot (children)
	defaultSlot, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}

	// Build slots map
	result.WriteString(fmt.Sprintf("{{ $__slots := dict \"default\" `%s`", escapeBackticks(defaultSlot)))

	for name, slot := range n.Slots {
		slotContent, err := c.compileChildren(slot.Children)
		if err != nil {
			return "", err
		}
		result.WriteString(fmt.Sprintf(" \"%s\" `%s`", name, escapeBackticks(slotContent)))
	}
	result.WriteString(" }}")

	// Render component
	if n.Data != "" {
		result.WriteString(fmt.Sprintf("{{ template \"components/%s\" (merge . (dict \"slot\" (index $__slots \"default\") \"slots\" $__slots) %s) }}", n.Name, n.Data))
	} else {
		result.WriteString(fmt.Sprintf("{{ template \"components/%s\" (merge . (dict \"slot\" (index $__slots \"default\") \"slots\" $__slots)) }}", n.Name))
	}

	return result.String(), nil
}

// compilePhp compiles @php...@endphp
func (c *Compiler) compilePhp(n *parser.PhpNode) string {
	// Map PHP-like code to Go template actions
	// This is a simplified implementation
	return fmt.Sprintf("{{ /* php: %s */ }}", n.Code)
}

// compileIsset compiles @isset...@endisset
func (c *Compiler) compileIsset(n *parser.IssetNode) (string, error) {
	var result strings.Builder

	variable := c.transformExpression(n.Variable)
	result.WriteString(fmt.Sprintf("{{ if isset %s }}", variable))

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileEmptyCheck compiles @empty...@endempty
func (c *Compiler) compileEmptyCheck(n *parser.EmptyCheckNode) (string, error) {
	var result strings.Builder

	variable := c.transformExpression(n.Variable)
	result.WriteString(fmt.Sprintf("{{ if empty %s }}", variable))

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileAuth compiles @auth...@endauth
func (c *Compiler) compileAuth(n *parser.AuthNode) (string, error) {
	var result strings.Builder

	if n.Guard != "" {
		result.WriteString(fmt.Sprintf("{{ if auth \"%s\" }}", n.Guard))
	} else {
		result.WriteString("{{ if .auth }}")
	}

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileGuest compiles @guest...@endguest
func (c *Compiler) compileGuest(n *parser.GuestNode) (string, error) {
	var result strings.Builder

	if n.Guard != "" {
		result.WriteString(fmt.Sprintf("{{ if not (auth \"%s\") }}", n.Guard))
	} else {
		result.WriteString("{{ if not .auth }}")
	}

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileEnv compiles @env...@endenv
func (c *Compiler) compileEnv(n *parser.EnvNode) (string, error) {
	var result strings.Builder

	if len(n.Environments) == 1 {
		result.WriteString(fmt.Sprintf("{{ if eq .env \"%s\" }}", n.Environments[0]))
	} else {
		conditions := make([]string, len(n.Environments))
		for i, env := range n.Environments {
			conditions[i] = fmt.Sprintf("(eq .env \"%s\")", env)
		}
		result.WriteString(fmt.Sprintf("{{ if or %s }}", strings.Join(conditions, " ")))
	}

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileProduction compiles @production...@endproduction
func (c *Compiler) compileProduction(n *parser.ProductionNode) (string, error) {
	var result strings.Builder

	result.WriteString(`{{ if eq .env "production" }}`)

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileError compiles @error...@enderror
func (c *Compiler) compileError(n *parser.ErrorNode) (string, error) {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("{{ if hasError .errors \"%s\" }}", n.Field))
	result.WriteString(fmt.Sprintf("{{ $message := getError .errors \"%s\" }}", n.Field))

	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}
	result.WriteString(children)
	result.WriteString("{{ end }}")

	return result.String(), nil
}

// compileOnce compiles @once...@endonce
func (c *Compiler) compileOnce(n *parser.OnceNode) (string, error) {
	children, err := c.compileChildren(n.Children)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("once_%s", children)
	if c.onceKeys[key] {
		return "", nil
	}
	c.onceKeys[key] = true

	return children, nil
}

// compileBreak compiles @break
func (c *Compiler) compileBreak(n *parser.BreakNode) string {
	if n.Condition != "" {
		cond := c.transformExpression(n.Condition)
		return fmt.Sprintf("{{ if %s }}{{ break }}{{ end }}", cond)
	}
	return "{{ break }}"
}

// compileContinue compiles @continue
func (c *Compiler) compileContinue(n *parser.ContinueNode) string {
	if n.Condition != "" {
		cond := c.transformExpression(n.Condition)
		return fmt.Sprintf("{{ if %s }}{{ continue }}{{ end }}", cond)
	}
	return "{{ continue }}"
}

// transformExpression transforms PHP-style expression to Go template
func (c *Compiler) transformExpression(expr string) string {
	expr = strings.TrimSpace(expr)

	// Transform $variable to .variable
	re := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)`)
	expr = re.ReplaceAllString(expr, ".$1")

	// Transform -> to .
	expr = strings.ReplaceAll(expr, "->", ".")

	// Transform array access $arr['key'] to (index .arr "key")
	arrayRe := regexp.MustCompile(`\.([a-zA-Z_][a-zA-Z0-9_]*)\[['"]([^'"]+)['"]\]`)
	expr = arrayRe.ReplaceAllString(expr, `(index .$1 "$2")`)

	// Transform !== to ne
	expr = strings.ReplaceAll(expr, "!==", " ne ")
	expr = strings.ReplaceAll(expr, "!=", " ne ")

	// Transform === to eq
	expr = strings.ReplaceAll(expr, "===", " eq ")
	expr = strings.ReplaceAll(expr, "==", " eq ")

	// Transform && to and
	expr = strings.ReplaceAll(expr, "&&", " and ")

	// Transform || to or
	expr = strings.ReplaceAll(expr, "||", " or ")

	// Transform ! to not (careful with != already transformed)
	expr = regexp.MustCompile(`!([^=])`).ReplaceAllString(expr, "not $1")

	// Transform >= and <=
	expr = strings.ReplaceAll(expr, ">=", " gte ")
	expr = strings.ReplaceAll(expr, "<=", " lte ")
	expr = strings.ReplaceAll(expr, ">", " gt ")
	expr = strings.ReplaceAll(expr, "<", " lt ")

	// Clean up multiple spaces
	expr = regexp.MustCompile(`\s+`).ReplaceAllString(expr, " ")

	return strings.TrimSpace(expr)
}

// escapeBackticks escapes backticks in string for Go raw string literals
func escapeBackticks(s string) string {
	return strings.ReplaceAll(s, "`", "` + \"`\" + `")
}
