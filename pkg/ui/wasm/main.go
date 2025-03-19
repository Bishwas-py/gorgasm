//go:build js && wasm
// +build js,wasm

// Package main implements the WebAssembly client for the Go fullstack framework
package main

import (
	"strconv"
	"syscall/js"
	"time"

	"gorgasm/internal/dom"
)

// Todo represents a single todo item
type Todo struct {
	ID        string `json:"id"`        // Unique identifier
	Text      string `json:"text"`      // Todo text
	Completed bool   `json:"completed"` // Completion status
	CreatedAt int64  `json:"createdAt"` // Creation timestamp
}

// Global state
var (
	todos         []Todo
	currentFilter = "all" // "all", "active", "completed"
)

// Storage key for todos
const todosKey = "gowasm-todos"

/**
 * Load todos from localStorage
 * @returns {void}
 */
func loadTodos() {
	// Get todos from localStorage or initialize empty array
	err := dom.LocalStorage().GetJSON(todosKey, &todos)
	if err != nil || todos == nil {
		todos = []Todo{}
	}

	// Render the todos with current filter
	renderTodos(currentFilter)
}

/**
 * Save todos to localStorage
 * @returns {boolean} Success status
 */
func saveTodos() bool {
	err := dom.LocalStorage().SetJSON(todosKey, todos)
	return err == nil
}

/**
 * Add a new todo
 * @param {string} text Todo text
 * @returns {boolean} Success status
 */
func addTodo(text string) bool {
	if text == "" {
		return false
	}

	// Create new todo
	newTodo := Todo{
		ID:        strconv.FormatInt(time.Now().UnixNano(), 10),
		Text:      text,
		Completed: false,
		CreatedAt: time.Now().Unix(),
	}

	// Add to list
	todos = append(todos, newTodo)

	// Save to localStorage
	success := saveTodos()

	// Render updated list
	renderTodos(currentFilter)

	return success
}

/**
 * Toggle todo completion status
 * @param {string} id Todo ID
 * @returns {boolean} Success status
 */
func toggleTodo(id string) bool {
	// Find and toggle the todo
	found := false
	for i := range todos {
		if todos[i].ID == id {
			todos[i].Completed = !todos[i].Completed
			found = true
			break
		}
	}

	if !found {
		return false
	}

	// Save to localStorage
	success := saveTodos()

	// Render updated list
	renderTodos(currentFilter)

	return success
}

/**
 * Delete a todo
 * @param {string} id Todo ID
 * @returns {boolean} Success status
 */
func deleteTodo(id string) bool {
	// Find and remove the todo
	newTodos := []Todo{}
	found := false

	for _, todo := range todos {
		if todo.ID != id {
			newTodos = append(newTodos, todo)
		} else {
			found = true
		}
	}

	if !found {
		return false
	}

	todos = newTodos

	// Save to localStorage
	success := saveTodos()

	// Render updated list
	renderTodos(currentFilter)

	return success
}

/**
 * Clear completed todos
 * @returns {number} Number of cleared todos
 */
func clearCompleted() int {
	// Count completed todos
	completedCount := 0
	for _, todo := range todos {
		if todo.Completed {
			completedCount++
		}
	}

	// Filter out completed todos
	newTodos := []Todo{}
	for _, todo := range todos {
		if !todo.Completed {
			newTodos = append(newTodos, todo)
		}
	}

	todos = newTodos

	// Save to localStorage
	saveTodos()

	// Render updated list
	renderTodos(currentFilter)

	return completedCount
}

/**
 * Set the current filter
 * @param {string} filter Filter name ('all', 'active', 'completed')
 * @returns {string} The current filter
 */
func setFilter(filter string) string {
	// Validate filter
	if filter != "all" && filter != "active" && filter != "completed" {
		filter = "all"
	}

	currentFilter = filter

	// Update filter buttons appearance
	document := dom.Document()
	// Remove active class from all filters
	filterButtons := document.QuerySelectorAll(".filters button")
	for _, btn := range filterButtons {
		classList := btn.El.Get("classList")
		classList.Call("remove", "active")
	}

	// Add active class to current filter
	activeFilter := document.QuerySelector(".filters button[data-filter='" + filter + "']")
	classList := activeFilter.El.Get("classList")
	classList.Call("add", "active")

	// Render todos with this filter
	renderTodos(filter)

	return filter
}

/**
 * Add a new todo from the input field
 * @returns {boolean} Success status
 */
func handleAddTodo() bool {
	document := dom.Document()
	input := document.QuerySelector("#new-todo")
	text := input.El.Get("value").String()

	// Trim the text (simulate JavaScript's trim())
	text = js.Global().Get("String").New(text).Call("trim").String()

	if text != "" {
		success := addTodo(text)
		// Clear input field
		input.El.Set("value", "")
		// Focus the input field again
		input.El.Call("focus")
		return success
	}

	return false
}

/**
 * Handle key press events on the input field
 * @param {js.Value} event The keypress event
 * @returns {void}
 */
