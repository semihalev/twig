package twig

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// DebugLevel represents the verbosity level of debugging
type DebugLevel int

const (
	// Debug levels
	DebugOff DebugLevel = iota
	DebugError
	DebugWarning
	DebugInfo
	DebugVerbose
)

// Debugger provides logging and debugging tools for the Twig engine
type Debugger struct {
	mu       sync.Mutex
	level    DebugLevel
	writer   io.Writer
	logger   *log.Logger
	traces   []string
	enabled  bool
	filename string
	line     int
}

// Global debugger instance
var debugger = &Debugger{
	level:   DebugOff,
	writer:  os.Stderr,
	logger:  log.New(os.Stderr, "[TWIG] ", log.LstdFlags),
	enabled: false,
}

// SetDebugLevel sets the global debug level
func SetDebugLevel(level DebugLevel) {
	debugger.mu.Lock()
	defer debugger.mu.Unlock()
	debugger.level = level
	debugger.enabled = level > DebugOff
}

// SetDebugWriter sets the output writer for debug messages
func SetDebugWriter(w io.Writer) {
	debugger.mu.Lock()
	defer debugger.mu.Unlock()
	debugger.writer = w
	debugger.logger = log.New(w, "[TWIG] ", log.LstdFlags)
}

// IsDebugEnabled returns true if debugging is enabled
func IsDebugEnabled() bool {
	return debugger.enabled
}

// LogError logs an error with source information
func LogError(err error, context ...string) {
	if debugger.level >= DebugError {
		_, file, line, _ := runtime.Caller(1)
		contextStr := ""
		if len(context) > 0 {
			contextStr = " " + strings.Join(context, " ")
		}
		debugger.logger.Printf("ERROR:%s:%d:%s%s", filepath.Base(file), line, err, contextStr)
	}
}

// LogWarning logs a warning with source information
func LogWarning(msg string, args ...interface{}) {
	if debugger.level >= DebugWarning {
		_, file, line, _ := runtime.Caller(1)
		debugger.logger.Printf("WARNING:%s:%d:%s", filepath.Base(file), line, fmt.Sprintf(msg, args...))
	}
}

// LogInfo logs an informational message
func LogInfo(msg string, args ...interface{}) {
	if debugger.level >= DebugInfo {
		debugger.logger.Printf("INFO:%s", fmt.Sprintf(msg, args...))
	}
}

// LogVerbose logs detailed information for debugging
func LogVerbose(msg string, args ...interface{}) {
	if debugger.level >= DebugVerbose {
		_, file, line, _ := runtime.Caller(1)
		debugger.logger.Printf("VERBOSE:%s:%d:%s", filepath.Base(file), line, fmt.Sprintf(msg, args...))
	}
}

// LogDebug logs debugging information when debug mode is enabled
func LogDebug(msg string, args ...interface{}) {
	if debugger.enabled {
		debugger.logger.Printf("DEBUG:%s", fmt.Sprintf(msg, args...))
	}
}

// StartTrace begins a trace of template rendering
func StartTrace(templateName string) func() {
	if !debugger.enabled {
		return func() {}
	}

	traceID := fmt.Sprintf("TRACE-%s-%d", templateName, time.Now().UnixNano())
	start := time.Now()

	debugger.mu.Lock()
	debugger.traces = append(debugger.traces, traceID)
	debugger.mu.Unlock()

	LogInfo("Begin rendering template: %s", templateName)

	return func() {
		elapsed := time.Since(start)
		LogInfo("Completed rendering template: %s (took %s)", templateName, elapsed)

		debugger.mu.Lock()
		defer debugger.mu.Unlock()

		// Remove this trace from active traces
		for i, t := range debugger.traces {
			if t == traceID {
				debugger.traces = append(debugger.traces[:i], debugger.traces[i+1:]...)
				break
			}
		}
	}
}

// TraceSection traces a section of template rendering
func TraceSection(name string) func() {
	if !debugger.enabled {
		return func() {}
	}

	start := time.Now()
	LogVerbose("Begin section: %s", name)

	return func() {
		elapsed := time.Since(start)
		LogVerbose("End section: %s (took %s)", name, elapsed)
	}
}

// DebugRender enables detailed rendering information
func DebugRender(w io.Writer, tmpl *Template, ctx *RenderContext) error {
	if !debugger.enabled {
		return tmpl.RenderTo(w, ctx.context)
	}

	LogInfo("Rendering template %s with context containing %d variables",
		tmpl.name, len(ctx.context))

	// Log context variables at verbose level
	if debugger.level >= DebugVerbose {
		for k, v := range ctx.context {
			typeName := "nil"
			if v != nil {
				typeName = fmt.Sprintf("%T", v)
			}
			LogVerbose("Context var: %s = %v (type: %s)", k, v, typeName)
		}
	}

	// Trace full template rendering
	defer StartTrace(tmpl.name)()

	return tmpl.RenderTo(w, ctx.context)
}

// FormatErrorContext creates a formatted context for syntax errors
// including the source line and position indicator
func FormatErrorContext(source string, position int, line int) string {
	if source == "" || position < 0 {
		return ""
	}

	lines := strings.Split(source, "\n")
	if line <= 0 || line > len(lines) {
		return ""
	}

	// Get the problematic line
	errorLine := lines[line-1]

	// Calculate column position within the line
	lineStartIdx := 0
	for i := 0; i < line-1; i++ {
		lineStartIdx += len(lines[i]) + 1 // +1 for the newline
	}
	colPosition := position - lineStartIdx

	// Ensure column position is valid
	if colPosition < 0 {
		colPosition = 0
	}
	if colPosition > len(errorLine) {
		colPosition = len(errorLine)
	}

	// Build the context output
	context := fmt.Sprintf("Line %d: %s\n", line, errorLine)
	if colPosition >= 0 {
		context += strings.Repeat(" ", colPosition+8) + "^\n"
	}

	return context
}

// EnhancedError provides more detailed error information for debugging
type EnhancedError struct {
	Err       error
	Template  string
	Line      int
	Column    int
	Source    string
	SourceCtx string
}

// Error implements the error interface
func (e *EnhancedError) Error() string {
	if e.Err == nil {
		return "unknown error"
	}

	location := ""
	if e.Template != "" {
		location = fmt.Sprintf("in template '%s' ", e.Template)
	}

	position := ""
	if e.Line > 0 {
		position = fmt.Sprintf("at line %d", e.Line)
		if e.Column > 0 {
			position += fmt.Sprintf(", column %d", e.Column)
		}
	}

	context := ""
	if e.SourceCtx != "" {
		context = "\n" + e.SourceCtx
	}

	return fmt.Sprintf("Error %s%s: %s%s", location, position, e.Err.Error(), context)
}

// Unwrap returns the underlying error
func (e *EnhancedError) Unwrap() error {
	return e.Err
}

// NewError creates an enhanced error with context
func NewError(err error, tmpl string, line int, col int, source string) error {
	if err == nil {
		return nil
	}

	e := &EnhancedError{
		Err:      err,
		Template: tmpl,
		Line:     line,
		Column:   col,
		Source:   source,
	}

	if source != "" && line > 0 {
		position := 0
		if col > 0 {
			// Calculate position from line and column
			lines := strings.Split(source, "\n")
			for i := 0; i < line-1 && i < len(lines); i++ {
				position += len(lines[i]) + 1 // +1 for newline
			}
			position += col - 1
		}
		e.SourceCtx = FormatErrorContext(source, position, line)
	}

	return e
}
