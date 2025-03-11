package twig

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"
)

// bytesBufferPool is used to reuse byte buffers during template serialization
var bytesBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// getBuffer gets a bytes.Buffer from the pool
func getBuffer() *bytes.Buffer {
	buf := bytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// putBuffer returns a bytes.Buffer to the pool
func putBuffer(buf *bytes.Buffer) {
	bytesBufferPool.Put(buf)
}

func init() {
	// Register all node types for serialization
	// This is only needed for the AST serialization which still uses gob
	gob.Register(&RootNode{})
	gob.Register(&TextNode{})
	gob.Register(&PrintNode{})
	gob.Register(&BlockNode{})
	gob.Register(&ForNode{})
	gob.Register(&IfNode{})
	gob.Register(&SetNode{})
	gob.Register(&IncludeNode{})
	gob.Register(&ExtendsNode{})
	gob.Register(&MacroNode{})
	gob.Register(&ImportNode{})
	gob.Register(&FromImportNode{})
	gob.Register(&FunctionNode{})
	gob.Register(&CommentNode{})

	// Expression nodes
	gob.Register(&ExpressionNode{})
	gob.Register(&VariableNode{})
	gob.Register(&LiteralNode{})
	gob.Register(&UnaryNode{})
	gob.Register(&BinaryNode{})
}

// CompiledTemplate represents a compiled Twig template
type CompiledTemplate struct {
	Name         string // Template name
	Source       string // Original template source
	LastModified int64  // Last modification timestamp
	CompileTime  int64  // Time when compilation occurred
	AST          []byte // Serialized AST data
}

// Size returns the approximate memory size of the compiled template in bytes
func (c *CompiledTemplate) Size() int {
	if c == nil {
		return 0
	}

	// Calculate approximate size
	size := 0
	size += len(c.Name)
	size += len(c.Source)
	size += len(c.AST)
	size += 16 // Size of int64 fields

	return size
}

// CompileTemplate compiles a parsed template into a compiled format
func CompileTemplate(tmpl *Template) (*CompiledTemplate, error) {
	if tmpl == nil {
		return nil, fmt.Errorf("%w: cannot compile nil template", ErrCompilation)
	}

	// Serialize the AST (Node tree)
	var astBuf bytes.Buffer
	enc := gob.NewEncoder(&astBuf)

	// Encode the AST
	if err := enc.Encode(tmpl.nodes); err != nil {
		// If serialization fails, we'll still create the template but without AST
		// We continue silently to allow template creation
	}

	// Store the template source, metadata, and AST
	compiled := &CompiledTemplate{
		Name:         tmpl.name,
		Source:       tmpl.source,
		LastModified: tmpl.lastModified,
		CompileTime:  time.Now().Unix(),
		AST:          astBuf.Bytes(),
	}

	return compiled, nil
}

// LoadFromCompiled loads a template from its compiled representation
func LoadFromCompiled(compiled *CompiledTemplate, env *Environment, engine *Engine) (*Template, error) {
	if compiled == nil {
		return nil, fmt.Errorf("%w: cannot load from nil compiled template", ErrCompilation)
	}

	var nodes Node

	// Try to use the cached AST if available
	if len(compiled.AST) > 0 {
		// Attempt to deserialize the AST
		dec := gob.NewDecoder(bytes.NewReader(compiled.AST))

		// Try to decode the AST
		err := dec.Decode(&nodes)
		if err != nil {
			// Fall back to parsing if AST deserialization fails
			// We continue silently and fall back to parsing
			nodes = nil
		}
	}

	// If AST deserialization failed or AST is not available, parse the source
	if nodes == nil {
		parser := &Parser{}
		var err error
		nodes, err = parser.Parse(compiled.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse compiled template: %w", err)
		}
	}

	// Create the template from the nodes (either from AST or freshly parsed)
	tmpl := &Template{
		name:         compiled.Name,
		source:       compiled.Source,
		nodes:        nodes,
		env:          env,
		engine:       engine,
		lastModified: compiled.LastModified,
	}

	return tmpl, nil
}

// writeString writes a string to a buffer with length prefix
func writeString(w io.Writer, s string) error {
	// Write the string length as uint32
	if err := binary.Write(w, binary.LittleEndian, uint32(len(s))); err != nil {
		return err
	}

	// Write the string data
	_, err := w.Write([]byte(s))
	return err
}

// readString reads a length-prefixed string from a reader
func readString(r io.Reader) (string, error) {
	// Read string length
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", err
	}

	// Read string data
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}

	return string(data), nil
}

