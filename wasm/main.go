//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
)

// DOM provides a JavaScript-like DOM interface
type DOM struct{}

// Element represents a DOM element with JS-like methods
type Element struct {
	el js.Value
}

// Global state
var (
	showModal bool = false
	document       = DOM{}
)

// querySelector mimics JS document.querySelector
func (d DOM) querySelector(selector string) Element {
	return Element{
		el: js.Global().Get("document").Call("querySelector", selector),
	}
}

// style mimics JS element.style property access
func (e Element) style() Style {
	return Style{
		styleObj: e.el.Get("style"),
	}
}

// text mimics JS element.textContent = "..."
func (e Element) text(content string) {
	e.el.Set("textContent", content)
}

// Style mimics JS style object
type Style struct {
	styleObj js.Value
}

// display mimics setting element.style.display = "..."
func (s Style) display(value string) {
	s.styleObj.Set("display", value)
}

// toggleModal function with JS-like syntax
func toggleModal(this js.Value, args []js.Value) interface{} {
	// Toggle state like in JS
	showModal = !showModal

	// jQuery/JS-like chaining and selection
	modal := document.querySelector(".modal")
	overlay := document.querySelector(".overlay")
	button := document.querySelector("#toggleBtn")

	// Set display like in JS: element.style.display = "block"/"none"
	if showModal {
		modal.style().display("block")
		overlay.style().display("block")
		button.text("Close Modal")
	} else {
		modal.style().display("none")
		overlay.style().display("none")
		button.text("Open Modal")
	}

	return nil
}

func main() {
	// Expose function to global scope (like a JS global function)
	js.Global().Set("toggleModal", js.FuncOf(toggleModal))

	// Prevent program from exiting
	select {}
}
