package engine

import (
	"crypto/md5"
	"encoding/hex"
	"html/template"
	"os"
	"sync"
	"time"
)

// CachedTemplate represents a compiled and cached template
type CachedTemplate struct {
	Template *template.Template
	ModTime  time.Time
	Checksum string
}

// TemplateCache manages template caching
type TemplateCache struct {
	templates map[string]*CachedTemplate
	mu        sync.RWMutex
	disabled  bool
}

// NewTemplateCache creates a new template cache
func NewTemplateCache() *TemplateCache {
	return &TemplateCache{
		templates: make(map[string]*CachedTemplate),
		disabled:  false,
	}
}

// Get retrieves a cached template if it exists and is valid
func (c *TemplateCache) Get(name string) (*CachedTemplate, bool) {
	if c.disabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.templates[name]
	return cached, ok
}

// Set stores a template in the cache
func (c *TemplateCache) Set(name string, tmpl *template.Template, modTime time.Time, checksum string) {
	if c.disabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.templates[name] = &CachedTemplate{
		Template: tmpl,
		ModTime:  modTime,
		Checksum: checksum,
	}
}

// Delete removes a template from the cache
func (c *TemplateCache) Delete(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.templates, name)
}

// Clear removes all templates from the cache
func (c *TemplateCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.templates = make(map[string]*CachedTemplate)
}

// Disable disables caching
func (c *TemplateCache) Disable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.disabled = true
}

// Enable enables caching
func (c *TemplateCache) Enable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.disabled = false
}

// IsValid checks if a cached template is still valid
// Returns false if the file has been modified since caching
func (c *TemplateCache) IsValid(name, filePath string) bool {
	if c.disabled {
		return false
	}

	cached, ok := c.Get(name)
	if !ok {
		return false
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return !info.ModTime().After(cached.ModTime)
}

// Checksum calculates MD5 checksum of content
func Checksum(content []byte) string {
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:])
}

// Size returns the number of cached templates
func (c *TemplateCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.templates)
}

// Names returns all cached template names
func (c *TemplateCache) Names() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.templates))
	for name := range c.templates {
		names = append(names, name)
	}
	return names
}