// SerializeCompiledTemplate serializes a compiled template to a byte array
func SerializeCompiledTemplate(compiled *CompiledTemplate) ([]byte, error) {
	// Get a buffer from the pool
	buf := getBuffer()
	defer putBuffer(buf)

	// Use binary encoding for metadata (more efficient than gob)
	// Write format version (for future compatibility)
	if err := binary.Write(buf, binary.LittleEndian, uint8(1)); err != nil {
		return nil, fmt.Errorf("failed to serialize version: %w", err)
	}

	// Write Name as length-prefixed string
	if err := writeString(buf, compiled.Name); err != nil {
		return nil, fmt.Errorf("failed to serialize name: %w", err)
	}

	// Write Source as length-prefixed string
	if err := writeString(buf, compiled.Source); err != nil {
		return nil, fmt.Errorf("failed to serialize source: %w", err)
	}

	// Write timestamps
	if err := binary.Write(buf, binary.LittleEndian, compiled.LastModified); err != nil {
		return nil, fmt.Errorf("failed to serialize LastModified: %w", err)
	}

	if err := binary.Write(buf, binary.LittleEndian, compiled.CompileTime); err != nil {
		return nil, fmt.Errorf("failed to serialize CompileTime: %w", err)
	}

	// Write AST data length followed by data
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(compiled.AST))); err != nil {
		return nil, fmt.Errorf("failed to serialize AST length: %w", err)
	}

	if _, err := buf.Write(compiled.AST); err != nil {
		return nil, fmt.Errorf("failed to serialize AST data: %w", err)
	}

	// Return a copy of the buffer data
	return bytes.Clone(buf.Bytes()), nil
}

// DeserializeCompiledTemplate deserializes a compiled template from a byte array
func DeserializeCompiledTemplate(data []byte) (*CompiledTemplate, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data cannot be deserialized")
	}

	// Try the new binary format first
	compiled, err := deserializeBinaryFormat(data)
	if err == nil {
		return compiled, nil
	}

	// Fall back to the old gob format if binary deserialization fails
	// This ensures backward compatibility with previously compiled templates
	return deserializeGobFormat(data)
}

// deserializeBinaryFormat deserializes using the new binary format
func deserializeBinaryFormat(data []byte) (*CompiledTemplate, error) {
	// Create a reader for the data
	r := bytes.NewReader(data)

	// Read and verify format version
	var version uint8
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return nil, fmt.Errorf("failed to read format version: %w", err)
	}

	if version != 1 {
		return nil, fmt.Errorf("unsupported format version: %d", version)
	}

	// Create a new compiled template
	compiled := new(CompiledTemplate)

	// Read Name
	var err error
	compiled.Name, err = readString(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read name: %w", err)
	}

	// Read Source
	compiled.Source, err = readString(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// Read timestamps
	if err := binary.Read(r, binary.LittleEndian, &compiled.LastModified); err != nil {
		return nil, fmt.Errorf("failed to read LastModified: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &compiled.CompileTime); err != nil {
		return nil, fmt.Errorf("failed to read CompileTime: %w", err)
	}

	// Read AST data length and data
	var astLength uint32
	if err := binary.Read(r, binary.LittleEndian, &astLength); err != nil {
		return nil, fmt.Errorf("failed to read AST length: %w", err)
	}

	compiled.AST = make([]byte, astLength)
	if _, err := io.ReadFull(r, compiled.AST); err != nil {
		return nil, fmt.Errorf("failed to read AST data: %w", err)
	}

	return compiled, nil
}

// deserializeGobFormat deserializes using the old gob format
func deserializeGobFormat(data []byte) (*CompiledTemplate, error) {
	dec := gob.NewDecoder(bytes.NewReader(data))

	var compiled CompiledTemplate
	if err := dec.Decode(&compiled); err != nil {
		return nil, fmt.Errorf("failed to deserialize compiled template: %w", err)
	}

	return &compiled, nil
}
