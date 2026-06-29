package engine_test

import (
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
