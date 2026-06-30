package engine_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/simonvetter/modbus"
	"morbus/backend/engine"
)

type mockModbusHandler struct{}

type recordingRegisterHandler struct {
	requests []modbus.HoldingRegistersRequest
	failAddr uint16
}

func (h *recordingRegisterHandler) HandleCoils(req *modbus.CoilsRequest) ([]bool, error) {
	return make([]bool, req.Quantity), nil
}
func (h *recordingRegisterHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) ([]bool, error) {
	return make([]bool, req.Quantity), nil
}
func (h *recordingRegisterHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) ([]uint16, error) {
	h.requests = append(h.requests, *req)
	if req.Addr == h.failAddr {
		return nil, fmt.Errorf("forced chunk failure")
	}
	regs := make([]uint16, req.Quantity)
	for i := range regs {
		regs[i] = req.Addr + uint16(i)
	}
	return regs, nil
}
func (h *recordingRegisterHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) ([]uint16, error) {
	return make([]uint16, req.Quantity), nil
}

func findGroup(t *testing.T, dev *engine.Device, name string) *engine.RegisterGroup {
	t.Helper()
	for _, group := range dev.Groups {
		if group.Name == name {
			return group
		}
	}
	t.Fatalf("expected group %q in %+v", name, dev.Groups)
	return nil
}

func (h *mockModbusHandler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	return make([]bool, req.Quantity), nil
}
func (h *mockModbusHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	return make([]bool, req.Quantity), nil
}
func (h *mockModbusHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	return make([]uint16, req.Quantity), nil
}
func (h *mockModbusHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {
	return make([]uint16, req.Quantity), nil
}

func TestEngineSaveLoadUsesConfigV2StableIdentityShape(t *testing.T) {
	eng := engine.NewEngine()
	if err := eng.AddDevice("dev1", "conn1", 1); err != nil {
		t.Fatalf("AddDevice failed: %v", err)
	}
	if err := eng.AddRegisterGroup("dev1", "Holding Registers", engine.TableHoldingRegister); err != nil {
		t.Fatalf("AddRegisterGroup failed: %v", err)
	}
	if err := eng.AddRegisterDefinition("dev1", "Holding Registers", 10, 1, "uint16"); err != nil {
		t.Fatalf("AddRegisterDefinition failed: %v", err)
	}

	path := filepath.Join(t.TempDir(), "project.json")
	if err := eng.SaveConfig(path); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved config: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("saved config is JSON: %v", err)
	}
	if got := raw["config_version"]; got != float64(2) {
		t.Fatalf("expected config_version 2, got %#v", got)
	}
	devices := raw["devices"].([]any)
	device := devices[0].(map[string]any)
	if got := device["byte_order"]; got != "ABCD" {
		t.Fatalf("expected default byte_order ABCD, got %#v", got)
	}
	groups, ok := device["groups"].([]any)
	if !ok {
		t.Fatalf("expected device groups to be a JSON array, got %#v", device["groups"])
	}
	group := groups[0].(map[string]any)
	if group["id"] == "" || group["id"] == nil {
		t.Fatalf("expected group to have a generated id: %#v", group)
	}
	if got := group["name"]; got != "Holding Registers" {
		t.Fatalf("expected group name to be saved, got %#v", got)
	}
	definitions := group["definitions"].([]any)
	definition := definitions[0].(map[string]any)
	if definition["id"] == "" || definition["id"] == nil {
		t.Fatalf("expected definition to have a generated id: %#v", definition)
	}

	loaded := engine.NewEngine()
	if err := loaded.LoadConfig(path); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	dev, err := loaded.GetDeviceConfig("dev1")
	if err != nil {
		t.Fatalf("GetDeviceConfig failed: %v", err)
	}
	loadedGroup := findGroup(t, dev, "Holding Registers")
	if loadedGroup.ID != group["id"] {
		t.Fatalf("expected group ID to remain stable after load, got %q want %q", loadedGroup.ID, group["id"])
	}
	if loadedGroup.Definitions[0].ID != definition["id"] {
		t.Fatalf("expected definition ID to remain stable after load")
	}
}

