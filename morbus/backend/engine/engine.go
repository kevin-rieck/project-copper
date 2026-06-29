package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

type Connection struct {
	ID     string
	URI    string
	client *modbus.ModbusClient
}

type RegisterDefinition struct {
	Register uint16 `json:"register"`
	Count    uint16 `json:"count"`
	DataType string `json:"data_type"`
}

type PollResult struct {
	Value interface{} `json:"value"`
	Raw   []uint16    `json:"raw"`
}

type ModbusTableType uint8

const (
	TableCoil            ModbusTableType = 0
	TableDiscreteInput   ModbusTableType = 1
	TableHoldingRegister ModbusTableType = 3
	TableInputRegister   ModbusTableType = 4
)

type RegisterGroup struct {
	ID          string               `json:"id"`
	ModbusTable ModbusTableType      `json:"modbus_table"`
	Definitions []RegisterDefinition `json:"definitions"`
}

type Device struct {
	ID      string                    `json:"id"`
	ConnID  string                    `json:"conn_id"`
	SlaveID uint8                     `json:"slave_id"`
	Groups  map[string]*RegisterGroup `json:"groups"`
}

type Engine struct {
	connections map[string]*Connection
	devices     map[string]*Device

	OnData  func(deviceID string, results map[string]map[uint16]PollResult)
	OnError func(deviceID string, err error)

	stopChan chan struct{}
	pollWG   sync.WaitGroup
}

func NewEngine() *Engine {
	return &Engine{
		connections: make(map[string]*Connection),
		devices:     make(map[string]*Device),
	}
}

func (e *Engine) GetConnection(id string) (*Connection, error) {
	conn, ok := e.connections[id]
	if !ok {
		return nil, fmt.Errorf("connection %s not found", id)
	}
	return conn, nil
}

func (e *Engine) AddConnection(id string, uri string) error {
	if _, exists := e.connections[id]; exists {
		// Connection already exists, reuse it
		return nil
	}

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL: uri,
	})
	if err != nil {
		return err
	}

	err = client.Open()
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
	e.devices[id] = &Device{
		ID:      id,
		ConnID:  connID,
		SlaveID: slaveID,
		Groups:  make(map[string]*RegisterGroup),
	}
	return nil
}

func (e *Engine) GetDeviceConfig(deviceID string) (*Device, error) {
	dev, ok := e.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}
	return dev, nil
}

func (e *Engine) DeviceIDs() []string {
	ids := make([]string, 0, len(e.devices))
	for id := range e.devices {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (e *Engine) AddRegisterGroup(deviceID string, groupID string, table ModbusTableType) error {
	dev, ok := e.devices[deviceID]
	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}
	dev.Groups[groupID] = &RegisterGroup{
		ID:          groupID,
		ModbusTable: table,
		Definitions: make([]RegisterDefinition, 0),
	}
	return nil
}

func (e *Engine) AddRegisterDefinition(deviceID string, groupID string, register uint16, count uint16, dataType string) error {
	dev, ok := e.devices[deviceID]
	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}
	group, ok := dev.Groups[groupID]
	if !ok {
		return fmt.Errorf("group %s not found in device %s", groupID, deviceID)
	}
	group.Definitions = append(group.Definitions, RegisterDefinition{
		Register: register,
		Count:    count,
		DataType: dataType,
	})
	return nil
}

func (e *Engine) PollDevice(deviceID string) (map[string]map[uint16]PollResult, error) {
	dev, ok := e.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}

	conn, ok := e.connections[dev.ConnID]
	if !ok {
		return nil, fmt.Errorf("connection %s not found", dev.ConnID)
	}

	conn.client.SetUnitId(dev.SlaveID)

	results := make(map[string]map[uint16]PollResult)

	for _, group := range dev.Groups {
		if len(group.Definitions) == 0 {
			continue
		}

		minReg := uint16(0xFFFF)
		var maxReg uint16 = 0

		for _, def := range group.Definitions {
			if def.Register < minReg {
				minReg = def.Register
			}
			endReg := def.Register + def.Count - 1
			if endReg > maxReg {
				maxReg = endReg
			}
		}

		quantity := maxReg - minReg + 1
		groupResults := make(map[uint16]PollResult)

		switch group.ModbusTable {
		case TableCoil, TableDiscreteInput:
			var bits []bool
			var err error
			if group.ModbusTable == TableCoil {
				bits, err = conn.client.ReadCoils(minReg, quantity)
			} else {
				bits, err = conn.client.ReadDiscreteInputs(minReg, quantity)
			}
			if err != nil {
				return nil, err
			}

			for _, def := range group.Definitions {
				offset := def.Register - minReg
				if def.DataType == "bool" {
					for i := uint16(0); i < def.Count; i++ {
						if offset+i < uint16(len(bits)) {
							groupResults[def.Register+i] = PollResult{
								Value: bits[offset+i],
								Raw:   nil,
							}
						}
					}
				}
			}

		case TableHoldingRegister, TableInputRegister:
			var regs []uint16
			var err error
			if group.ModbusTable == TableHoldingRegister {
				regs, err = conn.client.ReadRegisters(minReg, quantity, modbus.HOLDING_REGISTER)
			} else {
				regs, err = conn.client.ReadRegisters(minReg, quantity, modbus.INPUT_REGISTER)
			}
			if err != nil {
				return nil, err
			}

			for _, def := range group.Definitions {
				offset := def.Register - minReg

				if def.DataType == "float32" && def.Count == 2 {
					if offset+1 < uint16(len(regs)) {
						val := math.Float32frombits(uint32(regs[offset])<<16 | uint32(regs[offset+1]))
						groupResults[def.Register] = PollResult{
							Value: val,
							Raw:   []uint16{regs[offset], regs[offset+1]},
						}
					}
				} else if def.DataType == "uint16" {
					for i := uint16(0); i < def.Count; i++ {
						if offset+i < uint16(len(regs)) {
							groupResults[def.Register+i] = PollResult{
								Value: regs[offset+i],
								Raw:   []uint16{regs[offset+i]},
							}
						}
					}
				}
			}
		default:
			return nil, fmt.Errorf("unsupported table %d", group.ModbusTable)
		}

		results[group.ID] = groupResults
	}

	return results, nil
}

func (e *Engine) StartPolling(interval time.Duration) {
	if e.stopChan != nil {
		return
	}
	e.stopChan = make(chan struct{})

	for _, dev := range e.devices {
		e.pollWG.Add(1)
		go e.pollDeviceLoop(dev.ID, interval)
	}
}

func (e *Engine) pollDeviceLoop(deviceID string, interval time.Duration) {
	defer e.pollWG.Done()

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
		e.pollWG.Wait()
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
	ID      string                   `json:"id"`
	ConnID  string                   `json:"conn_id"`
	SlaveID uint8                    `json:"slave_id"`
	Groups  map[string]RegisterGroup `json:"groups"`
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
		devCfg := DeviceConfig{
			ID:      d.ID,
			ConnID:  d.ConnID,
			SlaveID: d.SlaveID,
			Groups:  make(map[string]RegisterGroup),
		}
		for k, v := range d.Groups {
			devCfg.Groups[k] = *v
		}
		cfg.Devices = append(cfg.Devices, devCfg)
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

	e.StopPolling()
	for _, c := range e.connections {
		_ = c.client.Close()
	}
	e.connections = make(map[string]*Connection)
	e.devices = make(map[string]*Device)

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
		for k, v := range d.Groups {
			val := v
			dev.Groups[k] = &val
		}
	}

	return nil
}
