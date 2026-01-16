package fiber

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/codingersid/legit-template/engine"
)

// Engine wraps the legit-view engine for Fiber compatibility
type Engine struct {
	*engine.Engine
	directory  string
	extension  string
	layout     string
	reload     bool
	debug      bool
	mutex      sync.RWMutex
	layoutFunc func() string
}

// New creates a new Fiber-compatible template engine
func New(directory string, extension ...string) *Engine {
	ext := ".legit"
	if len(extension) > 0 {
		ext = extension[0]
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
	}

	eng := engine.New(directory,
		engine.WithExtension(ext),
	)

	return &Engine{
		Engine:    eng,
		directory: directory,
		extension: ext,
		reload:    false,
		debug:     false,
	}
}

// NewFiber creates a new Fiber-compatible template engine (alias for New)
func NewFiber(directory string, extension ...string) *Engine {
	return New(directory, extension...)
}

// Layout sets the default layout template
func (e *Engine) Layout(layout string) *Engine {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.layout = layout
	return e
}

// LayoutFunc sets a function that returns the layout template name
func (e *Engine) LayoutFunc(fn func() string) *Engine {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.layoutFunc = fn
	return e
}

// Reload enables reloading of templates on each request (development mode)
func (e *Engine) Reload(reload bool) *Engine {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.reload = reload
	if reload {
		e.ClearCache()
	}
	return e
}

// Debug enables debug mode with error details
func (e *Engine) Debug(debug bool) *Engine {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.debug = debug
	return e
}

// Load pre-compiles all templates
// This implements the fiber.Views interface
func (e *Engine) Load() error {
	if e.reload {
		return nil // Don't pre-load in reload mode
	}

	return filepath.Walk(e.directory, func(path string, info os.FileInfo, err error) error {
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
		name := strings.TrimPrefix(path, e.directory+string(filepath.Separator))
		name = strings.TrimSuffix(name, e.extension)
		name = strings.ReplaceAll(name, string(filepath.Separator), "/")

		// Compile template by rendering with nil data
		// This validates the template and caches it
		_, err = e.Engine.RenderString(name, nil)
		if err != nil && e.debug {
			fmt.Printf("Warning: failed to pre-compile template %s: %v\n", name, err)
		}
		return nil
	})
}

// Render renders a template with the given data
// This implements the fiber.Views interface
func (e *Engine) Render(w io.Writer, name string, data interface{}, layouts ...string) error {
	// Clear cache in reload mode
	if e.reload {
		e.ClearCache()
	}

	// Prepare binding data
	binding := e.prepareBinding(data)

	// Determine layout to use
	layout := e.getLayout(layouts...)

	// If layout is specified, render the view into the layout
	if layout != "" {
		return e.renderWithLayout(w, name, layout, binding)
	}

	// Direct render
	return e.Engine.Render(w, name, binding)
}

// renderWithLayout renders a template with a layout
func (e *Engine) renderWithLayout(w io.Writer, name, layout string, binding map[string]interface{}) error {
	// First render the content template
	content, err := e.Engine.RenderString(name, binding)
	if err != nil {
		return err
	}

	// Add content to binding
	binding["Content"] = content
	binding["LayoutContent"] = content

	// Render the layout
	return e.Engine.Render(w, layout, binding)
}

// prepareBinding converts data to map[string]interface{}
func (e *Engine) prepareBinding(data interface{}) map[string]interface{} {
	if data == nil {
		return make(map[string]interface{})
	}

	switch d := data.(type) {
	case map[string]interface{}:
		return d
	case map[string]string:
		result := make(map[string]interface{}, len(d))
		for k, v := range d {
			result[k] = v
		}
		return result
	default:
		return map[string]interface{}{"data": data}
	}
}

// getLayout determines which layout to use
func (e *Engine) getLayout(layouts ...string) string {
	// Use layout from Render call if provided
	if len(layouts) > 0 && layouts[0] != "" {
		return layouts[0]
	}

	// Use layout function if set
	if e.layoutFunc != nil {
		return e.layoutFunc()
	}

	// Use default layout
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.layout
}

// FuncMap returns the template function map
func (e *Engine) FuncMap() map[string]interface{} {
	return engine.DefaultFunctions()
}

// AddFunc adds a custom template function
func (e *Engine) AddFunc(name string, fn interface{}) *Engine {
	e.Engine.AddFunction(name, fn)
	return e
}

// AddFuncMap adds multiple template functions
func (e *Engine) AddFuncMap(funcs map[string]interface{}) *Engine {
	for name, fn := range funcs {
		e.Engine.AddFunction(name, fn)
	}
	return e
}

// Delims is a no-op for compatibility with other engines
// Legit template uses {{ }} and {!! !!} delimiters which cannot be changed
func (e *Engine) Delims(left, right string) *Engine {
	// No-op: Legit template delimiters are fixed
	return e
}

// HTTPHandler returns an http.Handler that renders the template
func (e *Engine) HTTPHandler(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := e.Engine.Render(w, name, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// Templates returns all available template names
func (e *Engine) Templates() []string {
	templates, _ := e.Engine.Templates()
	return templates
}

// Options for the engine

// WithLayout sets the default layout
func WithLayout(layout string) func(*Engine) {
	return func(e *Engine) {
		e.layout = layout
	}
}

// WithReload enables reload mode
func WithReload(reload bool) func(*Engine) {
	return func(e *Engine) {
		e.reload = reload
	}
}

// WithDebug enables debug mode
func WithDebug(debug bool) func(*Engine) {
	return func(e *Engine) {
		e.debug = debug
	}
}

// NewWithOptions creates a new engine with options
func NewWithOptions(directory string, extension string, opts ...func(*Engine)) *Engine {
	e := New(directory, extension)
	for _, opt := range opts {
		opt(e)
	}
	return e
}