func handleInputKeypress(this js.Value, args []js.Value) interface{} {
	// Check if Enter key was pressed
	if len(args) > 0 {
		event := args[0]
		key := event.Get("key").String()

		if key == "Enter" {
			handleAddTodo()
		}
	}

	return nil
}

/**
 * Handle filter button clicks
 * @param {js.Value} event The click event
 * @returns {void}
 */
func handleFilterClick(this js.Value, args []js.Value) interface{} {
	filter := this.Get("dataset").Get("filter").String()
	setFilter(filter)
	return nil
}

/**
 * Render todos based on filter
 * @param {string} filter name ('all', 'active', 'completed')
 * @returns {number} Number of todos displayed
 */
func renderTodos(filter string) int {
	document := dom.Document()
	todoList := document.QuerySelector("#todo-list")
	todoList.SetHTML("") // Clear list

	activeCount := 0
	displayedCount := 0

	// Count active todos
	for _, todo := range todos {
		if !todo.Completed {
			activeCount++
		}
	}

	// Update counter
	itemsLeft := document.QuerySelector("#items-left")
	itemsLeft.SetText(strconv.Itoa(activeCount) + " items left")

	// Show/hide clear completed button
	clearCompletedBtn := document.QuerySelector("#clear-completed")
	if activeCount < len(todos) {
		clearCompletedBtn.Style().Display("inline-block")
	} else {
		clearCompletedBtn.Style().Display("none")
	}

	// Filter todos
	for _, todo := range todos {
		// Apply filter
		if filter == "active" && todo.Completed {
			continue
		}
		if filter == "completed" && !todo.Completed {
			continue
		}

		displayedCount++

		// Create todo item elements
		item := document.CreateElement("li")
		if todo.Completed {
			item.El.Get("classList").Call("add", "completed")
		}

		// Create checkbox
		checkbox := document.CreateElement("input")
		checkbox.SetAttribute("type", "checkbox")
		checkbox.SetAttribute("class", "toggle")
		checkbox.SetAttribute("data-id", todo.ID)
		if todo.Completed {
			checkbox.SetAttribute("checked", "checked")
		}

		// Create todo text
		todoText := document.CreateElement("span")
		todoText.SetText(todo.Text)
		todoText.SetAttribute("class", "todo-text")

		// Create delete button
		deleteBtn := document.CreateElement("button")
		deleteBtn.SetText("×")
		deleteBtn.SetAttribute("class", "delete")
		deleteBtn.SetAttribute("data-id", todo.ID)

		// Append elements to item
		item.El.Call("appendChild", checkbox.El)
		item.El.Call("appendChild", todoText.El)
		item.El.Call("appendChild", deleteBtn.El)

		// Add event listeners using callback references
		// Store these in a map to prevent garbage collection
		todoID := todo.ID

		checkbox.AddEventListener("change", func() {
			toggleTodo(todoID)
		})

		deleteBtn.AddEventListener("click", func() {
			deleteTodo(todoID)
		})

		// Add item to list
		todoList.El.Call("appendChild", item.El)
	}

	// Show/hide empty state message
	emptyState := document.QuerySelector("#empty-state")
	if len(todos) == 0 {
		emptyState.Style().Display("block")
	} else {
		emptyState.Style().Display("none")
	}

	return displayedCount
}

/**
 * Set up all event listeners for the app
 * @returns {void}
 */
func setupEventListeners() {
	document := dom.Document()

	// Add todo button
	addButton := document.QuerySelector("#add-todo")
	addButton.AddEventListener("click", func() {
		handleAddTodo()
	})

	// Enter key on input field
	newTodoInput := document.QuerySelector("#new-todo")
	newTodoInput.El.Call("addEventListener", "keypress", js.FuncOf(handleInputKeypress))

	// Clear completed button
	clearButton := document.QuerySelector("#clear-completed")
	clearButton.AddEventListener("click", func() {
		clearCompleted()
	})

	// Filter buttons
	filterButtons := document.QuerySelectorAll(".filters button")
	for _, btn := range filterButtons {
		btn.El.Call("addEventListener", "click", js.FuncOf(handleFilterClick))
	}
}

func main() {
	// Register exported functions for direct JavaScript calls
	js.Global().Set("loadTodos", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		loadTodos()
		return nil
	}))

	js.Global().Set("addTodo", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return false
		}
		return addTodo(args[0].String())
	}))

	js.Global().Set("toggleTodo", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return false
		}
		return toggleTodo(args[0].String())
	}))

	js.Global().Set("deleteTodo", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return false
		}
		return deleteTodo(args[0].String())
	}))

	js.Global().Set("clearCompleted", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		return clearCompleted()
	}))

	js.Global().Set("setFilter", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "all"
		}
		return setFilter(args[0].String())
	}))

	// Set up event listeners and load todos on startup
	setupEventListeners()
	loadTodos()

	// Print to console to confirm loading
	println("Go WebAssembly Todo App initialized")

	// Keep the program running
	select {}
}
