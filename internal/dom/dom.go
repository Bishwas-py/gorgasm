//go:build js && wasm
// +build js,wasm

// Package dom provides an enhanced JavaScript-like DOM manipulation interface for Go WebAssembly
package dom

import (
	"fmt"
	"strconv"
	"syscall/js"
	"time"
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

// ClassList represents a DOM element's classList object
type ClassList struct {
	ClassListObj js.Value
}

// Animation represents a CSS animation controller
type Animation struct {
	AnimObj js.Value
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

// CreateElement creates a new DOM element
func (d DOM) CreateElement(tag string) Element {
	return Element{
		El: js.Global().Get("document").Call("createElement", tag),
	}
}

// GetElementById returns an element by its ID
func (d DOM) GetElementById(id string) Element {
	return Element{
		El: js.Global().Get("document").Call("getElementById", id),
	}
}

func (e Element) QuerySelector(selector string) Element {
	return Element{
		El: e.El.Call("querySelector", selector),
	}
}

// Style returns the element's style object
func (e Element) Style() Style {
	return Style{
		StyleObj: e.El.Get("style"),
	}
}

// ClassList returns the element's classList object
func (e Element) ClassList() ClassList {
	return ClassList{
		ClassListObj: e.El.Get("classList"),
	}
}

// SetText sets the element's textContent property
func (e Element) SetText(content string) Element {
	e.El.Set("textContent", content)
	return e
}

// GetText gets the element's textContent
func (e Element) GetText() string {
	return e.El.Get("textContent").String()
}

// SetHTML sets the element's innerHTML property
func (e Element) SetHTML(content string) Element {
	e.El.Set("innerHTML", content)
	return e
}

// GetHTML gets the element's innerHTML
func (e Element) GetHTML() string {
	return e.El.Get("innerHTML").String()
}

// SetAttribute sets an attribute on the element
func (e Element) SetAttribute(name, value string) Element {
	e.El.Call("setAttribute", name, value)
	return e
}

// GetAttribute gets an attribute from the element
func (e Element) GetAttribute(name string) string {
	return e.El.Call("getAttribute", name).String()
}

// HasAttribute checks if the element has an attribute
func (e Element) HasAttribute(name string) bool {
	return e.El.Call("hasAttribute", name).Bool()
}

// RemoveAttribute removes an attribute from the element
func (e Element) RemoveAttribute(name string) Element {
	e.El.Call("removeAttribute", name)
	return e
}

// GetValue gets the value of an input element
func (e Element) GetValue() string {
	return e.El.Get("value").String()
}

// SetValue sets the value of an input element
func (e Element) SetValue(value string) Element {
	e.El.Set("value", value)
	return e
}

// Focus focuses the element
func (e Element) Focus() Element {
	e.El.Call("focus")
	return e
}

// Blur removes focus from the element
func (e Element) Blur() Element {
	e.El.Call("blur")
	return e
}

// AppendChild appends a child element
func (e Element) AppendChild(child Element) Element {
	e.El.Call("appendChild", child.El)
	return e
}

// RemoveChild removes a child element
func (e Element) RemoveChild(child Element) Element {
	e.El.Call("removeChild", child.El)
	return e
}

// GetRect gets the element's bounding rectangle
func (e Element) GetRect() map[string]float64 {
	rect := e.El.Call("getBoundingClientRect")
	return map[string]float64{
		"top":    rect.Get("top").Float(),
		"right":  rect.Get("right").Float(),
		"bottom": rect.Get("bottom").Float(),
		"left":   rect.Get("left").Float(),
		"width":  rect.Get("width").Float(),
		"height": rect.Get("height").Float(),
	}
}

// AddEventListener adds an event listener to the element with a callback
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

// AddEventListenerWithEvent adds an event listener with the event object
func (e Element) AddEventListenerWithEvent(event string, fn func(js.Value)) Element {
	callback := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			fn(args[0])
		}
		return nil
	})

	// Store callback to prevent garbage collection
	e.El.Call("addEventListener", event, callback)
	return e
}

