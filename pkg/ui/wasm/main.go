//go:build js && wasm
// +build js,wasm

// Package main implements the WebAssembly client for the Go fullstack framework
package main

import (
	"fmt"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"gorgasm/internal/dom"
)

// Todo represents a single todo item
type Todo struct {
	ID        string   `json:"id"`        // Unique identifier
	Text      string   `json:"text"`      // Todo text
	Completed bool     `json:"completed"` // Completion status
	CreatedAt int64    `json:"createdAt"` // Creation timestamp
	Position  int      `json:"position"`  // For reordering
	Priority  int      `json:"priority"`  // Priority level (1-3)
	Tags      []string `json:"tags"`      // Tags for categorization
}

// Global state
var (
	todos           []Todo
	currentFilter   = "all"             // "all", "active", "completed"
	themeSwitcher   dom.ThemeSwitcher   // Theme manager
	dragDropManager dom.DragDropManager // Drag and drop manager
	storage         dom.CachedStorage   // Cached storage for better performance
	settingsOpen    = false             // Settings panel state
	todoBeingEdited = ""                // ID of todo being edited
)

// Storage keys
const (
	todosKey         = "gowasm-todos"
	filterKey        = "gowasm-filter"
	themeKey         = "gowasm-theme"
	darkModeKey      = "gowasm-dark-mode"
	animSpeedKey     = "gowasm-anim-speed"
	fontSizeKey      = "gowasm-font-size"
	schemaVersionKey = "gowasm-schema-version"
)

// Event handler callbacks for UI interactions
var (
	inputKeyHandler      js.Func
	themeBtnHandler      js.Func
	keyboardHandler      js.Func
	settingsBtnHandler   js.Func
	settingsCloseHandler js.Func
	themeOptionHandler   js.Func
	animSpeedHandler     js.Func
	fontSizeHandler      js.Func
)

/**
 * Initialize the application and setup event handlers
 */
func initialize() {
	// Initialize cached storage
	storage = dom.NewCachedStorage(dom.LocalStorage(), 5*time.Minute)

	// Initialize theme switcher
	themeSwitcher = dom.NewThemeSwitcher()

	// Initialize drag and drop manager
	dragDropManager = dom.NewDragDropManager()

	// Run storage migration if needed
	migrator := dom.NewStorageMigrator(storage.Storage)
	migrator.RunMigration(2, migrateTodoSchema)

	// Load saved preferences
	loadPreferences()

	// Load todos
	loadTodos()

	// Setup event listeners
	setupEventListeners()

	// Hide loading indicator
	document := dom.Document()
	loading := document.GetElementById("loading")
	loading.ClassList().Add("hidden")

	// Log initialization
	fmt.Println("Go WebAssembly Todo App initialized with enhanced features")
}

/**
 * Load todos from localStorage and render
 */
func loadTodos() {
	// Get todos from localStorage or initialize empty array
	err := storage.GetJSON(todosKey, &todos)
	if err != nil || todos == nil {
		todos = []Todo{}
	}

	// Sort todos by position property
	sortTodosByPosition()

	// Render the todos with current filter
	renderTodos(currentFilter)
}

/**
 * Sort todos by their position property
 */
func sortTodosByPosition() {
	// Simple bubble sort (for small arrays it's fine)
	n := len(todos)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if todos[j].Position > todos[j+1].Position {
				todos[j], todos[j+1] = todos[j+1], todos[j]
			}
		}
	}
}

/**
 * Save todos to localStorage
 */
func saveTodos() bool {
	err := storage.SetJSON(todosKey, todos)
	return err == nil
}

/**
 * Add a new todo
 */
func addTodo(text string) bool {
	if text == "" {
		return false
	}

	// Find the highest position value
	highestPosition := 0
	for _, todo := range todos {
		if todo.Position > highestPosition {
			highestPosition = todo.Position
		}
	}

	// Create new todo
	newTodo := Todo{
		ID:        strconv.FormatInt(time.Now().UnixNano(), 10),
		Text:      processTodoText(text),
		Completed: false,
		CreatedAt: time.Now().Unix(),
		Position:  highestPosition + 1,
		Priority:  extractPriority(text),
		Tags:      extractTags(text),
	}

	// Add to list
	todos = append(todos, newTodo)

	// Save to localStorage
	success := saveTodos()

	// Animate the new todo
	window := dom.GetWindow()
	window.SetTimeout(func() {
		// Render updated list with animation
		renderTodos(currentFilter)
	}, 10)

	return success
}

