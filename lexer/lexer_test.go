package lexer

import (
	"testing"
)

func TestLexer_Text(t *testing.T) {
	input := "Hello World"
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 2 { // TEXT + EOF
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != TOKEN_TEXT {
		t.Errorf("expected TEXT token, got %s", tokens[0].Type)
	}

	if tokens[0].Value != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", tokens[0].Value)
	}
}

func TestLexer_EscapedEcho(t *testing.T) {
	input := "Hello {{ $name }}"
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 3 { // TEXT + ECHO_ESCAPED + EOF
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}

	if tokens[1].Type != TOKEN_ECHO_ESCAPED {
		t.Errorf("expected ECHO_ESCAPED token, got %s", tokens[1].Type)
	}

	if tokens[1].Value != "$name" {
		t.Errorf("expected '$name', got %q", tokens[1].Value)
	}
}

func TestLexer_RawEcho(t *testing.T) {
	input := "{!! $html !!}"
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 2 { // ECHO_RAW + EOF
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != TOKEN_ECHO_RAW {
		t.Errorf("expected ECHO_RAW token, got %s", tokens[0].Type)
	}

	if tokens[0].Value != "$html" {
		t.Errorf("expected '$html', got %q", tokens[0].Value)
	}
}

func TestLexer_Comment(t *testing.T) {
	input := "{{-- This is a comment --}}"
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 2 { // COMMENT + EOF
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != TOKEN_COMMENT {
		t.Errorf("expected COMMENT token, got %s", tokens[0].Type)
	}

	if tokens[0].Value != "This is a comment" {
		t.Errorf("expected 'This is a comment', got %q", tokens[0].Value)
	}
}

func TestLexer_Directive(t *testing.T) {
	input := "@if"
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 2 { // DIRECTIVE + EOF
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != TOKEN_DIRECTIVE {
		t.Errorf("expected DIRECTIVE token, got %s", tokens[0].Type)
	}

	if tokens[0].Value != "if" {
		t.Errorf("expected 'if', got %q", tokens[0].Value)
	}
}

func TestLexer_DirectiveWithArgs(t *testing.T) {
	input := "@if($condition)"
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 2 { // DIRECTIVE_ARGS + EOF
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != TOKEN_DIRECTIVE_ARGS {
		t.Errorf("expected DIRECTIVE_ARGS token, got %s", tokens[0].Type)
	}

	if tokens[0].Value != "if" {
		t.Errorf("expected 'if', got %q", tokens[0].Value)
	}

	if tokens[0].Args != "$condition" {
		t.Errorf("expected '$condition', got %q", tokens[0].Args)
	}
}

func TestLexer_DirectiveNestedParens(t *testing.T) {
	input := "@if(func($var, array('a', 'b')))"
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Args != "func($var, array('a', 'b'))" {
		t.Errorf("expected nested args, got %q", tokens[0].Args)
	}
}

func TestLexer_EscapedAt(t *testing.T) {
	input := "@@if"
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 3 { // TEXT (@) + TEXT (if) + EOF
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != TOKEN_TEXT {
		t.Errorf("expected TEXT token, got %s", tokens[0].Type)
	}

	if tokens[0].Value != "@" {
		t.Errorf("expected '@', got %q", tokens[0].Value)
	}
}

func TestLexer_Verbatim(t *testing.T) {
	input := "@verbatim{{ $notParsed }}@endverbatim"
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: VERBATIM_START + TEXT + VERBATIM_END + EOF
	hasVerbatimStart := false
	hasVerbatimEnd := false
	hasText := false

	for _, tok := range tokens {
		switch tok.Type {
		case TOKEN_VERBATIM_START:
			hasVerbatimStart = true
		case TOKEN_VERBATIM_END:
			hasVerbatimEnd = true
		case TOKEN_TEXT:
			if tok.Value == "{{ $notParsed }}" {
				hasText = true
			}
		}
	}

	if !hasVerbatimStart {
		t.Error("expected VERBATIM_START token")
	}
	if !hasVerbatimEnd {
		t.Error("expected VERBATIM_END token")
	}
	if !hasText {
		t.Error("expected TEXT token with verbatim content")
	}
}

func TestLexer_ComplexTemplate(t *testing.T) {
	input := `@extends('layouts.app')

@section('content')
<div class="container">
    @if($users)
        @foreach($users as $user)
            <p>{{ $user->name }}</p>
        @endforeach
    @else
        <p>No users found</p>
    @endif
</div>
@endsection`

	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count directive tokens
	directiveCount := 0
	for _, tok := range tokens {
		if tok.Type == TOKEN_DIRECTIVE || tok.Type == TOKEN_DIRECTIVE_ARGS {
			directiveCount++
		}
	}

	// Should have: extends, section, if, foreach, endforeach, else, endif, endsection = 8
	if directiveCount != 8 {
		t.Errorf("expected 8 directive tokens, got %d", directiveCount)
	}
}

func TestLexer_Position(t *testing.T) {
	input := "Line1\n@if($x)\nLine3"
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find the @if directive
	for _, tok := range tokens {
		if tok.Type == TOKEN_DIRECTIVE_ARGS && tok.Value == "if" {
			if tok.Position.Line != 2 {
				t.Errorf("expected line 2, got %d", tok.Position.Line)
			}
			break
		}
	}
}

func TestLexer_UnclosedEcho(t *testing.T) {
	input := "{{ $unclosed"
	lex := New(input)
	_, err := lex.Tokenize()

	if err == nil {
		t.Error("expected error for unclosed echo")
	}
}

func TestLexer_UnclosedComment(t *testing.T) {
	input := "{{-- unclosed"
	lex := New(input)
	_, err := lex.Tokenize()

	if err == nil {
		t.Error("expected error for unclosed comment")
	}
}

func TestLexer_StringsInArgs(t *testing.T) {
	input := `@include('partials.header', ['title' => 'Test'])`
	lex := New(input)
	tokens, err := lex.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Args != "'partials.header', ['title' => 'Test']" {
		t.Errorf("unexpected args: %q", tokens[0].Args)
	}
}