// RemoveEventListener removes an event listener (simplified without callback reference)
func (e Element) RemoveEventListener(event string) Element {
	// Note: This is simplified and won't actually work as expected
	// because we need the original callback reference
	return e
}

// Animate creates a CSS animation and returns the animation object
func (e Element) Animate(keyframes []map[string]interface{}, options map[string]interface{}) Animation {
	// Convert Go maps to JS objects
	jsKeyframes := js.Global().Get("Array").New(len(keyframes))
	for i, keyframe := range keyframes {
		jsKeyframe := js.Global().Get("Object").New()
		for key, value := range keyframe {
			jsKeyframe.Set(key, value)
		}
		jsKeyframes.SetIndex(i, jsKeyframe)
	}

	jsOptions := js.Global().Get("Object").New()
	for key, value := range options {
		jsOptions.Set(key, value)
	}

	return Animation{
		AnimObj: e.El.Call("animate", jsKeyframes, jsOptions),
	}
}

// AnimateWithOptions animates with preset options
func (e Element) AnimateWithOptions(animationType string, duration int) Animation {
	var keyframes []map[string]interface{}
	var options map[string]interface{}

	switch animationType {
	case "fadeIn":
		keyframes = []map[string]interface{}{
			{"opacity": 0},
			{"opacity": 1},
		}
	case "fadeOut":
		keyframes = []map[string]interface{}{
			{"opacity": 1},
			{"opacity": 0},
		}
	case "slideIn":
		keyframes = []map[string]interface{}{
			{"transform": "translateX(-20px)", "opacity": 0},
			{"transform": "translateX(0)", "opacity": 1},
		}
	case "slideOut":
		keyframes = []map[string]interface{}{
			{"transform": "translateX(0)", "opacity": 1},
			{"transform": "translateX(20px)", "opacity": 0},
		}
	case "slideInUp":
		keyframes = []map[string]interface{}{
			{"transform": "translateY(20px)", "opacity": 0},
			{"transform": "translateY(0)", "opacity": 1},
		}
	case "slideOutDown":
		keyframes = []map[string]interface{}{
			{"transform": "translateY(0)", "opacity": 1},
			{"transform": "translateY(20px)", "opacity": 0},
		}
	case "shake":
		keyframes = []map[string]interface{}{
			{"transform": "translateX(0)"},
			{"transform": "translateX(-5px)"},
			{"transform": "translateX(5px)"},
			{"transform": "translateX(-5px)"},
			{"transform": "translateX(5px)"},
			{"transform": "translateX(0)"},
		}
	}

	options = map[string]interface{}{
		"duration": duration,
		"easing":   "ease-in-out",
		"fill":     "forwards",
	}

	return e.Animate(keyframes, options)
}

// Play starts the animation
func (a Animation) Play() {
	a.AnimObj.Call("play")
}

// Pause pauses the animation
func (a Animation) Pause() {
	a.AnimObj.Call("pause")
}

// Cancel cancels the animation
func (a Animation) Cancel() {
	a.AnimObj.Call("cancel")
}

// Finish finishes the animation
func (a Animation) Finish() {
	a.AnimObj.Call("finish")
}

// OnFinish adds a callback to be executed when the animation finishes
func (a Animation) OnFinish(fn func()) {
	// Check if AnimObj is valid first
	if a.AnimObj.IsNull() || a.AnimObj.IsUndefined() {
		fmt.Println("Warning: Animation object is null or undefined")
		// Execute the function immediately as a fallback
		fn()
		return
	}

	// Create the callback
	callback := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		fn()
		return nil
	})

	// Instead of trying to access 'onfinish' property's addEventListener,
	// set the onfinish property directly to the callback
	a.AnimObj.Set("onfinish", callback)
}

// Add adds a class to the element
func (c ClassList) Add(className string) ClassList {
	c.ClassListObj.Call("add", className)
	return c
}

// Remove removes a class from the element
func (c ClassList) Remove(className string) ClassList {
	c.ClassListObj.Call("remove", className)
	return c
}

// Toggle toggles a class on the element
func (c ClassList) Toggle(className string) ClassList {
	c.ClassListObj.Call("toggle", className)
	return c
}