/**
 * Process todo text to extract metadata (priority, tags)
 */
func processTodoText(text string) string {
	// Remove priority marker
	for _, p := range []string{"!!!", "!!", "!"} {
		text = strings.Replace(text, p, "", 1)
	}

	// Remove tags
	words := strings.Fields(text)
	cleanedWords := []string{}

	for _, word := range words {
		if !strings.HasPrefix(word, "#") {
			cleanedWords = append(cleanedWords, word)
		}
	}

	return strings.TrimSpace(strings.Join(cleanedWords, " "))
}

/**
 * Extract priority from todo text (!, !!, !!!)
 */
func extractPriority(text string) int {
	if strings.Contains(text, "!!!") {
		return 3 // High priority
	} else if strings.Contains(text, "!!") {
		return 2 // Medium priority
	} else if strings.Contains(text, "!") {
		return 1 // Low priority
	}
	return 0 // No priority
}

/**
 * Extract tags from todo text (#tag)
 */
func extractTags(text string) []string {
	words := strings.Fields(text)
	tags := []string{}

	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			tag := strings.TrimPrefix(word, "#")
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

/**
 * Toggle todo completion status
 */
func toggleTodo(id string) bool {
	// Find and toggle the todo
	found := false
	for i := range todos {
		if todos[i].ID == id {
			// Create animation for the change
			document := dom.Document()
			element := document.QuerySelector(fmt.Sprintf("li[data-id='%s']", id))

			todos[i].Completed = !todos[i].Completed
			found = true

			// Apply animation based on new state
			if todos[i].Completed {
				element.AnimateWithOptions("fadeOut", 300).OnFinish(func() {
					element.ClassList().Add("completed")
					element.AnimateWithOptions("fadeIn", 300)
				})
			} else {
				element.AnimateWithOptions("fadeOut", 300).OnFinish(func() {
					element.ClassList().Remove("completed")
					element.AnimateWithOptions("fadeIn", 300)
				})
			}

			break
		}
	}

	if !found {
		return false
	}

	// Save to localStorage
	success := saveTodos()

	// Render updated list after animation
	window := dom.GetWindow()
	window.SetTimeout(func() {
		renderTodos(currentFilter)
	}, 600)

	return success
}

/**
 * Delete a todo
 */
func deleteTodo(id string) bool {
	// Find the todo
	index := -1
	for i, todo := range todos {
		if todo.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return false
	}

	// Apply delete animation first
	document := dom.Document()
	element := document.QuerySelector(fmt.Sprintf("li[data-id='%s']", id))
	element.ClassList().Add("todo-deleting")

	// Remove the todo after animation
	window := dom.GetWindow()
	window.SetTimeout(func() {
		// Remove the todo from the array
		todos = append(todos[:index], todos[index+1:]...)

		// Save to localStorage
		saveTodos()

		// Render updated list
		renderTodos(currentFilter)
	}, 300)

	return true
}

/**
 * Edit a todo
 */
func editTodo(id string, newText string) bool {
	if newText == "" {
		return false
	}

	// Find and update the todo
	found := false
	for i := range todos {
		if todos[i].ID == id {
			todos[i].Text = processTodoText(newText)
			todos[i].Priority = extractPriority(newText)
			todos[i].Tags = extractTags(newText)
			found = true
			break
		}
	}

	if !found {
		return false
	}

	// Save to localStorage
	success := saveTodos()

	// Exit edit mode
	todoBeingEdited = ""

	// Render updated list
	renderTodos(currentFilter)

	return success
}

/**
 * Clear completed todos
 */
func clearCompleted() int {
	// Count completed todos
	completedCount := 0
	completedIds := []string{}

	for _, todo := range todos {
		if todo.Completed {
			completedCount++
			completedIds = append(completedIds, todo.ID)
		}
	}

	// Apply animation to all completed todos
	document := dom.Document()
	for _, id := range completedIds {
		element := document.QuerySelector(fmt.Sprintf("li[data-id='%s']", id))
		element.ClassList().Add("todo-deleting")
	}

	// Remove completed todos after animation
	window := dom.GetWindow()
	window.SetTimeout(func() {
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
	}, 300)

	return completedCount
}

/**
 * Toggle all todos completion status
 */
func toggleAllTodos() int {
	// Determine if all todos are currently completed
	allCompleted := true
	for _, todo := range todos {
		if !todo.Completed {
			allCompleted = false
			break
		}
	}

	// Toggle all todos in opposite direction
	changedCount := 0
	for i := range todos {
		if todos[i].Completed == allCompleted {
			todos[i].Completed = !allCompleted
			changedCount++
		}
	}

	// Save to localStorage
	saveTodos()

	// Apply animation and render
	document := dom.Document()
	todoElements := document.QuerySelectorAll("li")

	for _, element := range todoElements {
		element.AnimateWithOptions("fadeOut", 200)
	}

	window := dom.GetWindow()
	window.SetTimeout(func() {
		renderTodos(currentFilter)
	}, 250)

	return changedCount
}

/**
 * Set the current filter
 */
func setFilter(filter string) string {
	// Validate filter
	if filter != "all" && filter != "active" && filter != "completed" && filter != "priority" {
		filter = "all"
	}

	currentFilter = filter
	storage.SetItem(filterKey, filter)

	// Update filter buttons appearance
	document := dom.Document()

	// First fade out the current list
	todoList := document.GetElementById("todo-list")
	todoList.AnimateWithOptions("fadeOut", 150)

	// Remove active class from all filters
	filterButtons := document.QuerySelectorAll(".filters button")
	for _, btn := range filterButtons {
		btn.ClassList().Remove("active")
	}

	// Add active class to current filter
	activeFilter := document.QuerySelector(".filters button[data-filter='" + filter + "']")
	activeFilter.ClassList().Add("active")

	// Render todos with animation after short delay
	window := dom.GetWindow()
	window.SetTimeout(func() {
		renderTodos(filter)
		todoList.AnimateWithOptions("fadeIn", 150)
	}, 200)

	return filter
}

/**
 * Start editing a todo
 */
/**
 * Start editing a todo
 */
func startEditTodo(id string) {
	// Set global edit state
	todoBeingEdited = id

	// Find the todo
	var todoText string
	var todoTags []string
	var todoPriority int

	for _, todo := range todos {
		if todo.ID == id {
			todoText = todo.Text
			todoTags = todo.Tags
			todoPriority = todo.Priority
			break
		}
	}

	// Create the edit input
	document := dom.Document()
	item := document.QuerySelector(fmt.Sprintf("li[data-id='%s']", id))

	// Hide the regular content
	textSpan := item.QuerySelector(".todo-text")
	textSpan.Style().Display("none")

	// Hide the delete button temporarily
	deleteBtn := item.QuerySelector(".delete")
	deleteBtn.Style().Display("none")

	// Format the todo for editing (with priority and tags)
	editValue := todoText

	// Add priority markers
	if todoPriority == 3 {
		editValue = "!!! " + editValue
	} else if todoPriority == 2 {
		editValue = "!! " + editValue
	} else if todoPriority == 1 {
		editValue = "! " + editValue
	}

	// Add tags
	for _, tag := range todoTags {
		editValue += " #" + tag
	}

	// Create edit input
	editInput := document.CreateElement("input")
	editInput.SetAttribute("type", "text")
	editInput.SetAttribute("class", "edit-todo")
	editInput.SetValue(editValue)

	// Add to DOM
	item.AppendChild(editInput)

	// Focus the input
	editInput.Focus()

	// Add event listeners
	editInput.AddEventListenerWithEvent("keydown", func(event js.Value) {
		key := event.Get("key").String()

		if key == "Enter" {
			// Save changes
			newText := editInput.GetValue()
			editTodo(id, newText)
		} else if key == "Escape" {
			// Cancel edit
			todoBeingEdited = ""
			renderTodos(currentFilter)
		}
	})

	editInput.AddEventListenerWithEvent("blur", func(_ js.Value) {
		// Save on blur if not already canceled
		if todoBeingEdited == id {
			newText := editInput.GetValue()
			editTodo(id, newText)
		}
	})

	// Add a hint about editing
	hintElement := document.CreateElement("small")
	hintElement.SetAttribute("class", "edit-hint")
	hintElement.SetText("Press Enter to save, Esc to cancel. Use ! for priority, #tag for tags")
	item.AppendChild(hintElement)
}

/**
 * Render todos based on filter
 */
func renderTodos(filter string) int {
	document := dom.Document()
	todoList := document.GetElementById("todo-list")
	todoList.SetHTML("") // Clear list

	activeCount := 0
	displayedCount := 0
	highPriorityCount := 0

	// Count active todos and high priority todos
	for _, todo := range todos {
		if !todo.Completed {
			activeCount++
			if todo.Priority >= 2 {
				highPriorityCount++
			}
		}
	}

	// Update counter
	itemsLeft := document.GetElementById("items-left")
	if activeCount == 1 {
		itemsLeft.SetText("1 item left")
	} else {
		itemsLeft.SetText(strconv.Itoa(activeCount) + " items left")
	}

	// Add high priority count if any
	if highPriorityCount > 0 {
		highPriorityText := fmt.Sprintf(" (%d high priority)", highPriorityCount)
		itemsLeftText := itemsLeft.GetText()
		itemsLeft.SetText(itemsLeftText + highPriorityText)
	}

	// Show/hide clear completed button
	clearCompletedBtn := document.GetElementById("clear-completed")
	if activeCount < len(todos) {
		clearCompletedBtn.Style().Display("inline-block")
	} else {
		clearCompletedBtn.Style().Display("none")
	}

	// Check if we're in filter mode for priorities
	isPriorityFilter := filter == "priority"

	// Filter and render todos
	for _, todo := range todos {
		// Apply filter
		if filter == "active" && todo.Completed {
			continue
		}
		if filter == "completed" && !todo.Completed {
			continue
		}
		if isPriorityFilter && todo.Priority < 1 {
			continue
		}

		displayedCount++

		// Create todo item elements
		item := document.CreateElement("li")
		if todo.Completed {
			item.ClassList().Add("completed")
		}

		// Add priority class if needed
		if todo.Priority > 0 {
			item.ClassList().Add(fmt.Sprintf("priority-%d", todo.Priority))
		}

		// Add data attributes
		item.SetAttribute("data-id", todo.ID)
		item.SetAttribute("data-position", strconv.Itoa(todo.Position))
		item.SetAttribute("draggable", "true")

		// Create checkbox with custom styling
		checkbox := document.CreateElement("input")
		checkbox.SetAttribute("type", "checkbox")
		checkbox.SetAttribute("class", fmt.Sprintf("toggle priority-%d", todo.Priority))
		checkbox.SetAttribute("data-id", todo.ID)
		if todo.Completed {
			checkbox.SetAttribute("checked", "checked")
		}

		// Create todo text with priority indicator if needed
		todoText := document.CreateElement("span")
		todoText.SetText(todo.Text)
		todoText.SetAttribute("class", "todo-text")

		// Create container for the text and tags
		textContainer := document.CreateElement("div")
		textContainer.SetAttribute("class", "text-container")
		textContainer.AppendChild(todoText)

		// Add tags if present
		if len(todo.Tags) > 0 {
			tagsElement := document.CreateElement("div")
			tagsElement.SetAttribute("class", "todo-tags")

			for _, tag := range todo.Tags {
				tagSpan := document.CreateElement("span")
				tagSpan.SetAttribute("class", "todo-tag")
				tagSpan.SetText("#" + tag)
				tagsElement.AppendChild(tagSpan)
			}

			textContainer.AppendChild(tagsElement)
		}

		// Create delete button
		deleteBtn := document.CreateElement("button")
		deleteBtn.SetText("×")
		deleteBtn.SetAttribute("class", "delete")
		deleteBtn.SetAttribute("data-id", todo.ID)

		// Create edit button
		editBtn := document.CreateElement("button")
		editBtn.SetText("✎")
		editBtn.SetAttribute("class", "edit")
		editBtn.SetAttribute("data-id", todo.ID)

		// Create button container
		buttonContainer := document.CreateElement("div")
		buttonContainer.SetAttribute("class", "button-container")
		buttonContainer.AppendChild(editBtn)
		buttonContainer.AppendChild(deleteBtn)

		// Append elements to item
		item.AppendChild(checkbox)
		item.AppendChild(textContainer)
		item.AppendChild(buttonContainer)

		// Add event listeners
		todoID := todo.ID

		checkbox.AddEventListener("change", func() {
			toggleTodo(todoID)
		})

		deleteBtn.AddEventListener("click", func() {
			deleteTodo(todoID)
		})

		editBtn.AddEventListener("click", func() {
			startEditTodo(todoID)
		})

		// Double click on text to edit
		textContainer.AddEventListener("dblclick", func() {
			startEditTodo(todoID)
		})

		// Make draggable for reordering
		setupDraggableItem(item)

		// Add item to list with staggered animation delay
		todoList.AppendChild(item)

		// Add staggered animation effect
		delay := displayedCount * 50 // staggered delay
		if delay > 500 {             // cap maximum delay
			delay = 500
		}

		window := dom.GetWindow()
		window.SetTimeout(func() {
			item.AnimateWithOptions("slideIn", 300)
		}, delay)
	}

	// Show/hide empty state message
	emptyState := document.GetElementById("empty-state")
	if len(todos) == 0 {
		emptyState.Style().Display("block")
		emptyState.AnimateWithOptions("fadeIn", 300)
	} else {
		emptyState.Style().Display("none")
	}

	return displayedCount
}

/**
 * Set up drag and drop for a todo item
 */
func setupDraggableItem(item dom.Element) {
	// Set up the drag events
	item.AddEventListenerWithEvent("dragstart", func(evt js.Value) {
		// Add dragging class
		item.ClassList().Add("dragging")

		// Set the data transfer
		id := item.GetAttribute("data-id")
		evt.Get("dataTransfer").Call("setData", "text/plain", id)
	})

	item.AddEventListenerWithEvent("dragend", func(_ js.Value) {
		// Remove dragging class
		item.ClassList().Remove("dragging")
	})

	// Set up drop events
	item.AddEventListenerWithEvent("dragover", func(evt js.Value) {
		evt.Call("preventDefault")

		// Add drop target indicator
		item.ClassList().Add("drop-target")
	})

	item.AddEventListenerWithEvent("dragleave", func(_ js.Value) {
		// Remove drop target indicator
		item.ClassList().Remove("drop-target")
	})

	item.AddEventListenerWithEvent("drop", func(evt js.Value) {
		evt.Call("preventDefault")

		// Remove drop target indicator
		item.ClassList().Remove("drop-target")

		// Get source and target IDs
		sourceID := evt.Get("dataTransfer").Call("getData", "text/plain").String()
		targetID := item.GetAttribute("data-id")

		// Don't do anything if dropped on self
		if sourceID == targetID {
			return
		}

		// Find source and target positions
		var sourcePosition, targetPosition int
		for _, todo := range todos {
			if todo.ID == sourceID {
				sourcePosition = todo.Position
			}
			if todo.ID == targetID {
				targetPosition = todo.Position
			}
		}

		// Update positions
		for i := range todos {
			if todos[i].ID == sourceID {
				if sourcePosition < targetPosition {
					// Moving down, place after target
					todos[i].Position = targetPosition
				} else {
					// Moving up, place before target
					todos[i].Position = targetPosition
				}
			} else if sourcePosition < targetPosition {
				// Moving down, decrement positions in between
				if todos[i].Position > sourcePosition && todos[i].Position <= targetPosition {
					todos[i].Position--
				}
			} else {
				// Moving up, increment positions in between
				if todos[i].Position >= targetPosition && todos[i].Position < sourcePosition {
					todos[i].Position++
				}
			}
		}

		// Save and re-render
		saveTodos()

		// Animate the reordering
		document := dom.Document()
		todoList := document.GetElementById("todo-list")
		todoList.AnimateWithOptions("fadeOut", 150).OnFinish(func() {
			renderTodos(currentFilter)
			todoList.AnimateWithOptions("fadeIn", 150)
		})
	})
}

/**
 * Load user preferences from storage
 */
func loadPreferences() {
	// Load filter preference
	filter := storage.GetItem(filterKey)
	if filter != "" {
		currentFilter = filter
	}

	// Load theme preference
	theme := storage.GetItem(themeKey)
	if theme != "" {
		themeSwitcher.SetTheme(theme)
	}

	// Load dark mode preference
	darkMode := storage.GetBool(darkModeKey, false)
	if darkMode {
		themeSwitcher.IsDarkMode = true
		document := dom.Document()
		document.QuerySelector("body").ClassList().Add("dark-theme")
	}

	// Load animation speed preference
	animSpeed := storage.GetItem(animSpeedKey)
	if animSpeed != "" {
		dom.SetAnimationSpeed(animSpeed)

		// Also update the select element
		document := dom.Document()
		animSpeedSelect := document.GetElementById("animation-speed")
		animSpeedSelect.SetValue(animSpeed)
	}

	// Load font size preference
	fontSize := storage.GetItem(fontSizeKey)
	if fontSize != "" {
		dom.SetFontSize(fontSize)

		// Also update the select element
		document := dom.Document()
		fontSizeSelect := document.GetElementById("font-size")
		fontSizeSelect.SetValue(fontSize)
	}
}

/**
 * Set up all event listeners
 */
func setupEventListeners() {
	document := dom.Document()
	window := dom.GetWindow()

	// Set up input keypress handler
	inputKeyHandler = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			event := args[0]
			key := event.Get("key").String()

			if key == "Enter" {
				handleAddTodo()
			}
		}
		return nil
	})

	// Add todo button
	addButton := document.GetElementById("add-todo")
	addButton.AddEventListener("click", func() {
		handleAddTodo()
	})

	// Enter key on input field
	newTodoInput := document.GetElementById("new-todo")
	newTodoInput.El.Call("addEventListener", "keypress", inputKeyHandler)

	// Clear completed button
	clearButton := document.GetElementById("clear-completed")
	clearButton.AddEventListener("click", func() {
		clearCompleted()
	})

	// Filter buttons
	filterButtons := document.QuerySelectorAll(".filters button")
	for _, btn := range filterButtons {
		filterName := btn.GetAttribute("data-filter")
		btn.AddEventListener("click", func() {
			setFilter(filterName)
		})
	}

	// Theme toggle button
	themeBtnHandler = js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		themeSwitcher.ToggleDarkMode()
		storage.SetBool(darkModeKey, themeSwitcher.IsDarkMode)
		return nil
	})

	themeBtn := document.GetElementById("theme-toggle")
	themeBtn.El.Call("addEventListener", "click", themeBtnHandler)

	// Settings button
	settingsBtnHandler = js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		toggleSettings()
		return nil
	})

	// Settings close button
	settingsCloseHandler = js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		toggleSettings()
		return nil
	})

	settingsClose := document.GetElementById("settings-close")
	settingsClose.El.Call("addEventListener", "click", settingsCloseHandler)

	// Theme options
	themeOptionHandler = js.FuncOf(func(this js.Value, _ []js.Value) interface{} {
		theme := this.Get("dataset").Get("theme").String()

		// Update all theme options
		themeOptions := document.QuerySelectorAll(".theme-option")
		for _, option := range themeOptions {
			option.ClassList().Remove("active")
		}

		// Set this option as active
		document.QuerySelector(fmt.Sprintf(".theme-option[data-theme='%s']", theme)).ClassList().Add("active")

		// Apply the theme
		themeSwitcher.SetTheme(theme)
		storage.SetItem(themeKey, theme)

		return nil
	})

	themeOptions := document.QuerySelectorAll(".theme-option")
	for _, option := range themeOptions {
		option.El.Call("addEventListener", "click", themeOptionHandler)
	}

	// Animation speed
	animSpeedHandler = js.FuncOf(func(this js.Value, _ []js.Value) interface{} {
		speed := this.Get("value").String()
		dom.SetAnimationSpeed(speed)
		storage.SetItem(animSpeedKey, speed)
		return nil
	})

	animSpeedSelect := document.GetElementById("animation-speed")
	animSpeedSelect.El.Call("addEventListener", "change", animSpeedHandler)

	// Font size
	fontSizeHandler = js.FuncOf(func(this js.Value, _ []js.Value) interface{} {
		size := this.Get("value").String()
		dom.SetFontSize(size)
		storage.SetItem(fontSizeKey, size)
		return nil
	})

	fontSizeSelect := document.GetElementById("font-size")
	fontSizeSelect.El.Call("addEventListener", "change", fontSizeHandler)

	// Global keyboard shortcuts
	keyboardHandler = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			event := args[0]
			key := event.Get("key").String()
			ctrlKey := event.Get("ctrlKey").Bool()

			// Ctrl+A to toggle all todos
			if ctrlKey && key == "a" {
				event.Call("preventDefault") // Prevent select all text
				toggleAllTodos()
			}

			// Esc to close settings
			if key == "Escape" && settingsOpen {
				toggleSettings()
			}

			// Esc to cancel edit
			if key == "Escape" && todoBeingEdited != "" {
				todoBeingEdited = ""
				renderTodos(currentFilter)
			}
		}
		return nil
	})

	window.AddEventListenerWithEvent("keydown", func(event js.Value) {
		key := event.Get("key").String()
		ctrlKey := event.Get("ctrlKey").Bool()

		// Ctrl+A to toggle all todos
		if ctrlKey && key == "a" {
			event.Call("preventDefault") // Prevent select all text
			toggleAllTodos()
		}

		// Esc to close settings
		if key == "Escape" && settingsOpen {
			toggleSettings()
		}

		// Esc to cancel edit
		if key == "Escape" && todoBeingEdited != "" {
			todoBeingEdited = ""
			renderTodos(currentFilter)
		}
	})
}

