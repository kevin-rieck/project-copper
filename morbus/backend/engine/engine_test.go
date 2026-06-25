package engine

import (
	"testing"
	"time"

	"github.com/simonvetter/modbus"
)

type mockServerHandler struct {
	readRequests int
}

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
	h.readRequests++
	res = make([]uint16, req.Quantity)
	
	for i := uint16(0); i < req.Quantity; i++ {
		addr := req.Addr + i
		if addr == 1 { // 40002 -> offset 1
			res[i] = 42
		}
		if addr == 5 { // 40006 -> offset 5
			res[i] = 99
		}
	}
	return res, nil
}

func TestEngineBlockPolling(t *testing.T) {
	handler := &mockServerHandler{}
	server, err := modbus.NewServer(&modbus.ServerConfiguration{
		URL:     "tcp://localhost:5505",
		Timeout: 10 * time.Second,
	}, handler)
	if err != nil {
		t.Fatalf("Failed to create mock server: %v", err)
	}
	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	eng := NewEngine()
	eng.AddConnection("conn1", "tcp://localhost:5505")
	eng.AddDevice("dev1", "conn1", 1)
	
	// Create a group for holding registers (4x)
	err = eng.AddRegisterGroup("dev1", "group1", modbus.HOLDING_REGISTER)
	if err != nil {
		t.Fatalf("AddRegisterGroup failed: %v", err)
	}

	// Add two definitions that are spread apart (offset 1 and offset 5)
	// 40002
	eng.AddRegisterDefinition("dev1", "group1", 1, 1, "uint16")
	// 40006
	eng.AddRegisterDefinition("dev1", "group1", 5, 1, "uint16")

	res, err := eng.PollDevice("dev1")
	if err != nil {
		t.Fatalf("PollDevice failed: %v", err)
	}

	// Verify we got both answers correctly
	if res[1].Value != uint16(42) {
		t.Errorf("Expected 42 at offset 1, got %v", res[1].Value)
	}
	if len(res[1].Raw) != 1 || res[1].Raw[0] != 42 {
		t.Errorf("Expected Raw to be [42], got %v", res[1].Raw)
	}

	if res[5].Value != uint16(99) {
		t.Errorf("Expected 99 at offset 5, got %v", res[5].Value)
	}
	if len(res[5].Raw) != 1 || res[5].Raw[0] != 99 {
		t.Errorf("Expected Raw to be [99], got %v", res[5].Raw)
	}

	// The crucial test: Did it execute in exactly ONE block read?
	if handler.readRequests != 1 {
		t.Errorf("Expected 1 block read request, but got %d", handler.readRequests)
	}
}
