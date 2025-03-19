# Go WebAssembly Todo App

A modern, responsive todo application built entirely with Go and WebAssembly - no JavaScript required (except for the
WebAssembly initialization)!

![golang dark mode all](https://github.com/user-attachments/assets/98709662-d94d-4c4d-a05f-0f8494a2c4d7)
![golang dark mode active](https://github.com/user-attachments/assets/f8b72ae1-c414-4e05-af39-516b81ca5914)
![golang todo app light mode](https://github.com/user-attachments/assets/428e7382-ac7a-4545-b598-3bda1f6b61d4)

## Features

- ✅ Create, toggle, and delete todos
- 📊 Filter todos by status (All/Active/Completed)
- 💾 Persistent storage using LocalStorage
- 🔄 Automatic state synchronization
- 📱 Responsive design that works on all devices
- 🚀 Pure Go implementation (no JavaScript code needed)

## Technology Stack

- **Go** - Backend logic compiled to WebAssembly
- **WebAssembly** - Frontend execution environment
- **HTML/CSS** - UI structure and styling
- **LocalStorage** - Client-side data persistence

## Project Structure

```
gorgasm/
├── cmd/
│   └── server/
│       └── main.go     # HTTP server implementation
├── internal/
│   └── dom/
│       ├── dom.go      # DOM manipulation utilities
│       └── storage.go  # LocalStorage wrapper
├── pkg/
│   └── ui/
│       └── wasm/
│           └── main.go # Main WebAssembly application logic
├── static/
│   ├── index.html      # Application HTML
│   ├── wasm_exec.js    # WebAssembly support code from Go
│   └── main.wasm       # Compiled WebAssembly binary
├── Makefile            # Build automation
└── README.md           # This file
```

## Setup and Installation

### Prerequisites

- Go 1.16+ (with WebAssembly support)
- A modern web browser

### Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/gorgasm.git
   cd gorgasm
   ```

2. Build and run the application:
   ```bash
   make serve
   ```

3. Open your browser and navigate to:
   ```
   http://localhost:8080
   ```

### Manual Build

1. Copy the WebAssembly support JavaScript:
   ```bash
   cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" static/
   ```

2. Compile the Go code to WebAssembly:
   ```bash
   GOOS=js GOARCH=wasm go build -o static/main.wasm ./pkg/ui/wasm
   ```

3. Build and run the server:
   ```bash
   go build -o server ./cmd/server
   ./server
   ```

## How It Works

This todo app demonstrates a unique approach to web development by using Go for both frontend and backend logic. Here's
how it works:

1. **Server**: A simple Go HTTP server serves static files.

2. **WebAssembly**: Go code is compiled to WebAssembly that runs in the browser.

3. **DOM Abstraction**: The app includes a custom DOM manipulation library written in Go that provides a JavaScript-like
   API for interacting with the DOM.

4. **LocalStorage**: A Go wrapper for browser's LocalStorage API enables persistent data storage.

5. **Event Handling**: All user interactions (clicks, key presses) are handled directly by Go code.

### Key Components

- **Todo Model**: Simple data structure with ID, text, completion status, and timestamp.
- **Event Handlers**: Functions that respond to user interactions.
- **Rendering Logic**: Go code that dynamically updates the DOM based on state changes.
- **Filter System**: Logic to show todos based on their completion status.

## Future Enhancements

- [ ] Todo editing functionality
- [ ] Drag-and-drop reordering
- [ ] Dark mode theme
- [ ] Due dates and priorities
- [ ] Multiple todo lists
- [ ] Syncing with a backend server

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- The Go team for their excellent WebAssembly support
- TodoMVC for inspiration on the todo app structure