/**
 * Toggle settings panel
 */
func toggleSettings() {
	document := dom.Document()
	settingsPanel := document.GetElementById("settings-panel")

	settingsOpen = !settingsOpen

	if settingsOpen {
		settingsPanel.ClassList().Add("open")
	} else {
		settingsPanel.ClassList().Remove("open")
	}
}

/**
 * Add a new todo from the input field
 */
func handleAddTodo() bool {
	document := dom.Document()
	input := document.GetElementById("new-todo")
	text := input.GetValue()

	// Trim the text
	text = strings.TrimSpace(text)

	if text != "" {
		success := addTodo(text)

		// Clear input field with animation
		input.AnimateWithOptions("fadeOut", 150).OnFinish(func() {
			input.SetValue("")
			input.AnimateWithOptions("fadeIn", 150)
			input.Focus()
		})

		return success
	}

	return false
}

/**
 * Migrate todo schema between versions
 */
func migrateTodoSchema(fromVersion, toVersion int) error {
	fmt.Printf("Migrating todos from schema version %d to %d\n", fromVersion, toVersion)

	// Get the current todos
	var oldTodos []map[string]interface{}
	err := storage.GetJSON(todosKey, &oldTodos)
	if err != nil {
		return err
	}

	// If no todos, nothing to migrate
	if oldTodos == nil || len(oldTodos) == 0 {
		return nil
	}

	// Migrate from version 0 or 1 to version 2
	if fromVersion < 2 && toVersion >= 2 {
		newTodos := []Todo{}

		// For each todo, add the new fields
		for i, oldTodo := range oldTodos {
			newTodo := Todo{
				ID:        oldTodo["id"].(string),
				Text:      oldTodo["text"].(string),
				Completed: oldTodo["completed"].(bool),
				CreatedAt: int64(oldTodo["createdAt"].(float64)),
				Position:  i,          // Default to current position
				Priority:  0,          // Default priority
				Tags:      []string{}, // Default tags
			}

			newTodos = append(newTodos, newTodo)
		}

		// Save the migrated todos
		todos = newTodos
		saveTodos()
	}

	fmt.Println("Migration complete")
	return nil
}

/**
 * Main function
 */
func main() {
	initialize()

	// Register exported functions for direct calling
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

	js.Global().Set("toggleAllTodos", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		return toggleAllTodos()
	}))

	js.Global().Set("toggleDarkMode", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		themeSwitcher.ToggleDarkMode()
		storage.SetBool(darkModeKey, themeSwitcher.IsDarkMode)
		return themeSwitcher.IsDarkMode
	}))

	js.Global().Set("setTheme", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return themeSwitcher.CurrentTheme
		}
		themeSwitcher.SetTheme(args[0].String())
		storage.SetItem(themeKey, args[0].String())
		return themeSwitcher.CurrentTheme
	}))

	// Keep the program running
	select {}
}