// Contains checks if the element has a class
func (c ClassList) Contains(className string) bool {
	return c.ClassListObj.Call("contains", className).Bool()
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

// GetProperty gets a style property
func (s Style) GetProperty(property string) string {
	return s.StyleObj.Get(property).String()
}

// SetVariable sets a CSS variable
func (s Style) SetVariable(name, value string) Style {
	s.StyleObj.Call("setProperty", "--"+name, value)
	return s
}

// GetVariable gets a CSS variable
func (s Style) GetVariable(name string) string {
	return s.StyleObj.Call("getPropertyValue", "--"+name).String()
}

// TransitionElement smoothly transitions element properties
func TransitionElement(element Element, property, value string, durationMs int) {
	// Save original transition
	originalTransition := element.Style().GetProperty("transition")

	// Set new transition
	element.Style().SetProperty("transition", property+" "+strconv.Itoa(durationMs)+"ms ease-in-out")

	// Set property (will trigger transition)
	element.Style().SetProperty(property, value)

	// Restore original transition after animation completes
	time.AfterFunc(time.Duration(durationMs)*time.Millisecond, func() {
		element.Style().SetProperty("transition", originalTransition)
	})
}

// Window provides access to the browser window object
type Window struct{}

// GetWindow returns the global window object
func GetWindow() Window {
	return Window{}
}

// SetTimeout executes a function after a specified delay
func (w Window) SetTimeout(fn func(), delayMs int) js.Value {
	callback := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		fn()
		return nil
	})

	// Store callback to prevent garbage collection
	return js.Global().Call("setTimeout", callback, delayMs)
}

// ClearTimeout clears a timeout
func (w Window) ClearTimeout(timeoutID js.Value) {
	js.Global().Call("clearTimeout", timeoutID)
}

// SetInterval executes a function at specified intervals
func (w Window) SetInterval(fn func(), intervalMs int) js.Value {
	callback := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		fn()
		return nil
	})

	// Store callback to prevent garbage collection
	return js.Global().Call("setInterval", callback, intervalMs)
}

// ClearInterval clears an interval
func (w Window) ClearInterval(intervalID js.Value) {
	js.Global().Call("clearInterval", intervalID)
}

// GetLocalStorage returns the localStorage object
func (w Window) GetLocalStorage() Storage {
	return LocalStorage()
}

// GetSessionStorage returns the sessionStorage object
func (w Window) GetSessionStorage() Storage {
	return SessionStorage()
}

// Alert displays an alert dialog
func (w Window) Alert(message string) {
	js.Global().Call("alert", message)
}

// Confirm displays a confirm dialog
func (w Window) Confirm(message string) bool {
	return js.Global().Call("confirm", message).Bool()
}

// Prompt displays a prompt dialog
func (w Window) Prompt(message, defaultValue string) string {
	return js.Global().Call("prompt", message, defaultValue).String()
}

// AddEventListener adds an event listener to the window
func (w Window) AddEventListener(event string, fn func()) {
	callback := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		fn()
		return nil
	})

	// Store callback to prevent garbage collection
	js.Global().Call("addEventListener", event, callback)
}

// AddEventListenerWithEvent adds an event listener to the window with the event object
func (w Window) AddEventListenerWithEvent(event string, fn func(js.Value)) {
	callback := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			fn(args[0])
		}
		return nil
	})

	// Store callback to prevent garbage collection
	js.Global().Call("addEventListener", event, callback)
}

// ThemeSwitcher manages theme switching
type ThemeSwitcher struct {
	CurrentTheme string
	IsDarkMode   bool
}

// NewThemeSwitcher creates a new theme switcher
func NewThemeSwitcher() ThemeSwitcher {
	return ThemeSwitcher{
		CurrentTheme: "blue",
		IsDarkMode:   false,
	}
}

// SetTheme sets the current theme
func (t *ThemeSwitcher) SetTheme(theme string) {
	t.CurrentTheme = theme
	body := Document().QuerySelector("body")

	// Remove all theme classes
	body.ClassList().Remove("theme-blue")
	body.ClassList().Remove("theme-green")
	body.ClassList().Remove("theme-purple")
	body.ClassList().Remove("theme-orange")

	// Add current theme class
	body.ClassList().Add("theme-" + theme)
}

