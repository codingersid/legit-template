package engine

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/codingersid/legit-template/compiler"
	"github.com/codingersid/legit-template/lexer"
	"github.com/codingersid/legit-template/parser"
	"github.com/codingersid/legit-template/runtime"
)

// Engine is the main template engine
type Engine struct {
	viewsPath   string
	extension   string
	cache       *TemplateCache
	functions   template.FuncMap
	shared      *runtime.SharedData
	development bool
	mutex       sync.RWMutex

	// Custom directives
	directives map[string]DirectiveHandler
}

// DirectiveHandler is a function that handles custom directives
type DirectiveHandler func(args string, data map[string]interface{}) string

// Option configures the engine
type Option func(*Engine)

// New creates a new template engine
func New(viewsPath string, opts ...Option) *Engine {
	e := &Engine{
		viewsPath:   viewsPath,
		extension:   ".legit",
		cache:       NewTemplateCache(),
		functions:   DefaultFunctions(),
		shared:      runtime.NewSharedData(),
		development: false,
		directives:  make(map[string]DirectiveHandler),
	}

	for _, opt := range opts {
		opt(e)
	}

	if e.development {
		e.cache.Disable()
	}

	return e
}

// WithExtension sets the template file extension
func WithExtension(ext string) Option {
	return func(e *Engine) {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		e.extension = ext
	}
}

// WithDevelopment enables development mode (disables caching)
func WithDevelopment(dev bool) Option {
	return func(e *Engine) {
		e.development = dev
	}
}

// WithFunctions adds custom template functions
func WithFunctions(funcs template.FuncMap) Option {
	return func(e *Engine) {
		for name, fn := range funcs {
			e.functions[name] = fn
		}
	}
}

// AddFunction adds a custom template function
func (e *Engine) AddFunction(name string, fn interface{}) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.functions[name] = fn
}

// AddDirective adds a custom directive handler
func (e *Engine) AddDirective(name string, handler DirectiveHandler) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.directives[name] = handler
}

// Share adds data that will be available to all templates
func (e *Engine) Share(key string, value interface{}) {
	e.shared.Set(key, value)
}

// Render renders a template to the given writer
func (e *Engine) Render(w io.Writer, name string, data interface{}) error {
	tmpl, err := e.getTemplate(name)
	if err != nil {
		return err
	}

	// Prepare data
	renderData := e.prepareData(data)

	return tmpl.Execute(w, renderData)
}

// RenderString renders a template and returns the result as a string
func (e *Engine) RenderString(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := e.Render(&buf, name, data)
	return buf.String(), err
}

// RenderTemplate renders a template string directly (not from file)
func (e *Engine) RenderTemplate(templateStr string, data interface{}) (string, error) {
	compiled, err := e.compileString(templateStr)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("inline").Funcs(e.functions).Parse(compiled)
	if err != nil {
		return "", fmt.Errorf("failed to parse compiled template: %w", err)
	}

	renderData := e.prepareData(data)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, renderData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ClearCache clears the template cache
func (e *Engine) ClearCache() {
	e.cache.Clear()
}

// getTemplate retrieves or compiles a template
func (e *Engine) getTemplate(name string) (*template.Template, error) {
	filePath := e.resolvePath(name)

	// Check cache
	if cached, ok := e.cache.Get(name); ok {
		if e.cache.IsValid(name, filePath) {
			return cached.Template, nil
		}
	}

	// Compile template
	tmpl, modTime, err := e.compileFile(name, filePath)
	if err != nil {
		return nil, err
	}

	// Cache compiled template
	content, _ := os.ReadFile(filePath)
	e.cache.Set(name, tmpl, modTime, Checksum(content))

	return tmpl, nil
}

// compileFile compiles a template file
func (e *Engine) compileFile(name, filePath string) (*template.Template, time.Time, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to read template %s: %w", name, err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, time.Time{}, err
	}

	compiled, extendsTemplate, sections, err := e.compile(string(content))
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to compile template %s: %w", name, err)
	}

	// Handle template inheritance
	if extendsTemplate != "" {
		return e.compileWithInheritance(name, compiled, extendsTemplate, sections)
	}

	tmpl, err := template.New(name).Funcs(e.functions).Parse(compiled)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to parse compiled template %s: %w", name, err)
	}

	return tmpl, info.ModTime(), nil
}

