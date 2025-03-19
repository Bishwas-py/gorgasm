//go:build js && wasm
// +build js,wasm

// Package dom provides a JavaScript-like DOM manipulation interface for Go WebAssembly
package dom

import (
	"syscall/js"
)

// DOM provides a JavaScript-like DOM interface
type DOM struct{}

// Document returns the global document object
func Document() DOM {
	return DOM{}
}

// Element represents a DOM element with JS-like methods
type Element struct {
	El js.Value
}

// Style represents a DOM element's style object
type Style struct {
	StyleObj js.Value
}

// QuerySelector mimics JS document.querySelector
func (d DOM) QuerySelector(selector string) Element {
	return Element{
		El: js.Global().Get("document").Call("querySelector", selector),
	}
}

// QuerySelectorAll mimics JS document.querySelectorAll
func (d DOM) QuerySelectorAll(selector string) []Element {
	nodeList := js.Global().Get("document").Call("querySelectorAll", selector)
	length := nodeList.Get("length").Int()
	elements := make([]Element, length)

	for i := 0; i < length; i++ {
		elements[i] = Element{
			El: nodeList.Call("item", i),
		}
	}

	return elements
}

func (d DOM) CreateElement(tag string) Element {
	return Element{
		El: js.Global().Get("document").Call("createElement", tag),
	}
}

// Style returns the element's style object
func (e Element) Style() Style {
	return Style{
		StyleObj: e.El.Get("style"),
	}
}

// SetText sets the element's textContent property
func (e Element) SetText(content string) Element {
	e.El.Set("textContent", content)
	return e
}

// SetHTML sets the element's innerHTML property
func (e Element) SetHTML(content string) Element {
	e.El.Set("innerHTML", content)
	return e
}

// SetAttribute sets an attribute on the element
func (e Element) SetAttribute(name, value string) Element {
	e.El.Call("setAttribute", name, value)
	return e
}

// AddEventListener adds an event listener to the element
func (e Element) AddEventListener(event string, fn func()) Element {
	callback := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		fn()
		return nil
	})

	// Store callback to prevent garbage collection
	// This is a simplified approach - in a real app you'd need a way to
	// manage and remove these callbacks to prevent memory leaks
	e.El.Call("addEventListener", event, callback)
	return e
}

// Display sets the element's style.display property
func (s Style) Display(value string) Style {
	s.StyleObj.Set("display", value)
	return s
}

// SetProperty sets any style property
func (s Style) SetProperty(property, value string) Style {
	s.StyleObj.Set(property, value)
	return s
}
