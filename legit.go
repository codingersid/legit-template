// Package legitview provides a Laravel Blade-like template engine for Go.
//
// Legit View is a powerful template engine that replicates Laravel Blade's features
// while being fully compatible with Go and the Fiber web framework.
//
// # Basic Usage
//
//	// Create a new engine
//	engine := legitview.New("./resources/views")
//
//	// Render a template
//	result, err := engine.RenderString("pages.home", map[string]interface{}{
//	    "title": "Welcome",
//	    "user":  user,
//	})
//
// # With Fiber
//
//	import "github.com/codingersid/legit-template/fiber"
//
//	engine := fiber.New("./resources/views")
//	app := fiber.New(fiber.Config{
//	    Views: engine,
//	})
//
// # Template Syntax
//
// Legit View supports the following Blade-like syntax:
//
//   - {{ $variable }} - Escaped output
//   - {!! $variable !!} - Raw/unescaped output
//   - {{-- comment --}} - Comments (not rendered)
//   - @if($condition)...@endif - Conditionals
//   - @foreach($items as $item)...@endforeach - Loops
//   - @extends('layout') - Template inheritance
//   - @section('content')...@endsection - Sections
//   - @yield('content') - Section placeholders
//   - @include('partial') - Include partials
//   - @component('alert')...@endcomponent - Components
//   - And many more...
//
// See the documentation for a complete list of supported directives.
package legitview

import (
	"html/template"
	"io"

	"github.com/codingersid/legit-template/engine"
	fiberAdapter "github.com/codingersid/legit-template/fiber"
)

// Version is the current version of legit-view
const Version = "1.0.0"

// Engine is an alias for engine.Engine
type Engine = engine.Engine

// Option is an alias for engine.Option
type Option = engine.Option

// New creates a new template engine
//
// Example:
//
//	engine := legitview.New("./resources/views")
//	engine := legitview.New("./views", legitview.WithExtension(".html"))
func New(viewsPath string, opts ...Option) *Engine {
	return engine.New(viewsPath, opts...)
}

// NewFiber creates a new Fiber-compatible template engine
//
// Example:
//
//	engine := legitview.NewFiber("./resources/views")
//	app := fiber.New(fiber.Config{
//	    Views: engine,
//	})
func NewFiber(directory string, extension ...string) *fiberAdapter.Engine {
	return fiberAdapter.New(directory, extension...)
}

// WithExtension sets the template file extension (default: .legit)
func WithExtension(ext string) Option {
	return engine.WithExtension(ext)
}

// WithDevelopment enables development mode (disables caching)
func WithDevelopment(dev bool) Option {
	return engine.WithDevelopment(dev)
}

// WithFunctions adds custom template functions
func WithFunctions(funcs template.FuncMap) Option {
	return engine.WithFunctions(funcs)
}

// Render is a convenience function that creates an engine and renders a template
func Render(w io.Writer, viewsPath, name string, data interface{}) error {
	eng := New(viewsPath)
	return eng.Render(w, name, data)
}

// RenderString is a convenience function that creates an engine and renders a template to string
func RenderString(viewsPath, name string, data interface{}) (string, error) {
	eng := New(viewsPath)
	return eng.RenderString(name, data)
}

// DefaultFunctions returns the default template functions available in all templates
func DefaultFunctions() template.FuncMap {
	return engine.DefaultFunctions()
}

// Directives lists all supported directives
var Directives = []string{
	// Output
	"{{ }}",     // Escaped output
	"{!! !!}",   // Raw output
	"{{-- --}}", // Comments

	// Template Inheritance
	"@extends",
	"@section",
	"@endsection",
	"@show",
	"@yield",
	"@parent",

	// Includes
	"@include",
	"@includeIf",
	"@includeWhen",
	"@includeUnless",
	"@includeFirst",
	"@each",

	// Conditionals
	"@if",
	"@elseif",
	"@else",
	"@endif",
	"@unless",
	"@endunless",
	"@isset",
	"@endisset",
	"@empty",
	"@endempty",
	"@switch",
	"@case",
	"@break",
	"@default",
	"@endswitch",

	// Loops
	"@for",
	"@endfor",
	"@foreach",
	"@endforeach",
	"@forelse",
	"@endforelse",
	"@while",
	"@endwhile",
	"@continue",

	// Authentication
	"@auth",
	"@endauth",
	"@guest",
	"@endguest",

	// Environment
	"@env",
	"@endenv",
	"@production",
	"@endproduction",

	// Stacks
	"@push",
	"@endpush",
	"@prepend",
	"@endprepend",
	"@pushOnce",
	"@endPushOnce",
	"@stack",

	// Components
	"@component",
	"@endcomponent",
	"@slot",
	"@endslot",

	// Forms
	"@csrf",
	"@method",
	"@error",
	"@enderror",
	"@old",

	// Attributes
	"@class",
	"@style",
	"@checked",
	"@selected",
	"@disabled",
	"@readonly",
	"@required",

	// Miscellaneous
	"@json",
	"@verbatim",
	"@endverbatim",
	"@php",
	"@endphp",
	"@once",
	"@endonce",
}

// Functions lists all built-in template functions
var Functions = []string{
	// String
	"upper", "lower", "title", "trim", "ltrim", "rtrim",
	"replace", "contains", "hasPrefix", "hasSuffix",
	"split", "join", "repeat", "substr", "length",
	"nl2br", "ucfirst", "lcfirst", "slug", "limit", "wordLimit",

	// HTML
	"html", "htmlAttr", "js", "url",
	"safeHTML", "safeJS", "safeURL", "safeCSS",

	// Array/Slice
	"first", "last", "reverse", "sortAsc", "sortDesc",
	"unique", "pluck", "where", "groupBy", "chunk",
	"flatten", "slice", "append", "prepend", "merge",

	// Map
	"dict", "set", "unset", "keys", "values", "hasKey",

	// Number
	"add", "sub", "mul", "div", "mod",
	"round", "floor", "ceil", "abs",
	"min", "max", "currency", "number", "percent",

	// Date
	"date", "now", "ago", "diff", "addDate", "subDate", "timestamp",

	// Comparison
	"eq", "ne", "lt", "gt", "lte", "gte", "and", "or", "not",

	// Utility
	"default", "isset", "empty", "dump", "json", "jsonDec",
	"seq", "until", "index", "printf", "print",
	"coalesce", "ternary", "typeof",
	"toInt", "toFloat", "toString", "toBool",

	// Loop
	"newLoop",

	// Validation
	"hasError", "getError",

	// Class/Style
	"classArray", "styleArray",
}