// compileWithInheritance handles @extends directive
func (e *Engine) compileWithInheritance(name, childCompiled, parentName string, childSections map[string]string) (*template.Template, time.Time, error) {
	parentPath := e.resolvePath(parentName)
	parentContent, err := os.ReadFile(parentPath)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to read parent template %s: %w", parentName, err)
	}

	parentInfo, err := os.Stat(parentPath)
	if err != nil {
		return nil, time.Time{}, err
	}

	parentCompiled, parentExtends, parentSections, err := e.compile(string(parentContent))
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to compile parent template %s: %w", parentName, err)
	}

	// Merge sections (child overrides parent)
	for name, content := range parentSections {
		if _, ok := childSections[name]; !ok {
			childSections[name] = content
		}
	}

	// Replace @yield with section content
	for sectionName, sectionContent := range childSections {
		// Handle @parent directive
		if strings.Contains(sectionContent, "{{__PARENT__}}") {
			if parentContent, ok := parentSections[sectionName]; ok {
				sectionContent = strings.ReplaceAll(sectionContent, "{{__PARENT__}}", parentContent)
			} else {
				sectionContent = strings.ReplaceAll(sectionContent, "{{__PARENT__}}", "")
			}
		}

		// Replace {{ block "name" . }}...{{ end }} with section content
		blockStart := fmt.Sprintf(`{{ block "%s" . }}`, sectionName)
		blockEnd := `{{ end }}`

		startIdx := strings.Index(parentCompiled, blockStart)
		if startIdx != -1 {
			// Find the matching {{ end }}
			searchFrom := startIdx + len(blockStart)
			depth := 1
			endIdx := -1

			for i := searchFrom; i < len(parentCompiled); {
				if strings.HasPrefix(parentCompiled[i:], "{{ end }}") {
					depth--
					if depth == 0 {
						endIdx = i + len(blockEnd)
						break
					}
					i += len(blockEnd)
				} else if strings.HasPrefix(parentCompiled[i:], "{{ if ") ||
					strings.HasPrefix(parentCompiled[i:], "{{ range ") ||
					strings.HasPrefix(parentCompiled[i:], "{{ with ") ||
					strings.HasPrefix(parentCompiled[i:], "{{ block ") {
					depth++
					i++
				} else {
					i++
				}
			}

			if endIdx != -1 {
				parentCompiled = parentCompiled[:startIdx] + sectionContent + parentCompiled[endIdx:]
			}
		}
	}

	// If parent also extends another template, recurse
	if parentExtends != "" {
		return e.compileWithInheritance(name, parentCompiled, parentExtends, childSections)
	}

	tmpl, err := template.New(name).Funcs(e.functions).Parse(parentCompiled)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to parse compiled template %s: %w", name, err)
	}

	return tmpl, parentInfo.ModTime(), nil
}

// compile compiles template content
func (e *Engine) compile(content string) (string, string, map[string]string, error) {
	// Tokenize
	lex := lexer.New(content)
	tokens, err := lex.Tokenize()
	if err != nil {
		return "", "", nil, fmt.Errorf("lexer error: %w", err)
	}

	// Parse
	p := parser.New(tokens)
	ast, err := p.Parse()
	if err != nil {
		return "", "", nil, fmt.Errorf("parser error: %w", err)
	}

	// Compile
	c := compiler.New()
	compiled, err := c.Compile(ast)
	if err != nil {
		return "", "", nil, fmt.Errorf("compiler error: %w", err)
	}

	// Add stack function
	compiled = e.processStacks(compiled, c)

	return compiled, c.GetExtends(), c.GetSections(), nil
}

// compileString compiles a template string
func (e *Engine) compileString(content string) (string, error) {
	compiled, _, _, err := e.compile(content)
	return compiled, err
}

// processStacks replaces @stack placeholders with actual content
func (e *Engine) processStacks(compiled string, c *compiler.Compiler) string {
	// This is a simple implementation - real implementation would be more sophisticated
	// to handle runtime stack evaluation

	// Add stack function that returns empty string (stacks are evaluated at runtime)
	return compiled
}

// prepareData prepares the render data
func (e *Engine) prepareData(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Add shared data
	for k, v := range e.shared.All() {
		result[k] = v
	}

	// Merge provided data
	if data != nil {
		switch d := data.(type) {
		case map[string]interface{}:
			for k, v := range d {
				result[k] = v
			}
		case map[string]string:
			for k, v := range d {
				result[k] = v
			}
		}
	}

	// Add stack function
	result["__stacks"] = make(map[string][]string)

	return result
}

// resolvePath resolves template name to file path
func (e *Engine) resolvePath(name string) string {
	// Replace dots with path separator
	name = strings.ReplaceAll(name, ".", string(filepath.Separator))

	// Add extension if not present
	if !strings.HasSuffix(name, e.extension) {
		name = name + e.extension
	}

	return filepath.Join(e.viewsPath, name)
}

// Exists checks if a template exists
func (e *Engine) Exists(name string) bool {
	filePath := e.resolvePath(name)
	_, err := os.Stat(filePath)
	return err == nil
}

// Load pre-compiles all templates in the views directory
func (e *Engine) Load() error {
	return filepath.Walk(e.viewsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, e.extension) {
			return nil
		}

		// Get template name from path
		name := strings.TrimPrefix(path, e.viewsPath+string(filepath.Separator))
		name = strings.TrimSuffix(name, e.extension)
		name = strings.ReplaceAll(name, string(filepath.Separator), ".")

		// Compile and cache
		_, err = e.getTemplate(name)
		return err
	})
}

// Templates returns all available template names
func (e *Engine) Templates() ([]string, error) {
	var templates []string

	err := filepath.Walk(e.viewsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, e.extension) {
			return nil
		}

		name := strings.TrimPrefix(path, e.viewsPath+string(filepath.Separator))
		name = strings.TrimSuffix(name, e.extension)
		name = strings.ReplaceAll(name, string(filepath.Separator), ".")

		templates = append(templates, name)
		return nil
	})

	return templates, err
}

// EngineError represents a template engine error
type EngineError struct {
	Message  string
	Template string
	Line     int
	Column   int
	Near     string
}

func (e *EngineError) Error() string {
	if e.Template != "" {
		return fmt.Sprintf("%s in %s at line %d, column %d\n%s",
			e.Message, e.Template, e.Line, e.Column, e.Near)
	}
	return e.Message
}
