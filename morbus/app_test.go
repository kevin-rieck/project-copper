package main

import (
	"context"
	"testing"
	"time"

	"morbus/backend/engine"

	"github.com/simonvetter/modbus"
)

type appTestModbusHandler struct{}

func (h *appTestModbusHandler) HandleCoils(req *modbus.CoilsRequest) ([]bool, error) {
	return make([]bool, req.Quantity), nil
}

func (h *appTestModbusHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) ([]bool, error) {
	return make([]bool, req.Quantity), nil
}

func (h *appTestModbusHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) ([]uint16, error) {
	return make([]uint16, req.Quantity), nil
}

func (h *appTestModbusHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) ([]uint16, error) {
	return make([]uint16, req.Quantity), nil
}

func startAppTestServer(t *testing.T, uri string) *modbus.ModbusServer {
	t.Helper()

	server, err := modbus.NewServer(&modbus.ServerConfiguration{
		URL:        uri,
		Timeout:    10 * time.Second,
		MaxClients: 5,
	}, &appTestModbusHandler{})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("server.Start failed: %v", err)
	}
	t.Cleanup(func() { _ = server.Stop() })
	return server
}

func TestAppEngineDelegation(t *testing.T) {
	uri := "tcp://127.0.0.1:55601"
	startAppTestServer(t, uri)

	app := NewApp()
	app.startup(context.Background())

	err := app.AddConnection("conn1", uri)
	if err != nil {
		t.Fatalf("AddConnection failed: %v", err)
	}

	// Verify delegation occurred by ensuring no crash and error is nil
	// Since engine encapsulates its state, we just check if it propagates correctly
	err = app.AddDevice("dev1", "conn1", 1)
	if err != nil {
		t.Fatalf("AddDevice failed: %v", err)
	}
}

func TestAppAddDeviceWithDefaults(t *testing.T) {
	uri := "tcp://127.0.0.1:55602"
	startAppTestServer(t, uri)

	app := NewApp()
	app.startup(context.Background())
	defer app.Engine.StopPolling()

	err := app.AddDeviceWithDefaults(uri, 1)
	if err != nil {
		t.Fatalf("AddDeviceWithDefaults failed: %v", err)
	}

	// URI acts as the connection ID
	dev, err := app.GetDeviceConfig(uri + "_1")
	if err != nil {
		t.Fatalf("Failed to retrieve device: %v", err)
	}

	if dev.ConnID != uri {
		t.Errorf("Expected ConnID %s, got %s", uri, dev.ConnID)
	}

	if len(dev.Groups) != 4 {
		t.Fatalf("Expected 4 register groups, got %d", len(dev.Groups))
	}

	expectedGroups := map[string]engine.ModbusTableType{
		"holding_regs":    engine.TableHoldingRegister,
		"input_regs":      engine.TableInputRegister,
		"coils":           engine.TableCoil,
		"discrete_inputs": engine.TableDiscreteInput,
	}
	for groupID, table := range expectedGroups {
		group, ok := dev.Groups[groupID]
		if !ok {
			t.Fatalf("Expected group %q not found", groupID)
		}
		if group.ModbusTable != table {
			t.Errorf("Expected group %q table %v, got %v", groupID, table, group.ModbusTable)
		}
		if len(group.Definitions) != 1 {
			t.Fatalf("Expected group %q to have 1 definition, got %d", groupID, len(group.Definitions))
		}
		def := group.Definitions[0]
		expectedType := "uint16"
		if table == engine.TableCoil || table == engine.TableDiscreteInput {
			expectedType = "bool"
		}
		if def.Register != 0 || def.Count != 10 || def.DataType != expectedType {
			t.Errorf("Unexpected definition for group %q: %+v", groupID, def)
		}
	}
}

func TestAppAddDeviceWithDefaultsReuseConnection(t *testing.T) {
	uri := "tcp://127.0.0.1:55603"
	startAppTestServer(t, uri)

	app := NewApp()
	app.startup(context.Background())
	defer app.Engine.StopPolling()

	if err := app.AddDeviceWithDefaults(uri, 1); err != nil {
		t.Fatalf("First AddDeviceWithDefaults failed: %v", err)
	}

	// Keep reference to the old client
	conn1, _ := app.Engine.GetConnection(uri)

	// Add second device with same URI
	err := app.AddDeviceWithDefaults(uri, 2)
	if err != nil {
		t.Fatalf("Second AddDeviceWithDefaults failed: %v", err)
	}

	conn2, _ := app.Engine.GetConnection(uri)

	if conn1 != conn2 {
		t.Errorf("Expected Connection to be reused, but a new one was created")
	}

	_, err = app.GetDeviceConfig(uri + "_2")
	if err != nil {
		t.Fatalf("Second device not created: %v", err)
	}
}
