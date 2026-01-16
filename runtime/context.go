package runtime

import (
	"sync"
)

// Context holds the data for a template render
type Context struct {
	data     map[string]interface{}
	stacks   map[string][]string
	sections map[string]string
	errors   map[string][]string
	old      map[string]string
	mu       sync.RWMutex
}

// NewContext creates a new render context
func NewContext() *Context {
	return &Context{
		data:     make(map[string]interface{}),
		stacks:   make(map[string][]string),
		sections: make(map[string]string),
		errors:   make(map[string][]string),
		old:      make(map[string]string),
	}
}

// Set sets a value in the context
func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Get gets a value from the context
func (c *Context) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

// Has checks if a key exists in the context
func (c *Context) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.data[key]
	return ok
}

// Merge merges additional data into the context
func (c *Context) Merge(data map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range data {
		c.data[k] = v
	}
}

// Data returns all data as a map
func (c *Context) Data() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a copy
	result := make(map[string]interface{}, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// Stack operations

// PushStack pushes content to a named stack
func (c *Context) PushStack(name, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stacks[name] = append(c.stacks[name], content)
}

// PrependStack prepends content to a named stack
func (c *Context) PrependStack(name, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stacks[name] = append([]string{content}, c.stacks[name]...)
}

// GetStack returns all content for a named stack
func (c *Context) GetStack(name string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stacks[name]
}

// Section operations

// SetSection sets content for a named section
func (c *Context) SetSection(name, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sections[name] = content
}

// GetSection gets content for a named section
func (c *Context) GetSection(name string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sections[name]
}

// HasSection checks if a section exists
func (c *Context) HasSection(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.sections[name]
	return ok
}

// Validation errors

// SetErrors sets validation errors
func (c *Context) SetErrors(errors map[string][]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errors = errors
}

// GetErrors returns all validation errors
func (c *Context) GetErrors() map[string][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.errors
}

// HasError checks if a field has an error
func (c *Context) HasError(field string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	errors, ok := c.errors[field]
	return ok && len(errors) > 0
}

// GetError returns the first error for a field
func (c *Context) GetError(field string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if errors, ok := c.errors[field]; ok && len(errors) > 0 {
		return errors[0]
	}
	return ""
}

// Old input

// SetOld sets old input values
func (c *Context) SetOld(old map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.old = old
}

// GetOld returns old input for a field
func (c *Context) GetOld(field string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.old[field]
}

// Clone creates a copy of the context
func (c *Context) Clone() *Context {
	c.mu.RLock()
	defer c.mu.RUnlock()

	newCtx := NewContext()

	for k, v := range c.data {
		newCtx.data[k] = v
	}
	for k, v := range c.stacks {
		newCtx.stacks[k] = append([]string(nil), v...)
	}
	for k, v := range c.sections {
		newCtx.sections[k] = v
	}
	for k, v := range c.errors {
		newCtx.errors[k] = append([]string(nil), v...)
	}
	for k, v := range c.old {
		newCtx.old[k] = v
	}

	return newCtx
}

// SharedData holds data that is shared across all templates
type SharedData struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewSharedData creates new shared data
func NewSharedData() *SharedData {
	return &SharedData{
		data: make(map[string]interface{}),
	}
}

// Set sets a shared value
func (s *SharedData) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get gets a shared value
func (s *SharedData) Get(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

// All returns all shared data
func (s *SharedData) All() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}