func TestEngineLoadConfigReplacesTheCurrentProjectLayout(t *testing.T) {
	oldURI := "tcp://127.0.0.1:55556"
	newURI := "tcp://127.0.0.1:55557"
	startMockServer := func(uri string) {
		t.Helper()
		server, err := modbus.NewServer(&modbus.ServerConfiguration{
			URL:        uri,
			Timeout:    10 * time.Second,
			MaxClients: 2,
		}, &mockModbusHandler{})
		if err != nil {
			t.Fatalf("Failed to create server %s: %v", uri, err)
		}
		if err := server.Start(); err != nil {
			t.Fatalf("Failed to start server %s: %v", uri, err)
		}
		t.Cleanup(func() { _ = server.Stop() })
	}
	startMockServer(oldURI)
	startMockServer(newURI)

	oldProject := engine.NewEngine()
	if err := oldProject.AddConnection("old_conn", oldURI); err != nil {
		t.Fatalf("Failed to add old connection: %v", err)
	}
	if err := oldProject.AddDevice("old_device", "old_conn", 1); err != nil {
		t.Fatalf("Failed to add old device: %v", err)
	}

	newProject := engine.NewEngine()
	if err := newProject.AddConnection("new_conn", newURI); err != nil {
		t.Fatalf("Failed to add new connection: %v", err)
	}
	if err := newProject.AddDevice("new_device", "new_conn", 2); err != nil {
		t.Fatalf("Failed to add new device: %v", err)
	}
	if err := newProject.AddRegisterGroup("new_device", "holding", engine.TableHoldingRegister); err != nil {
		t.Fatalf("Failed to add group: %v", err)
	}

	path := filepath.Join(t.TempDir(), "new-project.json")
	if err := newProject.SaveConfig(path); err != nil {
		t.Fatalf("Failed to save new project: %v", err)
	}

	if err := oldProject.LoadConfig(path); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if _, err := oldProject.GetDeviceConfig("old_device"); err == nil {
		t.Fatalf("expected old device to be removed after loading a new project")
	}
	loadedDevice, err := oldProject.GetDeviceConfig("new_device")
	if err != nil {
		t.Fatalf("expected new device to be available after load: %v", err)
	}
	findGroup(t, loadedDevice, "holding")
}

func TestEngineMutationAPIsReturnUpdatedDeviceAndApplyAtomically(t *testing.T) {
	eng := engine.NewEngine()
	if err := eng.AddDevice("dev1", "conn1", 1); err != nil {
		t.Fatalf("AddDevice failed: %v", err)
	}

	dev, err := eng.CreateRegisterGroup(engine.CreateRegisterGroupRequest{DeviceID: "dev1", Name: "Holding", ModbusTable: engine.TableHoldingRegister})
	if err != nil {
		t.Fatalf("CreateRegisterGroup failed: %v", err)
	}
	group := findGroup(t, dev, "Holding")
	if group.ID == "" || group.Name != "Holding" {
		t.Fatalf("expected generated group identity and name, got %+v", group)
	}

	dev, err = eng.CreateRegisterDefinition(engine.CreateRegisterDefinitionRequest{DeviceID: "dev1", GroupID: group.ID, Name: "Speed", Register: 10, DataType: "float32"})
	if err != nil {
		t.Fatalf("CreateRegisterDefinition failed: %v", err)
	}
	def := findGroup(t, dev, "Holding").Definitions[0]
	if def.ID == "" || def.Count != 2 || def.Name != "Speed" {
		t.Fatalf("expected default float32 count and generated definition ID, got %+v", def)
	}

	before := len(findGroup(t, dev, "Holding").Definitions)
	_, err = eng.BulkCreateRegisterDefinitions(engine.BulkCreateRegisterDefinitionsRequest{DeviceID: "dev1", GroupID: group.ID, StartRegister: 11, Quantity: 2, DataType: "uint16", NamePattern: "Register {address}"})
	if err == nil {
		t.Fatalf("expected overlapping bulk create to fail")
	}
	current, _ := eng.GetDeviceConfig("dev1")
	if got := len(findGroup(t, current, "Holding").Definitions); got != before {
		t.Fatalf("failed bulk create mutated definitions, got %d want %d", got, before)
	}
}

