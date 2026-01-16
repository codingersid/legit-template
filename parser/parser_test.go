package parser

import (
	"testing"

	"github.com/codingersid/legit-template/lexer"
)

func parseTemplate(t *testing.T, input string) *RootNode {
	lex := lexer.New(input)
	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	p := New(tokens)
	ast, err := p.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	return ast
}

func TestParser_Text(t *testing.T) {
	ast := parseTemplate(t, "Hello World")

	if len(ast.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(ast.Children))
	}

	node, ok := ast.Children[0].(*TextNode)
	if !ok {
		t.Fatal("expected TextNode")
	}

	if node.Content != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", node.Content)
	}
}

func TestParser_EscapedEcho(t *testing.T) {
	ast := parseTemplate(t, "{{ $name }}")

	if len(ast.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(ast.Children))
	}

	node, ok := ast.Children[0].(*EchoNode)
	if !ok {
		t.Fatal("expected EchoNode")
	}

	if !node.Escaped {
		t.Error("expected escaped echo")
	}

	if node.Expression != "$name" {
		t.Errorf("expected '$name', got %q", node.Expression)
	}
}

func TestParser_RawEcho(t *testing.T) {
	ast := parseTemplate(t, "{!! $html !!}")

	node, ok := ast.Children[0].(*EchoNode)
	if !ok {
		t.Fatal("expected EchoNode")
	}

	if node.Escaped {
		t.Error("expected raw echo")
	}
}

func TestParser_If(t *testing.T) {
	ast := parseTemplate(t, "@if($condition)content@endif")

	if len(ast.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(ast.Children))
	}

	node, ok := ast.Children[0].(*IfNode)
	if !ok {
		t.Fatal("expected IfNode")
	}

	if node.Condition != "$condition" {
		t.Errorf("expected '$condition', got %q", node.Condition)
	}

	if len(node.Children) != 1 {
		t.Fatalf("expected 1 child in if, got %d", len(node.Children))
	}
}

func TestParser_IfElseIf(t *testing.T) {
	ast := parseTemplate(t, "@if($a)A@elseif($b)B@else C@endif")

	node := ast.Children[0].(*IfNode)

	if len(node.ElseIfs) != 1 {
		t.Fatalf("expected 1 elseif, got %d", len(node.ElseIfs))
	}

	if node.Else == nil {
		t.Fatal("expected else node")
	}
}

func TestParser_Unless(t *testing.T) {
	ast := parseTemplate(t, "@unless($condition)content@endunless")

	node, ok := ast.Children[0].(*UnlessNode)
	if !ok {
		t.Fatal("expected UnlessNode")
	}

	if node.Condition != "$condition" {
		t.Errorf("expected '$condition', got %q", node.Condition)
	}
}

func TestParser_Foreach(t *testing.T) {
	ast := parseTemplate(t, "@foreach($items as $item){{ $item }}@endforeach")

	node, ok := ast.Children[0].(*ForeachNode)
	if !ok {
		t.Fatal("expected ForeachNode")
	}

	if node.Items != "$items" {
		t.Errorf("expected '$items', got %q", node.Items)
	}

	if node.Value != "$item" {
		t.Errorf("expected '$item', got %q", node.Value)
	}
}

func TestParser_ForeachKeyValue(t *testing.T) {
	ast := parseTemplate(t, "@foreach($items as $key => $value)@endforeach")

	node := ast.Children[0].(*ForeachNode)

	if node.Key != "$key" {
		t.Errorf("expected '$key', got %q", node.Key)
	}

	if node.Value != "$value" {
		t.Errorf("expected '$value', got %q", node.Value)
	}
}

func TestParser_Forelse(t *testing.T) {
	ast := parseTemplate(t, "@forelse($items as $item){{ $item }}@empty No items@endforelse")

	node, ok := ast.Children[0].(*ForelseNode)
	if !ok {
		t.Fatal("expected ForelseNode")
	}

	if len(node.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(node.Children))
	}

	if len(node.Empty) != 1 {
		t.Errorf("expected 1 empty child, got %d", len(node.Empty))
	}
}

func TestParser_For(t *testing.T) {
	ast := parseTemplate(t, "@for($i = 0; $i < 10; $i++){{ $i }}@endfor")

	node, ok := ast.Children[0].(*ForNode)
	if !ok {
		t.Fatal("expected ForNode")
	}

	if node.Init != "$i = 0" {
		t.Errorf("expected '$i = 0', got %q", node.Init)
	}
}

