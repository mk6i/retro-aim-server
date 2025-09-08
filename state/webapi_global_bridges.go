package state

import (
	"sync"
)

// Global WebAPI bridges accessible to OSCAR services
var (
	globalWebAPIMessageBridge  *WebAPIMessageBridge
	globalWebAPIPresenceBridge *WebAPIPresenceBridge
	globalBridgeMutex          sync.RWMutex
)

// SetGlobalWebAPIBridges sets the global WebAPI bridges.
// This should be called by the WebAPI factory after creating the bridges.
func SetGlobalWebAPIBridges(messageBridge *WebAPIMessageBridge, presenceBridge *WebAPIPresenceBridge) {
	globalBridgeMutex.Lock()
	defer globalBridgeMutex.Unlock()
	globalWebAPIMessageBridge = messageBridge
	globalWebAPIPresenceBridge = presenceBridge
}

// GetGlobalWebAPIMessageBridge returns the global WebAPI message bridge if available.
func GetGlobalWebAPIMessageBridge() *WebAPIMessageBridge {
	globalBridgeMutex.RLock()
	defer globalBridgeMutex.RUnlock()
	return globalWebAPIMessageBridge
}

// GetGlobalWebAPIPresenceBridge returns the global WebAPI presence bridge if available.
func GetGlobalWebAPIPresenceBridge() *WebAPIPresenceBridge {
	globalBridgeMutex.RLock()
	defer globalBridgeMutex.RUnlock()
	return globalWebAPIPresenceBridge
}