func TestEngineMutationAPIsCoverBulkEditDeleteMoveAndDuplicate(t *testing.T) {
	eng := engine.NewEngine()
	_ = eng.AddDevice("dev1", "conn1", 1)
	dev, err := eng.CreateRegisterGroup(engine.CreateRegisterGroupRequest{DeviceID: "dev1", Name: "Source", ModbusTable: engine.TableHoldingRegister})
	if err != nil {
		t.Fatalf("Create source group: %v", err)
	}
	source := findGroup(t, dev, "Source")
	dev, err = eng.CreateRegisterGroup(engine.CreateRegisterGroupRequest{DeviceID: "dev1", Name: "Target", ModbusTable: engine.TableHoldingRegister})
	if err != nil {
		t.Fatalf("Create target group: %v", err)
	}
	target := findGroup(t, dev, "Target")
	dev, err = eng.BulkCreateRegisterDefinitions(engine.BulkCreateRegisterDefinitionsRequest{DeviceID: "dev1", GroupID: source.ID, StartRegister: 0, Quantity: 3, DataType: "uint16", Count: 2, NamePattern: "Point {index} at {address}"})
	if err != nil {
		t.Fatalf("BulkCreateRegisterDefinitions failed: %v", err)
	}
	defs := findGroup(t, dev, "Source").Definitions
	ids := []string{defs[0].ID, defs[2].ID}

	dev, err = eng.BulkEditRegisterDefinitions(engine.BulkEditRegisterDefinitionsRequest{DeviceID: "dev1", GroupID: source.ID, DefinitionIDs: ids, DataType: "float32", Count: 2, ByteOrder: "BADC"})
	if err != nil {
		t.Fatalf("BulkEditRegisterDefinitions failed: %v", err)
	}
	if got := findGroup(t, dev, "Source").Definitions[0].ByteOrder; got != "BADC" {
		t.Fatalf("expected byte order edit, got %q", got)
	}

	dev, err = eng.MoveRegisterDefinitions(engine.MoveRegisterDefinitionsRequest{DeviceID: "dev1", SourceGroupID: source.ID, TargetGroupID: target.ID, DefinitionIDs: []string{ids[0]}})
	if err != nil {
		t.Fatalf("MoveRegisterDefinitions failed: %v", err)
	}
	if got := len(findGroup(t, dev, "Target").Definitions); got != 1 {
		t.Fatalf("expected one moved definition, got %d", got)
	}

	dev, err = eng.DuplicateRegisterDefinitions(engine.DuplicateRegisterDefinitionsRequest{DeviceID: "dev1", SourceGroupID: target.ID, TargetGroupID: source.ID, DefinitionIDs: []string{ids[0]}, AddressOffset: 20, NamePattern: "{name} Copy {offset}"})
	if err != nil {
		t.Fatalf("DuplicateRegisterDefinitions failed: %v", err)
	}
	if got := findGroup(t, dev, "Source").Definitions[len(findGroup(t, dev, "Source").Definitions)-1].Name; got != "Point 0 at 0 Copy 20" {
		t.Fatalf("unexpected duplicate name %q", got)
	}

	dev, err = eng.BulkDeleteRegisterDefinitions(engine.BulkDeleteRegisterDefinitionsRequest{DeviceID: "dev1", GroupID: source.ID, DefinitionIDs: []string{ids[1]}})
	if err != nil {
		t.Fatalf("BulkDeleteRegisterDefinitions failed: %v", err)
	}
	for _, def := range findGroup(t, dev, "Source").Definitions {
		if def.ID == ids[1] {
			t.Fatalf("expected deleted definition to be absent")
		}
	}
}

func TestEngineMutationValidationUsesNamesAndCompatibilityRules(t *testing.T) {
	eng := engine.NewEngine()
	_ = eng.AddDevice("dev1", "conn1", 1)
	dev, _ := eng.CreateRegisterGroup(engine.CreateRegisterGroupRequest{DeviceID: "dev1", Name: "Coils", ModbusTable: engine.TableCoil})
	coils := findGroup(t, dev, "Coils")
	if _, err := eng.CreateRegisterDefinition(engine.CreateRegisterDefinitionRequest{DeviceID: "dev1", GroupID: coils.ID, Name: "Bad", Register: 0, DataType: "uint16"}); err == nil {
		t.Fatalf("expected uint16 coil definition to be rejected")
	}
	if _, err := eng.CreateRegisterDefinition(engine.CreateRegisterDefinitionRequest{DeviceID: "dev1", GroupID: coils.ID, Name: "Bad", Register: 0, DataType: "bool", ByteOrder: "ABCD"}); err == nil {
		t.Fatalf("expected byte order on bool definition to be rejected")
	}
	if _, err := eng.CreateRegisterGroup(engine.CreateRegisterGroupRequest{DeviceID: "dev1", Name: "Coils", ModbusTable: engine.TableDiscreteInput}); err == nil || err.Error() != "register group \"Coils\" already exists in device dev1" {
		t.Fatalf("expected duplicate group name error, got %v", err)
	}
}

func TestEngineRejectsDuplicateGroupNamesWithinDevice(t *testing.T) {
	eng := engine.NewEngine()
	if err := eng.AddDevice("dev1", "conn1", 1); err != nil {
		t.Fatalf("AddDevice failed: %v", err)
	}
	if err := eng.AddRegisterGroup("dev1", "Holding Registers", engine.TableHoldingRegister); err != nil {
		t.Fatalf("AddRegisterGroup failed: %v", err)
	}
	if err := eng.AddRegisterGroup("dev1", "Holding Registers", engine.TableInputRegister); err == nil {
		t.Fatalf("expected duplicate group name to be rejected")
	}
}

