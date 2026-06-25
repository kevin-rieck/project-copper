package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/simonvetter/modbus"
)

type Connection struct {
	ID     string
	URI    string
	client *modbus.ModbusClient
}

type Watch struct {
	Register uint16 `json:"register"`
	Count    uint16 `json:"count"`
	DataType string `json:"data_type"`
}

type Device struct {
	ID       string
	ConnID   string
	SlaveID  uint8
	Watches  []Watch
}

type Engine struct {
	connections map[string]*Connection
	devices     map[string]*Device

	OnData  func(deviceID string, results map[uint16]interface{})
	OnError func(deviceID string, err error)

	stopChan chan struct{}
}

func NewEngine() *Engine {
	return &Engine{
		connections: make(map[string]*Connection),
		devices:     make(map[string]*Device),
	}
}

func (e *Engine) AddConnection(id string, uri string) error {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL: uri,
	})
	if err != nil {
		return err
	}
	
	e.connections[id] = &Connection{
		ID:     id,
		URI:    uri,
		client: client,
	}
	return nil
}

func (e *Engine) AddDevice(id string, connID string, slaveID uint8) error {
	if _, ok := e.connections[connID]; !ok {
		return fmt.Errorf("connection %s not found", connID)
	}
	e.devices[id] = &Device{
		ID:      id,
		ConnID:  connID,
		SlaveID: slaveID,
		Watches: make([]Watch, 0),
	}
	return nil
}

func (e *Engine) AddWatch(deviceID string, register uint16, count uint16, dataType string) error {
	dev, ok := e.devices[deviceID]
	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}
	dev.Watches = append(dev.Watches, Watch{
		Register: register,
		Count:    count,
		DataType: dataType,
	})
	return nil
}

func (e *Engine) PollDevice(deviceID string) (map[uint16]interface{}, error) {
	dev, ok := e.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}
	conn, ok := e.connections[dev.ConnID]
	if !ok {
		return nil, fmt.Errorf("connection %s not found", dev.ConnID)
	}

	// Open connection if not already open
	err := conn.client.Open()
	if err != nil {
		return nil, err
	}
	// Note: We don't defer Close() here to keep the connection persistent.
	// A proper connection manager would handle reconnects and keep-alives.

	conn.client.SetUnitId(dev.SlaveID)

	results := make(map[uint16]interface{})

	for _, watch := range dev.Watches {
		regs, err := conn.client.ReadRegisters(watch.Register, watch.Count, modbus.HOLDING_REGISTER)
		if err != nil {
			return nil, err
		}
		
		if watch.DataType == "float32" && len(regs) >= 2 {
			val := math.Float32frombits(uint32(regs[0])<<16 | uint32(regs[1]))
			results[watch.Register] = val
		} else if watch.DataType == "uint16" {
			for i, val := range regs {
				results[watch.Register+uint16(i)] = val
			}
		}
	}

	return results, nil
}

func (e *Engine) StartPolling(interval time.Duration) {
	if e.stopChan != nil {
		return // already polling
	}
	e.stopChan = make(chan struct{})

	for _, dev := range e.devices {
		go e.pollDeviceLoop(dev.ID, interval)
	}
}

func (e *Engine) pollDeviceLoop(deviceID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			return
		case <-ticker.C:
			res, err := e.PollDevice(deviceID)
			if err != nil {
				if e.OnError != nil {
					e.OnError(deviceID, err)
				}
			} else {
				if e.OnData != nil {
					e.OnData(deviceID, res)
				}
			}
		}
	}
}

func (e *Engine) StopPolling() {
	if e.stopChan != nil {
		close(e.stopChan)
		e.stopChan = nil
	}
}

type Config struct {
	Connections []ConnectionConfig `json:"connections"`
	Devices     []DeviceConfig     `json:"devices"`
}

type ConnectionConfig struct {
	ID  string `json:"id"`
	URI string `json:"uri"`
}

type DeviceConfig struct {
	ID      string  `json:"id"`
	ConnID  string  `json:"conn_id"`
	SlaveID uint8   `json:"slave_id"`
	Watches []Watch `json:"watches"`
}

func (e *Engine) SaveConfig(path string) error {
	cfg := Config{
		Connections: make([]ConnectionConfig, 0, len(e.connections)),
		Devices:     make([]DeviceConfig, 0, len(e.devices)),
	}

	for _, c := range e.connections {
		cfg.Connections = append(cfg.Connections, ConnectionConfig{ID: c.ID, URI: c.URI})
	}

	for _, d := range e.devices {
		cfg.Devices = append(cfg.Devices, DeviceConfig{
			ID:      d.ID,
			ConnID:  d.ConnID,
			SlaveID: d.SlaveID,
			Watches: d.Watches,
		})
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (e *Engine) LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	for _, c := range cfg.Connections {
		if err := e.AddConnection(c.ID, c.URI); err != nil {
			return err
		}
	}

	for _, d := range cfg.Devices {
		if err := e.AddDevice(d.ID, d.ConnID, d.SlaveID); err != nil {
			return err
		}
		dev := e.devices[d.ID]
		dev.Watches = d.Watches
	}

	return nil
}
