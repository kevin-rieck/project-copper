package engine

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

const currentConfigVersion = 2
const DefaultByteOrder = "ABCD"

type Connection struct {
	ID     string
	URI    string
	client *modbus.ModbusClient
}

type RegisterDefinition struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Register  uint16 `json:"register"`
	Count     uint16 `json:"count"`
	DataType  string `json:"data_type"`
	ByteOrder string `json:"byte_order,omitempty"`
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
	Name        string               `json:"name"`
	ModbusTable ModbusTableType      `json:"modbus_table"`
	Definitions []RegisterDefinition `json:"definitions"`
}

type Device struct {
	ID        string           `json:"id"`
	ConnID    string           `json:"conn_id"`
	SlaveID   uint8            `json:"slave_id"`
	ByteOrder string           `json:"byte_order"`
	Groups    []*RegisterGroup `json:"groups"`
}

type Engine struct {
	mu          sync.RWMutex
	connections map[string]*Connection
	devices     map[string]*Device

	OnData  func(deviceID string, results map[string]map[string]PollResult)
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

func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}

type CreateRegisterGroupRequest struct {
	DeviceID    string          `json:"device_id"`
	Name        string          `json:"name"`
	ModbusTable ModbusTableType `json:"modbus_table"`
}

type UpdateRegisterGroupRequest struct {
	DeviceID    string          `json:"device_id"`
	GroupID     string          `json:"group_id"`
	Name        string          `json:"name"`
	ModbusTable ModbusTableType `json:"modbus_table"`
}

type DeleteRegisterGroupRequest struct {
	DeviceID string `json:"device_id"`
	GroupID  string `json:"group_id"`
}

type CreateRegisterDefinitionRequest struct {
	DeviceID  string `json:"device_id"`
	GroupID   string `json:"group_id"`
	Name      string `json:"name"`
	Register  uint16 `json:"register"`
	Count     uint16 `json:"count"`
	DataType  string `json:"data_type"`
	ByteOrder string `json:"byte_order"`
}

type UpdateRegisterDefinitionRequest struct {
	DeviceID     string `json:"device_id"`
	GroupID      string `json:"group_id"`
	DefinitionID string `json:"definition_id"`
	Name         string `json:"name"`
	Register     uint16 `json:"register"`
	Count        uint16 `json:"count"`
	DataType     string `json:"data_type"`
	ByteOrder    string `json:"byte_order"`
}

type DeleteRegisterDefinitionRequest struct {
	DeviceID     string `json:"device_id"`
	GroupID      string `json:"group_id"`
	DefinitionID string `json:"definition_id"`
}

type BulkCreateRegisterDefinitionsRequest struct {
	DeviceID      string `json:"device_id"`
	GroupID       string `json:"group_id"`
	StartRegister uint16 `json:"start_register"`
	Quantity      uint16 `json:"quantity"`
	DataType      string `json:"data_type"`
	Count         uint16 `json:"count"`
	ByteOrder     string `json:"byte_order"`
	NamePattern   string `json:"name_pattern"`
}

type BulkEditRegisterDefinitionsRequest struct {
	DeviceID      string   `json:"device_id"`
	GroupID       string   `json:"group_id"`
	DefinitionIDs []string `json:"definition_ids"`
	DataType      string   `json:"data_type"`
	Count         uint16   `json:"count"`
	ByteOrder     string   `json:"byte_order"`
}

type BulkDeleteRegisterDefinitionsRequest struct {
	DeviceID      string   `json:"device_id"`
	GroupID       string   `json:"group_id"`
	DefinitionIDs []string `json:"definition_ids"`
}

type MoveRegisterDefinitionsRequest struct {
	DeviceID      string   `json:"device_id"`
	SourceGroupID string   `json:"source_group_id"`
	TargetGroupID string   `json:"target_group_id"`
	DefinitionIDs []string `json:"definition_ids"`
}

