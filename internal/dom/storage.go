//go:build js && wasm
// +build js,wasm

package dom

import (
	"encoding/json"
	"syscall/js"
)

// Storage represents a browser storage object (localStorage or sessionStorage)
type Storage struct {
	storageObj js.Value
}

// LocalStorage returns the browser's localStorage object
func LocalStorage() Storage {
	return Storage{
		storageObj: js.Global().Get("localStorage"),
	}
}

// SessionStorage returns the browser's sessionStorage object
func SessionStorage() Storage {
	return Storage{
		storageObj: js.Global().Get("sessionStorage"),
	}
}

// GetItem retrieves an item from storage
func (s Storage) GetItem(key string) string {
	val := s.storageObj.Call("getItem", key)
	if val.IsNull() || val.IsUndefined() {
		return ""
	}
	return val.String()
}

// SetItem sets an item in storage
func (s Storage) SetItem(key, value string) Storage {
	s.storageObj.Call("setItem", key, value)
	return s
}

// RemoveItem removes an item from storage
func (s Storage) RemoveItem(key string) Storage {
	s.storageObj.Call("removeItem", key)
	return s
}

// Clear removes all items from storage
func (s Storage) Clear() Storage {
	s.storageObj.Call("clear")
	return s
}

// Length returns the number of items in storage
func (s Storage) Length() int {
	return s.storageObj.Get("length").Int()
}

// Key returns the key at the specified index
func (s Storage) Key(index int) string {
	val := s.storageObj.Call("key", index)
	if val.IsNull() || val.IsUndefined() {
		return ""
	}
	return val.String()
}

// Keys returns all keys in storage
func (s Storage) Keys() []string {
	length := s.Length()
	keys := make([]string, length)
	for i := 0; i < length; i++ {
		keys[i] = s.Key(i)
	}
	return keys
}

// GetJSON retrieves an item from storage and unmarshals it from JSON
func (s Storage) GetJSON(key string, target interface{}) error {
	value := s.GetItem(key)
	if value == "" {
		return nil // No value stored
	}
	return json.Unmarshal([]byte(value), target)
}

// SetJSON marshals an object to JSON and stores it
func (s Storage) SetJSON(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.SetItem(key, string(data))
	return nil
}