// ToggleDarkMode toggles dark mode
func (t *ThemeSwitcher) ToggleDarkMode() {
	t.IsDarkMode = !t.IsDarkMode
	body := Document().QuerySelector("body")

	if t.IsDarkMode {
		body.ClassList().Add("dark-theme")
	} else {
		body.ClassList().Remove("dark-theme")
	}
}

// SetAnimationSpeed sets the animation speed
func SetAnimationSpeed(speed string) {
	root := Document().QuerySelector(":root").Style()

	switch speed {
	case "faster":
		root.SetVariable("anim-speed-fast", "0.1s")
		root.SetVariable("anim-speed-normal", "0.2s")
		root.SetVariable("anim-speed-slow", "0.3s")
	case "normal":
		root.SetVariable("anim-speed-fast", "0.15s")
		root.SetVariable("anim-speed-normal", "0.3s")
		root.SetVariable("anim-speed-slow", "0.5s")
	case "slower":
		root.SetVariable("anim-speed-fast", "0.3s")
		root.SetVariable("anim-speed-normal", "0.5s")
		root.SetVariable("anim-speed-slow", "0.8s")
	case "none":
		root.SetVariable("anim-speed-fast", "0s")
		root.SetVariable("anim-speed-normal", "0s")
		root.SetVariable("anim-speed-slow", "0s")
	}
}

// SetFontSize sets the font size
func SetFontSize(size string) {
	root := Document().QuerySelector(":root").Style()

	switch size {
	case "small":
		root.SetVariable("font-size-base", "14px")
	case "medium":
		root.SetVariable("font-size-base", "16px")
	case "large":
		root.SetVariable("font-size-base", "18px")
	}
}

// DragDropManager manages drag and drop functionality
type DragDropManager struct {
	DragElement Element
	DropTargets []Element
	OnDrop      func(source, target Element)
	IsDragging  bool
	OriginalPos map[string]float64
	OffsetX     float64
	OffsetY     float64
}

// NewDragDropManager creates a new drag and drop manager
func NewDragDropManager() DragDropManager {
	return DragDropManager{
		DropTargets: []Element{},
		IsDragging:  false,
		OriginalPos: map[string]float64{},
	}
}

// MakeDraggable makes an element draggable
func (d *DragDropManager) MakeDraggable(element Element) {
	element.SetAttribute("draggable", "true")

	element.AddEventListenerWithEvent("dragstart", func(event js.Value) {
		d.DragElement = element
		d.IsDragging = true

		// Store original position
		rect := element.GetRect()
		d.OriginalPos["top"] = rect["top"]
		d.OriginalPos["left"] = rect["left"]

		// Calculate offset
		d.OffsetX = event.Get("clientX").Float() - rect["left"]
		d.OffsetY = event.Get("clientY").Float() - rect["top"]

		// Add dragging class
		element.ClassList().Add("dragging")
	})

	element.AddEventListenerWithEvent("dragend", func(_ js.Value) {
		d.IsDragging = false
		element.ClassList().Remove("dragging")
	})
}

// AddDropTarget adds a drop target
func (d *DragDropManager) AddDropTarget(target Element, onDrop func(source, target Element)) {
	d.DropTargets = append(d.DropTargets, target)
	d.OnDrop = onDrop

	target.AddEventListenerWithEvent("dragover", func(event js.Value) {
		event.Call("preventDefault")
		target.ClassList().Add("drag-over")
	})

	target.AddEventListenerWithEvent("dragleave", func(event js.Value) {
		event.Call("preventDefault")
		target.ClassList().Remove("drag-over")
	})

	target.AddEventListenerWithEvent("drop", func(event js.Value) {
		event.Call("preventDefault")
		target.ClassList().Remove("drag-over")

		if d.IsDragging && d.OnDrop != nil {
			d.OnDrop(d.DragElement, target)
		}
	})
}