type DuplicateRegisterDefinitionsRequest struct {
	DeviceID      string   `json:"device_id"`
	SourceGroupID string   `json:"source_group_id"`
	TargetGroupID string   `json:"target_group_id"`
	DefinitionIDs []string `json:"definition_ids"`
	AddressOffset int      `json:"address_offset"`
	NamePattern   string   `json:"name_pattern"`
}

func (d *Device) findGroup(groupIDOrName string) (*RegisterGroup, bool) {
	for _, group := range d.Groups {
		if group.ID == groupIDOrName || group.Name == groupIDOrName {
			return group, true
		}
	}
	return nil, false
}

func sortDefinitions(defs []RegisterDefinition) {
	sort.SliceStable(defs, func(i, j int) bool {
		return defs[i].Register < defs[j].Register
	})
}

func defaultCount(dataType string) uint16 {
	switch dataType {
	case "float32":
		return 2
	default:
		return 1
	}
}

func renderNamePattern(pattern string, def RegisterDefinition, index int, address uint16, offset int) string {
	if pattern == "" {
		pattern = "{name} Copy"
		if def.Name == "" {
			pattern = "Register {address}"
		}
	}
	name := pattern
	replacements := map[string]string{
		"{address}": fmt.Sprintf("%d", address),
		"{index}":   fmt.Sprintf("%d", index),
		"{name}":    def.Name,
		"{offset}":  fmt.Sprintf("%d", offset),
	}
	for token, value := range replacements {
		name = strings.ReplaceAll(name, token, value)
	}
	return name
}

func validateDefinitionForGroup(group *RegisterGroup, def RegisterDefinition) error {
	if def.Count == 0 {
		return fmt.Errorf("definition %q in group %q must span at least 1 address", def.Name, group.Name)
	}
	if def.DataType == "" {
		return fmt.Errorf("definition %q in group %q must have a data type", def.Name, group.Name)
	}
	switch group.ModbusTable {
	case TableCoil, TableDiscreteInput:
		if def.DataType != "bool" {
			return fmt.Errorf("group %q only supports bool definitions", group.Name)
		}
		if def.ByteOrder != "" {
			return fmt.Errorf("byte order is not meaningful for group %q", group.Name)
		}
	case TableHoldingRegister, TableInputRegister:
		if def.DataType != "uint16" && def.DataType != "float32" {
			return fmt.Errorf("group %q only supports uint16 and float32 definitions", group.Name)
		}
		if def.DataType == "float32" && def.Count%2 != 0 {
			return fmt.Errorf("float32 definition %q in group %q must span an even number of registers", def.Name, group.Name)
		}
		if def.DataType == "uint16" && def.ByteOrder != "" {
			return fmt.Errorf("byte order is not meaningful for definition %q in group %q", def.Name, group.Name)
		}
	default:
		return fmt.Errorf("unsupported table %d", group.ModbusTable)
	}
	return nil
}

func validateNoOverlaps(groupName string, defs []RegisterDefinition) error {
	sorted := append([]RegisterDefinition(nil), defs...)
	sortDefinitions(sorted)
	for i := 1; i < len(sorted); i++ {
		prev := sorted[i-1]
		cur := sorted[i]
		prevEnd := uint32(prev.Register) + uint32(prev.Count) - 1
		if uint32(cur.Register) <= prevEnd {
			return fmt.Errorf("register definition at %d spanning %d register(s) overlaps existing definition at %d spanning %d register(s) in group %s", cur.Register, cur.Count, prev.Register, prev.Count, groupName)
		}
	}
	return nil
}

func cloneDevice(dev *Device) *Device {
	copyDev := *dev
	copyDev.Groups = make([]*RegisterGroup, 0, len(dev.Groups))
	for _, group := range dev.Groups {
		copyGroup := *group
		copyGroup.Definitions = append([]RegisterDefinition(nil), group.Definitions...)
		copyDev.Groups = append(copyDev.Groups, &copyGroup)
	}
	return &copyDev
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
		return nil
	}

	client, err := modbus.NewClient(&modbus.ClientConfiguration{URL: uri})
	if err != nil {
		return err
	}
	if err = client.Open(); err != nil {
		return err
	}

	e.connections[id] = &Connection{ID: id, URI: uri, client: client}
	return nil
}