func TestEngineSortsDefinitionsByAddress(t *testing.T) {
	eng := engine.NewEngine()
	if err := eng.AddDevice("dev1", "conn1", 1); err != nil {
		t.Fatalf("AddDevice failed: %v", err)
	}
	if err := eng.AddRegisterGroup("dev1", "holding", engine.TableHoldingRegister); err != nil {
		t.Fatalf("AddRegisterGroup failed: %v", err)
	}
	if err := eng.AddRegisterDefinition("dev1", "holding", 20, 1, "uint16"); err != nil {
		t.Fatalf("AddRegisterDefinition failed: %v", err)
	}
	if err := eng.AddRegisterDefinition("dev1", "holding", 10, 1, "uint16"); err != nil {
		t.Fatalf("AddRegisterDefinition failed: %v", err)
	}
	dev, err := eng.GetDeviceConfig("dev1")
	if err != nil {
		t.Fatalf("GetDeviceConfig failed: %v", err)
	}
	defs := findGroup(t, dev, "holding").Definitions
	if defs[0].Register != 10 || defs[1].Register != 20 {
		t.Fatalf("expected definitions sorted by address, got %+v", defs)
	}
}

func TestEngineRejectsDuplicateRegisterDefinitionsInSameGroup(t *testing.T) {
	eng := engine.NewEngine()
	if err := eng.AddDevice("dev1", "conn1", 1); err != nil {
		t.Fatalf("Failed to add device: %v", err)
	}
	if err := eng.AddRegisterGroup("dev1", "holding", engine.TableHoldingRegister); err != nil {
		t.Fatalf("Failed to add group: %v", err)
	}
	if err := eng.AddRegisterDefinition("dev1", "holding", 10, 1, "uint16"); err != nil {
		t.Fatalf("Failed to add first register definition: %v", err)
	}

	err := eng.AddRegisterDefinition("dev1", "holding", 10, 1, "uint16")
	if err == nil {
		t.Fatalf("expected duplicate register definition to be rejected")
	}
	if got, want := err.Error(), "register definition at 10 spanning 1 register(s) overlaps existing definition at 10 spanning 1 register(s) in group holding"; got != want {
		t.Fatalf("unexpected error message:\n got: %s\nwant: %s", got, want)
	}
}

func TestEngineRejectsOverlappingRegisterDefinitionsInSameGroup(t *testing.T) {
	eng := engine.NewEngine()
	if err := eng.AddDevice("dev1", "conn1", 1); err != nil {
		t.Fatalf("Failed to add device: %v", err)
	}
	if err := eng.AddRegisterGroup("dev1", "holding", engine.TableHoldingRegister); err != nil {
		t.Fatalf("Failed to add group: %v", err)
	}
	if err := eng.AddRegisterDefinition("dev1", "holding", 10, 1, "uint16"); err != nil {
		t.Fatalf("Failed to add first register definition: %v", err)
	}

	err := eng.AddRegisterDefinition("dev1", "holding", 10, 2, "float32")
	if err == nil {
		t.Fatalf("expected overlapping register definition to be rejected")
	}
	if got, want := err.Error(), "register definition at 10 spanning 2 register(s) overlaps existing definition at 10 spanning 1 register(s) in group holding"; got != want {
		t.Fatalf("unexpected error message:\n got: %s\nwant: %s", got, want)
	}
}

func TestEngineAllowsAdjacentRegisterDefinitionsInSameGroup(t *testing.T) {
	eng := engine.NewEngine()
	if err := eng.AddDevice("dev1", "conn1", 1); err != nil {
		t.Fatalf("Failed to add device: %v", err)
	}
	if err := eng.AddRegisterGroup("dev1", "holding", engine.TableHoldingRegister); err != nil {
		t.Fatalf("Failed to add group: %v", err)
	}
	if err := eng.AddRegisterDefinition("dev1", "holding", 10, 1, "uint16"); err != nil {
		t.Fatalf("Failed to add first register definition: %v", err)
	}
	if err := eng.AddRegisterDefinition("dev1", "holding", 11, 1, "uint16"); err != nil {
		t.Fatalf("expected adjacent register definition to be accepted: %v", err)
	}

	dev, err := eng.GetDeviceConfig("dev1")
	if err != nil {
		t.Fatalf("Failed to load device config: %v", err)
	}
	if got, want := len(findGroup(t, dev, "holding").Definitions), 2; got != want {
		t.Fatalf("expected adjacent definitions to be stored, got %d", got)
	}
}