func TestParser_While(t *testing.T) {
	ast := parseTemplate(t, "@while($condition)content@endwhile")

	node, ok := ast.Children[0].(*WhileNode)
	if !ok {
		t.Fatal("expected WhileNode")
	}

	if node.Condition != "$condition" {
		t.Errorf("expected '$condition', got %q", node.Condition)
	}
}

func TestParser_Switch(t *testing.T) {
	ast := parseTemplate(t, "@switch($type)@case('a')A@break@case('b')B@break@default Default@endswitch")

	node, ok := ast.Children[0].(*SwitchNode)
	if !ok {
		t.Fatal("expected SwitchNode")
	}

	if len(node.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(node.Cases))
	}

	if node.Default == nil {
		t.Fatal("expected default node")
	}
}

func TestParser_Extends(t *testing.T) {
	ast := parseTemplate(t, "@extends('layouts.app')")

	node, ok := ast.Children[0].(*ExtendsNode)
	if !ok {
		t.Fatal("expected ExtendsNode")
	}

	if node.Template != "layouts.app" {
		t.Errorf("expected 'layouts.app', got %q", node.Template)
	}
}

func TestParser_Section(t *testing.T) {
	ast := parseTemplate(t, "@section('content')Hello@endsection")

	node, ok := ast.Children[0].(*SectionNode)
	if !ok {
		t.Fatal("expected SectionNode")
	}

	if node.Name != "content" {
		t.Errorf("expected 'content', got %q", node.Name)
	}
}

func TestParser_SectionInline(t *testing.T) {
	ast := parseTemplate(t, "@section('title', 'Page Title')")

	node, ok := ast.Children[0].(*SectionNode)
	if !ok {
		t.Fatal("expected SectionNode")
	}

	if node.Content != "Page Title" {
		t.Errorf("expected 'Page Title', got %q", node.Content)
	}
}

func TestParser_Yield(t *testing.T) {
	ast := parseTemplate(t, "@yield('content', 'default')")

	node, ok := ast.Children[0].(*YieldNode)
	if !ok {
		t.Fatal("expected YieldNode")
	}

	if node.Name != "content" {
		t.Errorf("expected 'content', got %q", node.Name)
	}

	if node.Default != "default" {
		t.Errorf("expected 'default', got %q", node.Default)
	}
}

func TestParser_Include(t *testing.T) {
	ast := parseTemplate(t, "@include('partials.header')")

	node, ok := ast.Children[0].(*IncludeNode)
	if !ok {
		t.Fatal("expected IncludeNode")
	}

	if node.Template != "partials.header" {
		t.Errorf("expected 'partials.header', got %q", node.Template)
	}
}

func TestParser_IncludeWhen(t *testing.T) {
	ast := parseTemplate(t, "@includeWhen($condition, 'partials.header')")

	node, ok := ast.Children[0].(*IncludeNode)
	if !ok {
		t.Fatal("expected IncludeNode")
	}

	if node.Variant != "includeWhen" {
		t.Errorf("expected 'includeWhen', got %q", node.Variant)
	}

	if node.Condition != "$condition" {
		t.Errorf("expected '$condition', got %q", node.Condition)
	}
}

func TestParser_Push(t *testing.T) {
	ast := parseTemplate(t, "@push('scripts')<script>alert('hi')</script>@endpush")

	node, ok := ast.Children[0].(*PushNode)
	if !ok {
		t.Fatal("expected PushNode")
	}

	if node.Stack != "scripts" {
		t.Errorf("expected 'scripts', got %q", node.Stack)
	}
}

func TestParser_Component(t *testing.T) {
	ast := parseTemplate(t, "@component('alert')Message@slot('title')Title@endslot@endcomponent")

	node, ok := ast.Children[0].(*ComponentNode)
	if !ok {
		t.Fatal("expected ComponentNode")
	}

	if node.Name != "alert" {
		t.Errorf("expected 'alert', got %q", node.Name)
	}

	if _, ok := node.Slots["title"]; !ok {
		t.Error("expected 'title' slot")
	}
}

func TestParser_Auth(t *testing.T) {
	ast := parseTemplate(t, "@auth Logged in @endauth")

	node, ok := ast.Children[0].(*AuthNode)
	if !ok {
		t.Fatal("expected AuthNode")
	}

	if len(node.Children) == 0 {
		t.Error("expected children")
	}
}