func (e *Engine) AddDevice(id string, connID string, slaveID uint8) error {
	e.devices[id] = &Device{
		ID:        id,
		ConnID:    connID,
		SlaveID:   slaveID,
		ByteOrder: DefaultByteOrder,
		Groups:    make([]*RegisterGroup, 0),
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

func (e *Engine) CreateRegisterGroup(req CreateRegisterGroupRequest) (*Device, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	dev, ok := e.devices[req.DeviceID]
	if !ok {
		return nil, fmt.Errorf("device %s not found", req.DeviceID)
	}
	for _, group := range dev.Groups {
		if group.Name == req.Name {
			return nil, fmt.Errorf("register group %q already exists in device %s", req.Name, req.DeviceID)
		}
	}
	dev.Groups = append(dev.Groups, &RegisterGroup{ID: newID("grp"), Name: req.Name, ModbusTable: req.ModbusTable, Definitions: make([]RegisterDefinition, 0)})
	return dev, nil
}

func (e *Engine) UpdateRegisterGroup(req UpdateRegisterGroupRequest) (*Device, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	dev, ok := e.devices[req.DeviceID]
	if !ok {
		return nil, fmt.Errorf("device %s not found", req.DeviceID)
	}
	group, ok := dev.findGroup(req.GroupID)
	if !ok {
		return nil, fmt.Errorf("group %s not found in device %s", req.GroupID, req.DeviceID)
	}
	for _, existing := range dev.Groups {
		if existing.ID != group.ID && existing.Name == req.Name {
			return nil, fmt.Errorf("register group %q already exists in device %s", req.Name, req.DeviceID)
		}
	}
	if len(group.Definitions) > 0 && group.ModbusTable != req.ModbusTable {
		return nil, fmt.Errorf("group %q table can only change while it is empty", group.Name)
	}
	group.Name = req.Name
	group.ModbusTable = req.ModbusTable
	return dev, nil
}

func (e *Engine) DeleteRegisterGroup(req DeleteRegisterGroupRequest) (*Device, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	dev, ok := e.devices[req.DeviceID]
	if !ok {
		return nil, fmt.Errorf("device %s not found", req.DeviceID)
	}
	for i, group := range dev.Groups {
		if group.ID == req.GroupID || group.Name == req.GroupID {
			dev.Groups = append(dev.Groups[:i], dev.Groups[i+1:]...)
			return dev, nil
		}
	}
	return nil, fmt.Errorf("group %s not found in device %s", req.GroupID, req.DeviceID)
}

func (e *Engine) CreateRegisterDefinition(req CreateRegisterDefinitionRequest) (*Device, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	dev, group, err := e.deviceAndGroup(req.DeviceID, req.GroupID)
	if err != nil {
		return nil, err
	}
	count := req.Count
	if count == 0 {
		count = defaultCount(req.DataType)
	}
	name := req.Name
	if name == "" {
		name = fmt.Sprintf("Register %d", req.Register)
	}
	def := RegisterDefinition{ID: newID("def"), Name: name, Register: req.Register, Count: count, DataType: req.DataType, ByteOrder: req.ByteOrder}
	candidate := append(append([]RegisterDefinition(nil), group.Definitions...), def)
	if err := validateDefinitionForGroup(group, def); err != nil {
		return nil, err
	}
	if err := validateNoOverlaps(group.Name, candidate); err != nil {
		return nil, err
	}
	group.Definitions = candidate
	sortDefinitions(group.Definitions)
	return dev, nil
}

func (e *Engine) UpdateRegisterDefinition(req UpdateRegisterDefinitionRequest) (*Device, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	dev, group, err := e.deviceAndGroup(req.DeviceID, req.GroupID)
	if err != nil {
		return nil, err
	}
	defs := append([]RegisterDefinition(nil), group.Definitions...)
	for i, def := range defs {
		if def.ID == req.DefinitionID {
			count := req.Count
			if count == 0 {
				count = defaultCount(req.DataType)
			}
			updated := RegisterDefinition{ID: def.ID, Name: req.Name, Register: req.Register, Count: count, DataType: req.DataType, ByteOrder: req.ByteOrder}
			if updated.Name == "" {
				updated.Name = def.Name
			}
			if err := validateDefinitionForGroup(group, updated); err != nil {
				return nil, err
			}
			defs[i] = updated
			if err := validateNoOverlaps(group.Name, defs); err != nil {
				return nil, err
			}
			group.Definitions = defs
			sortDefinitions(group.Definitions)
			return dev, nil
		}
	}
	return nil, fmt.Errorf("definition %s not found in group %s", req.DefinitionID, group.Name)
}

func (e *Engine) DeleteRegisterDefinition(req DeleteRegisterDefinitionRequest) (*Device, error) {
	return e.BulkDeleteRegisterDefinitions(BulkDeleteRegisterDefinitionsRequest{DeviceID: req.DeviceID, GroupID: req.GroupID, DefinitionIDs: []string{req.DefinitionID}})
}

func (e *Engine) AddRegisterGroup(deviceID string, groupName string, table ModbusTableType) error {
	_, err := e.CreateRegisterGroup(CreateRegisterGroupRequest{DeviceID: deviceID, Name: groupName, ModbusTable: table})
	return err
}

func (e *Engine) AddRegisterDefinition(deviceID string, groupID string, register uint16, count uint16, dataType string) error {
	_, err := e.CreateRegisterDefinition(CreateRegisterDefinitionRequest{DeviceID: deviceID, GroupID: groupID, Register: register, Count: count, DataType: dataType})
	return err
}

func (e *Engine) deviceAndGroup(deviceID, groupID string) (*Device, *RegisterGroup, error) {
	dev, ok := e.devices[deviceID]
	if !ok {
		return nil, nil, fmt.Errorf("device %s not found", deviceID)
	}
	group, ok := dev.findGroup(groupID)
	if !ok {
		return nil, nil, fmt.Errorf("group %s not found in device %s", groupID, deviceID)
	}
	return dev, group, nil
}

func (e *Engine) BulkCreateRegisterDefinitions(req BulkCreateRegisterDefinitionsRequest) (*Device, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	dev, group, err := e.deviceAndGroup(req.DeviceID, req.GroupID)
	if err != nil {
		return nil, err
	}
	if req.Quantity == 0 {
		return nil, fmt.Errorf("bulk create in group %q must create at least one definition", group.Name)
	}
	count := req.Count
	if count == 0 {
		count = defaultCount(req.DataType)
	}
	candidate := append([]RegisterDefinition(nil), group.Definitions...)
	for i := uint16(0); i < req.Quantity; i++ {
		addr := req.StartRegister + i*count
		def := RegisterDefinition{ID: newID("def"), Name: renderNamePattern(req.NamePattern, RegisterDefinition{}, int(i), addr, 0), Register: addr, Count: count, DataType: req.DataType, ByteOrder: req.ByteOrder}
		if err := validateDefinitionForGroup(group, def); err != nil {
			return nil, err
		}
		candidate = append(candidate, def)
	}
	if err := validateNoOverlaps(group.Name, candidate); err != nil {
		return nil, err
	}
	group.Definitions = candidate
	sortDefinitions(group.Definitions)
	return dev, nil
}

func (e *Engine) BulkEditRegisterDefinitions(req BulkEditRegisterDefinitionsRequest) (*Device, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	dev, group, err := e.deviceAndGroup(req.DeviceID, req.GroupID)
	if err != nil {
		return nil, err
	}
	wanted := make(map[string]struct{}, len(req.DefinitionIDs))
	for _, id := range req.DefinitionIDs {
		wanted[id] = struct{}{}
	}
	defs := append([]RegisterDefinition(nil), group.Definitions...)
	found := 0
	for i, def := range defs {
		if _, ok := wanted[def.ID]; !ok {
			continue
		}
		found++
		updated := def
		if req.DataType != "" {
			updated.DataType = req.DataType
		}
		if req.Count != 0 {
			updated.Count = req.Count
		} else if req.DataType != "" {
			updated.Count = defaultCount(updated.DataType)
		}
		updated.ByteOrder = req.ByteOrder
		if err := validateDefinitionForGroup(group, updated); err != nil {
			return nil, err
		}
		defs[i] = updated
	}
	if found != len(wanted) {
		return nil, fmt.Errorf("one or more definitions were not found in group %q", group.Name)
	}
	if err := validateNoOverlaps(group.Name, defs); err != nil {
		return nil, err
	}
	group.Definitions = defs
	sortDefinitions(group.Definitions)
	return dev, nil
}

func (e *Engine) BulkDeleteRegisterDefinitions(req BulkDeleteRegisterDefinitionsRequest) (*Device, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	dev, group, err := e.deviceAndGroup(req.DeviceID, req.GroupID)
	if err != nil {
		return nil, err
	}
	wanted := make(map[string]struct{}, len(req.DefinitionIDs))
	for _, id := range req.DefinitionIDs {
		wanted[id] = struct{}{}
	}
	kept := make([]RegisterDefinition, 0, len(group.Definitions))
	deleted := 0
	for _, def := range group.Definitions {
		if _, ok := wanted[def.ID]; ok {
			deleted++
			continue
		}
		kept = append(kept, def)
	}
	if deleted != len(wanted) {
		return nil, fmt.Errorf("one or more definitions were not found in group %q", group.Name)
	}
	group.Definitions = kept
	return dev, nil
}

func (e *Engine) MoveRegisterDefinitions(req MoveRegisterDefinitionsRequest) (*Device, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	dev, source, err := e.deviceAndGroup(req.DeviceID, req.SourceGroupID)
	if err != nil {
		return nil, err
	}
	_, target, err := e.deviceAndGroup(req.DeviceID, req.TargetGroupID)
	if err != nil {
		return nil, err
	}
	if source.ModbusTable != target.ModbusTable {
		return nil, fmt.Errorf("definitions can only move between groups in the same Modbus table")
	}
	wanted := make(map[string]struct{}, len(req.DefinitionIDs))
	for _, id := range req.DefinitionIDs {
		wanted[id] = struct{}{}
	}
	sourceKept := make([]RegisterDefinition, 0, len(source.Definitions))
	targetDefs := append([]RegisterDefinition(nil), target.Definitions...)
	moved := 0
	for _, def := range source.Definitions {
		if _, ok := wanted[def.ID]; ok {
			moved++
			if err := validateDefinitionForGroup(target, def); err != nil {
				return nil, err
			}
			targetDefs = append(targetDefs, def)
			continue
		}
		sourceKept = append(sourceKept, def)
	}
	if moved != len(wanted) {
		return nil, fmt.Errorf("one or more definitions were not found in group %q", source.Name)
	}
	if err := validateNoOverlaps(target.Name, targetDefs); err != nil {
		return nil, err
	}
	source.Definitions = sourceKept
	target.Definitions = targetDefs
	sortDefinitions(target.Definitions)
	return dev, nil
}

func (e *Engine) DuplicateRegisterDefinitions(req DuplicateRegisterDefinitionsRequest) (*Device, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	dev, source, err := e.deviceAndGroup(req.DeviceID, req.SourceGroupID)
	if err != nil {
		return nil, err
	}
	_, target, err := e.deviceAndGroup(req.DeviceID, req.TargetGroupID)
	if err != nil {
		return nil, err
	}
	if source.ModbusTable != target.ModbusTable {
		return nil, fmt.Errorf("definitions can only duplicate between groups in the same Modbus table")
	}
	wanted := make(map[string]struct{}, len(req.DefinitionIDs))
	for _, id := range req.DefinitionIDs {
		wanted[id] = struct{}{}
	}
	candidate := append([]RegisterDefinition(nil), target.Definitions...)
	copied := 0
	for _, def := range source.Definitions {
		if _, ok := wanted[def.ID]; !ok {
			continue
		}
		newAddr := int(def.Register) + req.AddressOffset
		if newAddr < 0 || newAddr > 65535 {
			return nil, fmt.Errorf("duplicated definition %q address is out of range", def.Name)
		}
		copyDef := def
		copyDef.ID = newID("def")
		copyDef.Register = uint16(newAddr)
		copyDef.Name = renderNamePattern(req.NamePattern, def, copied, copyDef.Register, req.AddressOffset)
		if err := validateDefinitionForGroup(target, copyDef); err != nil {
			return nil, err
		}
		candidate = append(candidate, copyDef)
		copied++
	}
	if copied != len(wanted) {
		return nil, fmt.Errorf("one or more definitions were not found in group %q", source.Name)
	}
	if err := validateNoOverlaps(target.Name, candidate); err != nil {
		return nil, err
	}
	target.Definitions = candidate
	sortDefinitions(target.Definitions)
	return dev, nil
}

type pollChunk struct {
	start       uint16
	end         uint16
	definitions []RegisterDefinition
}

func chunkDefinitions(group *RegisterGroup) []pollChunk {
	defs := append([]RegisterDefinition(nil), group.Definitions...)
	sortDefinitions(defs)
	maxSpan := uint16(125)
	gapLimit := uint16(16)
	if group.ModbusTable == TableCoil || group.ModbusTable == TableDiscreteInput {
		maxSpan = 2000
		gapLimit = 64
	}
	chunks := make([]pollChunk, 0)
	for _, def := range defs {
		defEnd := def.Register + def.Count - 1
		if len(chunks) == 0 {
			chunks = append(chunks, pollChunk{start: def.Register, end: defEnd, definitions: []RegisterDefinition{def}})
			continue
		}
		last := &chunks[len(chunks)-1]
		gap := uint16(0)
		if def.Register > last.end+1 {
			gap = def.Register - last.end - 1
		}
		newSpan := defEnd - last.start + 1
		if gap > gapLimit || newSpan > maxSpan {
			chunks = append(chunks, pollChunk{start: def.Register, end: defEnd, definitions: []RegisterDefinition{def}})
			continue
		}
		last.end = defEnd
		last.definitions = append(last.definitions, def)
	}
	return chunks
}

func (e *Engine) PollDevice(deviceID string) (map[string]map[string]PollResult, error) {
	e.mu.RLock()
	dev, ok := e.devices[deviceID]
	if !ok {
		e.mu.RUnlock()
		return nil, fmt.Errorf("device %s not found", deviceID)
	}
	conn, ok := e.connections[dev.ConnID]
	if !ok {
		e.mu.RUnlock()
		return nil, fmt.Errorf("connection %s not found", dev.ConnID)
	}
	devSnapshot := cloneDevice(dev)
	e.mu.RUnlock()

	conn.client.SetUnitId(devSnapshot.SlaveID)
	results := make(map[string]map[string]PollResult)

	for _, group := range devSnapshot.Groups {
		if len(group.Definitions) == 0 {
			continue
		}
		groupResults := make(map[string]PollResult)
		chunks := chunkDefinitions(group)

		for _, chunk := range chunks {
			quantity := chunk.end - chunk.start + 1
			switch group.ModbusTable {
			case TableCoil, TableDiscreteInput:
				var bits []bool
				var err error
				if group.ModbusTable == TableCoil {
					bits, err = conn.client.ReadCoils(chunk.start, quantity)
				} else {
					bits, err = conn.client.ReadDiscreteInputs(chunk.start, quantity)
				}
				if err != nil {
					return nil, err
				}
				for _, def := range chunk.definitions {
					offset := def.Register - chunk.start
					if def.DataType == "bool" && offset < uint16(len(bits)) {
						groupResults[def.ID] = PollResult{Value: bits[offset], Raw: nil}
					}
				}
			case TableHoldingRegister, TableInputRegister:
				var regs []uint16
				var err error
				if group.ModbusTable == TableHoldingRegister {
					regs, err = conn.client.ReadRegisters(chunk.start, quantity, modbus.HOLDING_REGISTER)
				} else {
					regs, err = conn.client.ReadRegisters(chunk.start, quantity, modbus.INPUT_REGISTER)
				}
				if err != nil {
					return nil, err
				}
				for _, def := range chunk.definitions {
					offset := def.Register - chunk.start
					if def.DataType == "float32" && def.Count == 2 {
						if offset+1 < uint16(len(regs)) {
							val := math.Float32frombits(uint32(regs[offset])<<16 | uint32(regs[offset+1]))
							groupResults[def.ID] = PollResult{Value: val, Raw: []uint16{regs[offset], regs[offset+1]}}
						}
					} else if def.DataType == "uint16" && offset < uint16(len(regs)) {
						groupResults[def.ID] = PollResult{Value: regs[offset], Raw: []uint16{regs[offset]}}
					}
				}
			default:
				return nil, fmt.Errorf("unsupported table %d", group.ModbusTable)
			}
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
			} else if e.OnData != nil {
				e.OnData(deviceID, res)
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
	ConfigVersion int                `json:"config_version"`
	Connections   []ConnectionConfig `json:"connections"`
	Devices       []DeviceConfig     `json:"devices"`
}

type ConnectionConfig struct {
	ID  string `json:"id"`
	URI string `json:"uri"`
}

type DeviceConfig struct {
	ID        string          `json:"id"`
	ConnID    string          `json:"conn_id"`
	SlaveID   uint8           `json:"slave_id"`
	ByteOrder string          `json:"byte_order"`
	Groups    []RegisterGroup `json:"groups"`
}

func (e *Engine) SaveConfig(path string) error {
	cfg := Config{
		ConfigVersion: currentConfigVersion,
		Connections:   make([]ConnectionConfig, 0, len(e.connections)),
		Devices:       make([]DeviceConfig, 0, len(e.devices)),
	}
	for _, c := range e.connections {
		cfg.Connections = append(cfg.Connections, ConnectionConfig{ID: c.ID, URI: c.URI})
	}
	sort.Slice(cfg.Connections, func(i, j int) bool { return cfg.Connections[i].ID < cfg.Connections[j].ID })

	for _, d := range e.devices {
		devCfg := DeviceConfig{ID: d.ID, ConnID: d.ConnID, SlaveID: d.SlaveID, ByteOrder: d.ByteOrder, Groups: make([]RegisterGroup, 0, len(d.Groups))}
		for _, group := range d.Groups {
			copyGroup := *group
			copyGroup.Definitions = append([]RegisterDefinition(nil), group.Definitions...)
			sortDefinitions(copyGroup.Definitions)
			devCfg.Groups = append(devCfg.Groups, copyGroup)
		}
		cfg.Devices = append(cfg.Devices, devCfg)
	}
	sort.Slice(cfg.Devices, func(i, j int) bool { return cfg.Devices[i].ID < cfg.Devices[j].ID })

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
	if cfg.ConfigVersion != currentConfigVersion {
		return fmt.Errorf("unsupported config version %d", cfg.ConfigVersion)
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
		if d.ByteOrder != "" {
			dev.ByteOrder = d.ByteOrder
		}
		dev.Groups = make([]*RegisterGroup, 0, len(d.Groups))
		seenNames := make(map[string]struct{}, len(d.Groups))
		for _, g := range d.Groups {
			if _, exists := seenNames[g.Name]; exists {
				return fmt.Errorf("register group %q already exists in device %s", g.Name, d.ID)
			}
			seenNames[g.Name] = struct{}{}
			group := g
			sortDefinitions(group.Definitions)
			dev.Groups = append(dev.Groups, &group)
		}
	}
	return nil
}
