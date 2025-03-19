//go:build js && wasm
// +build js,wasm

package dom

import (
	"encoding/json"
	"strconv"
	"syscall/js"
	"time"
)

// Storage represents a browser storage object (localStorage or sessionStorage)
type Storage struct {
	storageObj js.Value
}

// StorageEvent represents a storage change event
type StorageEvent struct {
	Key         string
	OldValue    string
	NewValue    string
	StorageArea string
}

// StorageObserver represents a function that observes storage changes
type StorageObserver func(event StorageEvent)

// observers holds a map of storage observers
var observers = make(map[string][]StorageObserver)

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
	oldValue := s.GetItem(key)
	s.storageObj.Call("setItem", key, value)

	// Notify observers
	s.notifyObservers(key, oldValue, value)

	return s
}

// RemoveItem removes an item from storage
func (s Storage) RemoveItem(key string) Storage {
	oldValue := s.GetItem(key)
	s.storageObj.Call("removeItem", key)

	// Notify observers
	s.notifyObservers(key, oldValue, "")

	return s
}

// Clear removes all items from storage
func (s Storage) Clear() Storage {
	keys := s.Keys()
	s.storageObj.Call("clear")

	// Notify observers for each key
	for _, key := range keys {
		s.notifyObservers(key, s.GetItem(key), "")
	}

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

// HasKey checks if a key exists in storage
func (s Storage) HasKey(key string) bool {
	for _, k := range s.Keys() {
		if k == key {
			return true
		}
	}
	return false
}

// GetInt retrieves an integer from storage
func (s Storage) GetInt(key string, defaultValue int) int {
	value := s.GetItem(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// SetInt stores an integer in storage
func (s Storage) SetInt(key string, value int) Storage {
	return s.SetItem(key, strconv.Itoa(value))
}

// GetFloat retrieves a float from storage
func (s Storage) GetFloat(key string, defaultValue float64) float64 {
	value := s.GetItem(key)
	if value == "" {
		return defaultValue
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}

	return floatValue
}

// SetFloat stores a float in storage
func (s Storage) SetFloat(key string, value float64) Storage {
	return s.SetItem(key, strconv.FormatFloat(value, 'f', -1, 64))
}

// GetBool retrieves a boolean from storage
func (s Storage) GetBool(key string, defaultValue bool) bool {
	value := s.GetItem(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// SetBool stores a boolean in storage
func (s Storage) SetBool(key string, value bool) Storage {
	return s.SetItem(key, strconv.FormatBool(value))
}

// GetTime retrieves a time from storage
func (s Storage) GetTime(key string, defaultValue time.Time) time.Time {
	value := s.GetItem(key)
	if value == "" {
		return defaultValue
	}

	// Time stored as Unix timestamp in milliseconds
	timeValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return time.Unix(0, timeValue*int64(time.Millisecond))
}

// SetTime stores a time in storage
func (s Storage) SetTime(key string, value time.Time) Storage {
	// Store as Unix timestamp in milliseconds
	return s.SetItem(key, strconv.FormatInt(value.UnixNano()/int64(time.Millisecond), 10))
}

// ObserveKey adds an observer for a specific key
func (s Storage) ObserveKey(key string, observer StorageObserver) {
	observers[key] = append(observers[key], observer)

	// Set up window storage event listener if not already done
	setupStorageEventListener()
}

// ObserveAll adds an observer for all keys
func (s Storage) ObserveAll(observer StorageObserver) {
	observers["*"] = append(observers["*"], observer)

	// Set up window storage event listener if not already done
	setupStorageEventListener()
}

// notifyObservers notifies all observers of a storage change
func (s Storage) notifyObservers(key, oldValue, newValue string) {
	event := StorageEvent{
		Key:         key,
		OldValue:    oldValue,
		NewValue:    newValue,
		StorageArea: s.getStorageAreaName(),
	}

	// Notify observers for this specific key
	for _, observer := range observers[key] {
		observer(event)
	}

	// Notify observers for all keys
	for _, observer := range observers["*"] {
		observer(event)
	}
}

// getStorageAreaName returns the name of the storage area
func (s Storage) getStorageAreaName() string {
	if s.storageObj.Equal(js.Global().Get("localStorage")) {
		return "localStorage"
	}
	return "sessionStorage"
}

// eventListenerSet keeps track of whether the storage event listener has been set
var eventListenerSet = false

// setupStorageEventListener sets up the window storage event listener
func setupStorageEventListener() {
	if eventListenerSet {
		return
	}

	eventListenerSet = true

	callback := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			storageEvent := args[0]

			key := storageEvent.Get("key").String()
			oldValue := storageEvent.Get("oldValue").String()
			newValue := storageEvent.Get("newValue").String()
			storageArea := "localStorage"

			event := StorageEvent{
				Key:         key,
				OldValue:    oldValue,
				NewValue:    newValue,
				StorageArea: storageArea,
			}

			// Notify observers for this specific key
			for _, observer := range observers[key] {
				observer(event)
			}

			// Notify observers for all keys
			for _, observer := range observers["*"] {
				observer(event)
			}
		}

		return nil
	})

	js.Global().Call("addEventListener", "storage", callback)
}

// StorageMigrator helps migrate data between schema versions
type StorageMigrator struct {
	Storage           Storage
	CurrentVersionKey string
}

// NewStorageMigrator creates a new storage migrator
func NewStorageMigrator(storage Storage) StorageMigrator {
	return StorageMigrator{
		Storage:           storage,
		CurrentVersionKey: "schemaVersion",
	}
}

// GetCurrentVersion gets the current schema version
func (m StorageMigrator) GetCurrentVersion() int {
	return m.Storage.GetInt(m.CurrentVersionKey, 0)
}

// SetCurrentVersion sets the current schema version
func (m StorageMigrator) SetCurrentVersion(version int) {
	m.Storage.SetInt(m.CurrentVersionKey, version)
}

// RunMigration runs a migration if needed
func (m StorageMigrator) RunMigration(targetVersion int, migrationFunc func(fromVersion, toVersion int) error) error {
	currentVersion := m.GetCurrentVersion()

	if currentVersion < targetVersion {
		err := migrationFunc(currentVersion, targetVersion)
		if err != nil {
			return err
		}

		m.SetCurrentVersion(targetVersion)
	}

	return nil
}

// CachedStorage adds caching to storage operations
type CachedStorage struct {
	Storage    Storage
	Cache      map[string]string
	TTL        map[string]time.Time
	DefaultTTL time.Duration
}

// NewCachedStorage creates a new cached storage
func NewCachedStorage(storage Storage, defaultTTL time.Duration) CachedStorage {
	return CachedStorage{
		Storage:    storage,
		Cache:      make(map[string]string),
		TTL:        make(map[string]time.Time),
		DefaultTTL: defaultTTL,
	}
}

// GetItem retrieves an item from cache or storage
func (c CachedStorage) GetItem(key string) string {
	// Check cache first
	if value, ok := c.Cache[key]; ok {
		// Check if TTL has expired
		if ttl, hasTTL := c.TTL[key]; !hasTTL || ttl.After(time.Now()) {
			return value
		}

		// TTL expired, remove from cache
		delete(c.Cache, key)
		delete(c.TTL, key)
	}

	// Get from storage and update cache
	value := c.Storage.GetItem(key)
	if value != "" {
		c.Cache[key] = value
		c.TTL[key] = time.Now().Add(c.DefaultTTL)
	}

	return value
}

// SetItem sets an item in cache and storage
func (c CachedStorage) SetItem(key, value string) CachedStorage {
	c.Cache[key] = value
	c.TTL[key] = time.Now().Add(c.DefaultTTL)
	c.Storage.SetItem(key, value)
	return c
}

// RemoveItem removes an item from cache and storage
func (c CachedStorage) RemoveItem(key string) CachedStorage {
	delete(c.Cache, key)
	delete(c.TTL, key)
	c.Storage.RemoveItem(key)
	return c
}

// Clear clears both cache and storage
func (c CachedStorage) Clear() CachedStorage {
	c.Cache = make(map[string]string)
	c.TTL = make(map[string]time.Time)
	c.Storage.Clear()
	return c
}

// GetBool retrieves a boolean from cache or storage
func (c CachedStorage) GetBool(key string, defaultValue bool) bool {
	value := c.GetItem(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// SetBool stores a boolean in cache and storage
func (c CachedStorage) SetBool(key string, value bool) CachedStorage {
	c.SetItem(key, strconv.FormatBool(value))
	return c
}

// InvalidateCache invalidates the entire cache
func (c CachedStorage) InvalidateCache() {
	c.Cache = make(map[string]string)
	c.TTL = make(map[string]time.Time)
}

// InvalidateKey invalidates a specific key in the cache
func (c CachedStorage) InvalidateKey(key string) {
	delete(c.Cache, key)
	delete(c.TTL, key)
}

// SetTTL sets the TTL for a specific key
func (c CachedStorage) SetTTL(key string, ttl time.Duration) {
	c.TTL[key] = time.Now().Add(ttl)
}

// GetJSON retrieves and unmarshals a JSON item from cache or storage
func (c CachedStorage) GetJSON(key string, target interface{}) error {
	value := c.GetItem(key)
	if value == "" {
		return nil
	}
	return json.Unmarshal([]byte(value), target)
}

// SetJSON marshals and stores a JSON item in cache and storage
func (c CachedStorage) SetJSON(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.SetItem(key, string(data))
	return nil
}
