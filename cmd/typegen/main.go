package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// JSDocComment represents a parsed JSDoc comment
type JSDocComment struct {
	Description string
	Params      []JSDocParam
	Returns     string
}

// JSDocParam represents a parameter in a JSDoc comment
type JSDocParam struct {
	Name        string
	Description string
	Type        string
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <wasm-source-dir> <output-file>\n", os.Args[0])
		os.Exit(1)
	}

	sourceDir := os.Args[1]
	outputFile := os.Args[2]

	// Collect all exported functions
	exports, err := collectExports(sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting exports: %v\n", err)
		os.Exit(1)
	}

	// Generate TypeScript definitions
	typeScript := generateTypeScript(exports)

	// Write to file
	err = os.WriteFile(outputFile, []byte(typeScript), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated TypeScript definitions in %s\n", outputFile)
}

// collectExports finds all functions that are exported to JavaScript
func collectExports(sourceDir string) (map[string]JSDocComment, error) {
	exports := make(map[string]JSDocComment)

	// Walk through all .go files in the source directory
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Check if file has the js,wasm build tag
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fileStr := string(fileContent)
		if !strings.Contains(fileStr, "//go:build js && wasm") &&
			!strings.Contains(fileStr, "// +build js,wasm") {
			return nil
		}

		// Parse the file
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, fileContent, parser.ParseComments)
		if err != nil {
			return err
		}

		// Find Set calls on js.Global()
		ast.Inspect(file, func(n ast.Node) bool {
			// Look for js.Global().Set("functionName", ...)
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check if it's a method call
			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			// Check if it's the Set method
			if selExpr.Sel.Name != "Set" {
				return true
			}

			// Make sure there are at least 2 arguments
			if len(callExpr.Args) < 2 {
				return true
			}

			// First argument should be the function name as a string literal
			funcNameLit, ok := callExpr.Args[0].(*ast.BasicLit)
			if !ok || funcNameLit.Kind != token.STRING {
				return true
			}

			// Extract the function name without quotes
			funcName := strings.Trim(funcNameLit.Value, "\"'")

			// Look for JSDoc comments above this statement
			var jsDoc JSDocComment
			var comment string

			// Find the closest comment
			for _, cg := range file.Comments {
				if cg.End() < callExpr.Pos() {
					comment = cg.Text()
				}
			}

			// Parse JSDoc if available
			if comment != "" {
				jsDoc = parseJSDoc(comment)
			} else {
				// Default JSDoc if none found
				jsDoc = JSDocComment{
					Description: fmt.Sprintf("Function %s exported to JavaScript", funcName),
					Returns:     "void",
				}
			}

			exports[funcName] = jsDoc
			return true
		})

		return nil
	})

	return exports, err
}

// parseJSDoc extracts JSDoc information from a comment
func parseJSDoc(comment string) JSDocComment {
	jsDoc := JSDocComment{}

	// Extract description (first line)
	lines := strings.Split(comment, "\n")
	if len(lines) > 0 {
		jsDoc.Description = strings.TrimSpace(lines[0])
	}

	// Extract @param and @returns tags
	paramRegex := regexp.MustCompile(`@param\s+(\w+)\s+\{([^}]+)\}\s+(.*)`)
	returnsRegex := regexp.MustCompile(`@returns?\s+\{([^}]+)\}`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for @param
		if paramMatches := paramRegex.FindStringSubmatch(line); len(paramMatches) > 3 {
			jsDoc.Params = append(jsDoc.Params, JSDocParam{
				Name:        paramMatches[1],
				Type:        paramMatches[2],
				Description: paramMatches[3],
			})
		}

		// Check for @returns
		if returnsMatches := returnsRegex.FindStringSubmatch(line); len(returnsMatches) > 1 {
			jsDoc.Returns = returnsMatches[1]
		}
	}

	// Default return type if none specified
	if jsDoc.Returns == "" {
		jsDoc.Returns = "void"
	}

	return jsDoc
}

// generateTypeScript creates TypeScript definitions from the collected exports
func generateTypeScript(exports map[string]JSDocComment) string {
	var sb strings.Builder

	sb.WriteString("// This file is auto-generated. Do not edit directly.\n\n")

	// Add the Go class definition
	sb.WriteString("/**\n")
	sb.WriteString(" * Global Go object provided by wasm_exec.js\n")
	sb.WriteString(" */\n")
	sb.WriteString("declare class Go {\n")
	sb.WriteString("  importObject: WebAssembly.Imports;\n")
	sb.WriteString("  run(instance: WebAssembly.Instance): Promise<void>;\n")
	sb.WriteString("}\n\n")

	// Add each exported function
	for funcName, jsDoc := range exports {
		// Function JSDoc
		sb.WriteString("/**\n")
		sb.WriteString(fmt.Sprintf(" * %s\n", jsDoc.Description))

		// Parameters
		for _, param := range jsDoc.Params {
			sb.WriteString(fmt.Sprintf(" * @param {%s} %s %s\n", param.Type, param.Name, param.Description))
		}

		// Return type
		sb.WriteString(fmt.Sprintf(" * @returns {%s}\n", jsDoc.Returns))
		sb.WriteString(" */\n")

		// Function declaration
		sb.WriteString(fmt.Sprintf("declare function %s(", funcName))

		// Parameters
		for i, param := range jsDoc.Params {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s: %s", param.Name, param.Type))
		}

		sb.WriteString(fmt.Sprintf("): %s;\n\n", jsDoc.Returns))
	}

	return sb.String()
}
