package main

import (
	"context"
	"time"

	"morbus/backend/engine"

	"github.com/simonvetter/modbus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	Engine *engine.Engine
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		Engine: engine.NewEngine(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	a.Engine.OnData = func(deviceID string, results map[uint16]engine.PollResult) {
		runtime.EventsEmit(ctx, "modbusData", map[string]interface{}{
			"deviceID": deviceID,
			"data":     results,
		})
	}
	
	a.Engine.OnError = func(deviceID string, err error) {
		runtime.EventsEmit(ctx, "modbusError", map[string]interface{}{
			"deviceID": deviceID,
			"error":    err.Error(),
		})
	}
}

// AddConnection adds a new Modbus connection
func (a *App) AddConnection(id string, uri string) error {
	return a.Engine.AddConnection(id, uri)
}

// AddDevice adds a new logical device to a connection
func (a *App) AddDevice(id string, connID string, slaveID uint8) error {
	return a.Engine.AddDevice(id, connID, slaveID)
}

// AddRegisterGroup adds a logical group of registers
func (a *App) AddRegisterGroup(deviceID string, groupID string, table uint8) error {
	// Wails usually passes uint8 as integer, we cast to modbus.RegType
	return a.Engine.AddRegisterGroup(deviceID, groupID, modbus.RegType(table))
}

func (a *App) AddRegisterDefinition(deviceID string, groupID string, register uint16, count uint16, dataType string) error {
	return a.Engine.AddRegisterDefinition(deviceID, groupID, register, count, dataType)
}

// GetDeviceConfig returns the full device configuration (groups and definitions)
func (a *App) GetDeviceConfig(deviceID string) (*engine.Device, error) {
	return a.Engine.GetDeviceConfig(deviceID)
}

// StartPolling begins the background polling loop
func (a *App) StartPolling() {
	a.Engine.StartPolling(500 * time.Millisecond)
}
