package engine_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/simonvetter/modbus"
	"morbus/backend/engine"
)

type mockModbusHandler struct{}

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
	if _, ok := loadedDevice.Groups["holding"]; !ok {
		t.Fatalf("expected loaded register groups to be available, got %+v", loadedDevice.Groups)
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
	if got, want := len(dev.Groups["holding"].Definitions), 2; got != want {
		t.Fatalf("expected adjacent definitions to be stored, got %d", got)
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

	if _, ok := results["g_coil"]; !ok {
		t.Errorf("Expected g_coil results")
	}
	if _, ok := results["g_hr"]; !ok {
		t.Errorf("Expected g_hr results")
	}
}
