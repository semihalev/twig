package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/semihalev/twig"
)

// benchmarkSerialization compares the performance of old and new serialization methods
func benchmarkSerialization() {
	fmt.Println("\n=== Template Serialization Benchmark ===")
	
	// Create a more complex template to serialize
	engine := twig.New()
	source := `
{% extends "base.html" %}
{% block content %}
  <h1>{{ title }}</h1>
  <ul>
    {% for item in items %}
      <li>{{ item.name }} - {{ item.value|number_format(2, '.', ',') }}</li>
    {% endfor %}
  </ul>
  
  {% if showDetails %}
    <div class="details">
      {% include "details.html" with {'id': user.id} %}
    </div>
  {% endif %}
  
  <p>{{ footer|raw }}</p>
{% endblock %}
`
	engine.RegisterString("template", source)
	
	// Compile the template
	tmpl, _ := engine.Load("template")
	compiled, _ := tmpl.Compile()
	
	// Serialize using both methods
	oldData, _ := oldGobSerialize(compiled)
	newData, _ := twig.SerializeCompiledTemplate(compiled)
	
	// Size comparison
	fmt.Printf("Old format (gob) size: %d bytes\n", len(oldData))
	fmt.Printf("New format (binary) size: %d bytes\n", len(newData))
	fmt.Printf("Size reduction: %.2f%%\n\n", (1.0-float64(len(newData))/float64(len(oldData)))*100)
	
	// Benchmark serialization
	fmt.Println("Serialization Performance (1000 operations):")
	
	// Old method
	iterations := 1000
	startOldSer := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = oldGobSerialize(compiled)
	}
	oldSerTime := time.Since(startOldSer)
	
	// New method
	startNewSer := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = twig.SerializeCompiledTemplate(compiled)
	}
	newSerTime := time.Since(startNewSer)
	
	fmt.Printf("Old serialization: %v (%.2f µs/op)\n", oldSerTime, float64(oldSerTime.Nanoseconds())/float64(iterations)/1000)
	fmt.Printf("New serialization: %v (%.2f µs/op)\n", newSerTime, float64(newSerTime.Nanoseconds())/float64(iterations)/1000)
	fmt.Printf("Serialization speedup: %.2fx\n\n", float64(oldSerTime.Nanoseconds())/float64(newSerTime.Nanoseconds()))
	
	// Benchmark deserialization
	fmt.Println("Deserialization Performance (1000 operations):")
	
	// Old method
	startOldDeser := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = oldGobDeserialize(oldData)
	}
	oldDeserTime := time.Since(startOldDeser)
	
	// New method
	startNewDeser := time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = twig.DeserializeCompiledTemplate(newData)
	}
	newDeserTime := time.Since(startNewDeser)
	
	fmt.Printf("Old deserialization: %v (%.2f µs/op)\n", oldDeserTime, float64(oldDeserTime.Nanoseconds())/float64(iterations)/1000)
	fmt.Printf("New deserialization: %v (%.2f µs/op)\n", newDeserTime, float64(newDeserTime.Nanoseconds())/float64(iterations)/1000)
	fmt.Printf("Deserialization speedup: %.2fx\n\n", float64(oldDeserTime.Nanoseconds())/float64(newDeserTime.Nanoseconds()))
	
	// Total round-trip comparison
	fmt.Println("Round-trip Performance (1000 operations):")
	
	// Old method
	startOldTotal := time.Now()
	for i := 0; i < iterations; i++ {
		data, _ := oldGobSerialize(compiled)
		_, _ = oldGobDeserialize(data)
	}
	oldTotalTime := time.Since(startOldTotal)
	
	// New method
	startNewTotal := time.Now()
	for i := 0; i < iterations; i++ {
		data, _ := twig.SerializeCompiledTemplate(compiled)
		_, _ = twig.DeserializeCompiledTemplate(data)
	}
	newTotalTime := time.Since(startNewTotal)
	
	fmt.Printf("Old total: %v (%.2f µs/op)\n", oldTotalTime, float64(oldTotalTime.Nanoseconds())/float64(iterations)/1000)
	fmt.Printf("New total: %v (%.2f µs/op)\n", newTotalTime, float64(newTotalTime.Nanoseconds())/float64(iterations)/1000)
	fmt.Printf("Overall speedup: %.2fx\n\n", float64(oldTotalTime.Nanoseconds())/float64(newTotalTime.Nanoseconds()))
	
	// Memory usage estimation
	templateCount := 100
	fmt.Printf("Memory usage for %d templates:\n", templateCount)
	fmt.Printf("Old format: %.2f KB\n", float64(len(oldData)*templateCount)/1024)
	fmt.Printf("New format: %.2f KB\n", float64(len(newData)*templateCount)/1024)
	fmt.Printf("Memory saved: %.2f KB\n", float64(len(oldData)-len(newData))*float64(templateCount)/1024)
}

// oldGobSerialize simulates the old gob serialization
func oldGobSerialize(compiled *twig.CompiledTemplate) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	
	if err := enc.Encode(compiled); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// oldGobDeserialize simulates the old gob deserialization
func oldGobDeserialize(data []byte) (*twig.CompiledTemplate, error) {
	dec := gob.NewDecoder(bytes.NewReader(data))
	
	var compiled twig.CompiledTemplate
	if err := dec.Decode(&compiled); err != nil {
		return nil, err
	}
	
	return &compiled, nil
}