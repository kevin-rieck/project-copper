package engine

import (
	"os"
	"testing"
	"time"

	"github.com/simonvetter/modbus"
)

func TestEngineBasicPolling(t *testing.T) {
	// 1. Setup Mock Modbus Server
	server, err := modbus.NewServer(&modbus.ServerConfiguration{
		URL:     "tcp://localhost:5502",
		Timeout: 10 * time.Second,
	}, &mockServerHandler{})
	
	if err != nil {
		t.Fatalf("Failed to create mock server: %v", err)
	}
	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	// 2. Setup Engine
	eng := NewEngine()
	
	err = eng.AddConnection("conn1", "tcp://localhost:5502")
	if err != nil {
		t.Fatalf("AddConnection failed: %v", err)
	}

	err = eng.AddDevice("dev1", "conn1", 1)
	if err != nil {
		t.Fatalf("AddDevice failed: %v", err)
	}

	err = eng.AddWatch("dev1", 1, 1, "uint16")
	if err != nil {
		t.Fatalf("AddWatch failed: %v", err)
	}

	// 3. Act: Force a poll
	results, err := eng.PollDevice("dev1")
	if err != nil {
		t.Fatalf("PollDevice failed: %v", err)
	}

	// 4. Assert
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	val, ok := results[1].(uint16)
	if !ok {
		t.Fatalf("Expected result to be uint16, got %T", results[1])
	}
	
	if val != 42 {
		t.Errorf("Expected register 1 to be 42, got %d", val)
	}
}

func TestEngineDataDecoding(t *testing.T) {
	server, err := modbus.NewServer(&modbus.ServerConfiguration{
		URL:     "tcp://localhost:5503",
		Timeout: 10 * time.Second,
	}, &mockServerHandler{})
	
	if err != nil {
		t.Fatalf("Failed to create mock server: %v", err)
	}
	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	eng := NewEngine()
	eng.AddConnection("conn1", "tcp://localhost:5503")
	eng.AddDevice("dev1", "conn1", 1)
	eng.AddWatch("dev1", 2, 2, "float32")

	results, err := eng.PollDevice("dev1")
	if err != nil {
		t.Fatalf("PollDevice failed: %v", err)
	}

	val, ok := results[2].(float32)
	if !ok {
		t.Fatalf("Expected result to be float32, got %T", results[2])
	}
	
	// Compare with tolerance
	diff := val - 3.14159
	if diff < -0.0001 || diff > 0.0001 {
		t.Errorf("Expected register 2 to be ~3.14159, got %f", val)
	}
}

// mockServerHandler implements modbus.RequestHandler
type mockServerHandler struct{}

func (h *mockServerHandler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	return nil, modbus.ErrIllegalFunction
}

func (h *mockServerHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	return nil, modbus.ErrIllegalFunction
}

func (h *mockServerHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {
	return nil, modbus.ErrIllegalFunction
}

func (h *mockServerHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	res = make([]uint16, req.Quantity)
	for i := 0; i < int(req.Quantity); i++ {
		regAddr := req.Addr + uint16(i)
		if regAddr == 1 {
			res[i] = 42
		} else if regAddr == 2 {
			res[i] = 0x4049 // High word of 3.14159
		} else if regAddr == 3 {
			res[i] = 0x0fd0 // Low word of 3.14159
		}
	}
	return res, nil
}

func TestEngineJSONPersistence(t *testing.T) {
	eng := NewEngine()
	eng.AddConnection("conn1", "tcp://192.168.1.50:502")
	eng.AddDevice("dev1", "conn1", 1)
	eng.AddWatch("dev1", 40001, 2, "float32")

	tmpFile := "test_config.json"
	defer os.Remove(tmpFile)

	err := eng.SaveConfig(tmpFile)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	eng2 := NewEngine()
	err = eng2.LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify restored state
	if len(eng2.connections) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(eng2.connections))
	}
	if eng2.connections["conn1"].URI != "tcp://192.168.1.50:502" {
		t.Errorf("Expected URI tcp://192.168.1.50:502, got %s", eng2.connections["conn1"].URI)
	}

	if len(eng2.devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(eng2.devices))
	}
	
	dev1 := eng2.devices["dev1"]
	if dev1.ConnID != "conn1" || dev1.SlaveID != 1 {
		t.Errorf("Device state incorrect: %+v", dev1)
	}
	
	if len(dev1.Watches) != 1 {
		t.Errorf("Expected 1 watch, got %d", len(dev1.Watches))
	} else if dev1.Watches[0].DataType != "float32" {
		t.Errorf("Expected float32 watch, got %s", dev1.Watches[0].DataType)
	}
}

func TestEngineContinuousPolling(t *testing.T) {
	server, err := modbus.NewServer(&modbus.ServerConfiguration{
		URL:     "tcp://localhost:5504",
		Timeout: 10 * time.Second,
	}, &mockServerHandler{})
	if err != nil {
		t.Fatalf("Failed to create mock server: %v", err)
	}
	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}

	eng := NewEngine()
	eng.AddConnection("conn1", "tcp://localhost:5504")
	eng.AddDevice("dev1", "conn1", 1)
	eng.AddWatch("dev1", 1, 1, "uint16")

	dataCh := make(chan map[uint16]interface{}, 5)
	errCh := make(chan error, 5)

	eng.OnData = func(deviceID string, results map[uint16]interface{}) {
		if deviceID == "dev1" {
			dataCh <- results
		}
	}
	eng.OnError = func(deviceID string, err error) {
		if deviceID == "dev1" {
			errCh <- err
		}
	}

	// Start polling every 50ms
	eng.StartPolling(50 * time.Millisecond)

	// Expect at least 2 successful polls
	for i := 0; i < 2; i++ {
		select {
		case res := <-dataCh:
			if res[1] != uint16(42) {
				t.Errorf("Expected 42, got %v", res[1])
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("Timeout waiting for poll %d", i+1)
		}
	}

	// Test Resilience: Stop the server, expect an error but no crash
	server.Stop()
	select {
	case err := <-errCh:
		if err == nil {
			t.Error("Expected an error after stopping server")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for error event after server stop")
	}

	// Stop polling
	eng.StopPolling()
}