func TestParser_Guest(t *testing.T) {
	ast := parseTemplate(t, "@guest Not logged in @endguest")

	_, ok := ast.Children[0].(*GuestNode)
	if !ok {
		t.Fatal("expected GuestNode")
	}
}

func TestParser_Env(t *testing.T) {
	ast := parseTemplate(t, "@env('local')Debug@endenv")

	node, ok := ast.Children[0].(*EnvNode)
	if !ok {
		t.Fatal("expected EnvNode")
	}

	if len(node.Environments) != 1 || node.Environments[0] != "local" {
		t.Errorf("expected ['local'], got %v", node.Environments)
	}
}

func TestParser_Error(t *testing.T) {
	ast := parseTemplate(t, "@error('email'){{ $message }}@enderror")

	node, ok := ast.Children[0].(*ErrorNode)
	if !ok {
		t.Fatal("expected ErrorNode")
	}

	if node.Field != "email" {
		t.Errorf("expected 'email', got %q", node.Field)
	}
}

func TestParser_Isset(t *testing.T) {
	ast := parseTemplate(t, "@isset($var)Variable is set@endisset")

	node, ok := ast.Children[0].(*IssetNode)
	if !ok {
		t.Fatal("expected IssetNode")
	}

	if node.Variable != "$var" {
		t.Errorf("expected '$var', got %q", node.Variable)
	}
}

func TestParser_Empty(t *testing.T) {
	ast := parseTemplate(t, "@empty($var)Variable is empty@endempty")

	_, ok := ast.Children[0].(*EmptyCheckNode)
	if !ok {
		t.Fatal("expected EmptyCheckNode")
	}
}

func TestParser_CSRF(t *testing.T) {
	ast := parseTemplate(t, "@csrf")

	node, ok := ast.Children[0].(*DirectiveNode)
	if !ok {
		t.Fatal("expected DirectiveNode")
	}

	if node.Name != "csrf" {
		t.Errorf("expected 'csrf', got %q", node.Name)
	}
}

func TestParser_Method(t *testing.T) {
	ast := parseTemplate(t, "@method('PUT')")

	node, ok := ast.Children[0].(*DirectiveNode)
	if !ok {
		t.Fatal("expected DirectiveNode")
	}

	if node.Name != "method" {
		t.Errorf("expected 'method', got %q", node.Name)
	}
}

func TestParser_ComplexTemplate(t *testing.T) {
	input := `@extends('layouts.app')

@section('content')
<div class="container">
    @if($users->count() > 0)
        @foreach($users as $user)
            <div class="user">
                <h2>{{ $user->name }}</h2>
                @if($user->isAdmin())
                    <span class="badge">Admin</span>
                @endif
            </div>
        @endforeach
    @else
        <p>No users found.</p>
    @endif
</div>
@endsection

@push('scripts')
<script>console.log('Users loaded');</script>
@endpush`

	ast := parseTemplate(t, input)

	// Count specific node types
	extendsCount := 0
	sectionCount := 0
	ifCount := 0
	foreachCount := 0
	pushCount := 0

	var countNodes func(nodes []Node)
	countNodes = func(nodes []Node) {
		for _, node := range nodes {
			switch n := node.(type) {
			case *ExtendsNode:
				extendsCount++
			case *SectionNode:
				sectionCount++
				countNodes(n.Children)
			case *IfNode:
				ifCount++
				countNodes(n.Children)
				for _, elif := range n.ElseIfs {
					countNodes(elif.Children)
				}
				if n.Else != nil {
					countNodes(n.Else.Children)
				}
			case *ForeachNode:
				foreachCount++
				countNodes(n.Children)
			case *PushNode:
				pushCount++
				countNodes(n.Children)
			}
		}
	}

	countNodes(ast.Children)

	if extendsCount != 1 {
		t.Errorf("expected 1 extends, got %d", extendsCount)
	}

	if sectionCount != 1 {
		t.Errorf("expected 1 section, got %d", sectionCount)
	}

	if ifCount != 2 { // Outer if and inner if for isAdmin
		t.Errorf("expected 2 ifs, got %d", ifCount)
	}

	if foreachCount != 1 {
		t.Errorf("expected 1 foreach, got %d", foreachCount)
	}

	if pushCount != 1 {
		t.Errorf("expected 1 push, got %d", pushCount)
	}
}
