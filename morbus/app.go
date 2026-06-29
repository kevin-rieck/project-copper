package main

import (
	"context"
	"fmt"
	"time"

	"morbus/backend/engine"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	Engine *engine.Engine

	saveFileDialog func() (string, error)
	loadFileDialog func() (string, error)
}

type ConfigLoadResult struct {
	ActiveDeviceID string   `json:"activeDeviceID"`
	DeviceIDs      []string `json:"deviceIDs"`
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
	if a.saveFileDialog == nil {
		a.saveFileDialog = func() (string, error) {
			return runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
				Title:           "Save Modbus Configuration",
				DefaultFilename: "morbus-config.json",
				Filters: []runtime.FileFilter{{
					DisplayName: "JSON Files (*.json)",
					Pattern:     "*.json",
				}},
			})
		}
	}
	if a.loadFileDialog == nil {
		a.loadFileDialog = func() (string, error) {
			return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
				Title: "Load Modbus Configuration",
				Filters: []runtime.FileFilter{{
					DisplayName: "JSON Files (*.json)",
					Pattern:     "*.json",
				}},
			})
		}
	}

	a.Engine.OnData = func(deviceID string, results map[string]map[uint16]engine.PollResult) {
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

// SaveConfig prompts for a target JSON file and saves the current Modbus project layout.
func (a *App) SaveConfig() error {
	path, err := a.saveFileDialog()
	if err != nil {
		return err
	}
	if path == "" {
		return nil
	}
	return a.Engine.SaveConfig(path)
}

// LoadConfig prompts for a JSON project file, loads it, and restarts polling.
func (a *App) LoadConfig() (*ConfigLoadResult, error) {
	path, err := a.loadFileDialog()
	if err != nil {
		return nil, err
	}
	if path == "" {
		return &ConfigLoadResult{}, nil
	}
	if err := a.Engine.LoadConfig(path); err != nil {
		return nil, err
	}

	deviceIDs := a.Engine.DeviceIDs()
	activeDeviceID := ""
	if len(deviceIDs) > 0 {
		activeDeviceID = deviceIDs[0]
		a.StartPolling()
	}

	return &ConfigLoadResult{
		ActiveDeviceID: activeDeviceID,
		DeviceIDs:      deviceIDs,
	}, nil
}

// AddDevice adds a new logical device to a connection
func (a *App) AddDevice(id string, connID string, slaveID uint8) error {
	return a.Engine.AddDevice(id, connID, slaveID)
}

// AddRegisterGroup adds a logical group of registers
func (a *App) AddRegisterGroup(deviceID string, groupID string, table uint8) error {
	// Wails usually passes uint8 as integer, we cast to engine.ModbusTableType
	return a.Engine.AddRegisterGroup(deviceID, groupID, engine.ModbusTableType(table))
}

func (a *App) AddRegisterDefinition(deviceID string, groupID string, register uint16, count uint16, dataType string) error {
	return a.Engine.AddRegisterDefinition(deviceID, groupID, register, count, dataType)
}

// AddDeviceWithDefaults handles the complexity of creating a Connection, Device, and default registers
func (a *App) AddDeviceWithDefaults(uri string, slaveID uint8) error {
	// 1. Connection (use URI as the connection ID). AddConnection errors if it fails to parse,
	// but we can ignore "already exists" errors by just attempting it and ignoring or checking first.
	// Actually engine.AddConnection just overwrites if it exists. So we can just call it.
	err := a.AddConnection(uri, uri)
	if err != nil {
		return err
	}

	// 2. Device (ID = URI + _ + SlaveID)
	// We use a simple composite key for Device ID to prevent duplicates
	devID := fmt.Sprintf("%s_%d", uri, slaveID)

	err = a.AddDevice(devID, uri, slaveID)
	if err != nil {
		return err
	}

	// 3. Default Groups for all 4 tables
	_ = a.AddRegisterGroup(devID, "holding_regs", uint8(engine.TableHoldingRegister))
	_ = a.AddRegisterDefinition(devID, "holding_regs", 0, 10, "uint16")

	_ = a.AddRegisterGroup(devID, "input_regs", uint8(engine.TableInputRegister))
	_ = a.AddRegisterDefinition(devID, "input_regs", 0, 10, "uint16")

	_ = a.AddRegisterGroup(devID, "coils", uint8(engine.TableCoil))
	_ = a.AddRegisterDefinition(devID, "coils", 0, 10, "bool")

	_ = a.AddRegisterGroup(devID, "discrete_inputs", uint8(engine.TableDiscreteInput))
	_ = a.AddRegisterDefinition(devID, "discrete_inputs", 0, 10, "bool")

	// Ensure polling loop is started
	a.StartPolling()

	return nil
}

// GetDeviceConfig returns the full device configuration (groups and definitions)
func (a *App) GetDeviceConfig(deviceID string) (*engine.Device, error) {
	return a.Engine.GetDeviceConfig(deviceID)
}

// StartPolling begins the background polling loop
func (a *App) StartPolling() {
	a.Engine.StartPolling(500 * time.Millisecond)
}