func TestEnginePollsByStableDefinitionIDAndSplitsSparseOrLargeChunks(t *testing.T) {
	handler := &recordingRegisterHandler{failAddr: 9999}
	server, err := modbus.NewServer(&modbus.ServerConfiguration{URL: "tcp://127.0.0.1:55558", Timeout: 10 * time.Second, MaxClients: 1}, handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	eng := engine.NewEngine()
	if err := eng.AddConnection("conn1", "tcp://127.0.0.1:55558"); err != nil {
		t.Fatalf("AddConnection failed: %v", err)
	}
	_ = eng.AddDevice("dev1", "conn1", 1)
	dev, _ := eng.CreateRegisterGroup(engine.CreateRegisterGroupRequest{DeviceID: "dev1", Name: "Holding", ModbusTable: engine.TableHoldingRegister})
	group := findGroup(t, dev, "Holding")
	dev, _ = eng.CreateRegisterDefinition(engine.CreateRegisterDefinitionRequest{DeviceID: "dev1", GroupID: group.ID, Name: "A", Register: 0, DataType: "uint16"})
	dev, _ = eng.CreateRegisterDefinition(engine.CreateRegisterDefinitionRequest{DeviceID: "dev1", GroupID: group.ID, Name: "B", Register: 18, DataType: "uint16"})
	dev, _ = eng.CreateRegisterDefinition(engine.CreateRegisterDefinitionRequest{DeviceID: "dev1", GroupID: group.ID, Name: "C", Register: 200, DataType: "uint16"})
	group = findGroup(t, dev, "Holding")

	results, err := eng.PollDevice("dev1")
	if err != nil {
		t.Fatalf("PollDevice failed: %v", err)
	}
	if _, ok := results[group.ID][group.Definitions[0].ID]; !ok {
		t.Fatalf("expected result keyed by definition ID, got %+v", results[group.ID])
	}
	if _, ok := results[group.ID]["0"]; ok {
		t.Fatalf("did not expect address-keyed results")
	}
	if got := len(handler.requests); got != 3 {
		t.Fatalf("expected sparse and large-span definitions to split into 3 chunks, got %d: %+v", got, handler.requests)
	}
	if handler.requests[0].Addr != 0 || handler.requests[0].Quantity != 1 || handler.requests[1].Addr != 18 || handler.requests[2].Addr != 200 {
		t.Fatalf("unexpected chunks: %+v", handler.requests)
	}

	handler.failAddr = 18
	if _, err := eng.PollDevice("dev1"); err == nil {
		t.Fatalf("expected any failed chunk to fail the whole poll")
	}
}

func TestEnginePollingAllTables(t *testing.T) {
	// Start mock server
	server, err := modbus.NewServer(&modbus.ServerConfiguration{
		URL:        "tcp://127.0.0.1:55555",
		Timeout:    10 * time.Second,
		MaxClients: 1,
	}, &mockModbusHandler{})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Initialize engine
	eng := engine.NewEngine()
	err = eng.AddConnection("conn1", "tcp://127.0.0.1:55555")
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	err = eng.AddDevice("dev1", "conn1", 1)
	if err != nil {
		t.Fatalf("Failed to add device: %v", err)
	}

	// Add groups for all 4 tables
	eng.AddRegisterGroup("dev1", "g_coil", engine.TableCoil)
	eng.AddRegisterDefinition("dev1", "g_coil", 0, 5, "bool")

	eng.AddRegisterGroup("dev1", "g_di", engine.TableDiscreteInput)
	eng.AddRegisterDefinition("dev1", "g_di", 10, 5, "bool")

	eng.AddRegisterGroup("dev1", "g_hr", engine.TableHoldingRegister)
	eng.AddRegisterDefinition("dev1", "g_hr", 100, 2, "uint16")

	eng.AddRegisterGroup("dev1", "g_ir", engine.TableInputRegister)
	eng.AddRegisterDefinition("dev1", "g_ir", 200, 2, "uint16")

	results, err := eng.PollDevice("dev1")
	if err != nil {
		t.Fatalf("PollDevice failed: %v", err)
	}

	if len(results) != 4 {
		t.Errorf("Expected 4 groups, got %d", len(results))
	}

	dev, err := eng.GetDeviceConfig("dev1")
	if err != nil {
		t.Fatalf("GetDeviceConfig failed: %v", err)
	}
	if _, ok := results[findGroup(t, dev, "g_coil").ID]; !ok {
		t.Errorf("Expected g_coil results")
	}
	if _, ok := results[findGroup(t, dev, "g_hr").ID]; !ok {
		t.Errorf("Expected g_hr results")
	}
}
