package main

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

type simulationHandler struct {
	lock             sync.RWMutex
	holdingRegisters [100]uint16
	inputRegisters   [100]uint16
	coils            [100]bool
	discreteInputs   [100]bool
}

func (h *simulationHandler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if req.IsWrite {
		for i := 0; i < int(req.Quantity); i++ {
			addr := int(req.Addr) + i
			if addr < len(h.coils) {
				h.coils[addr] = req.Args[i]
			}
		}
		return req.Args, nil
	}

	res = make([]bool, req.Quantity)
	for i := 0; i < int(req.Quantity); i++ {
		addr := int(req.Addr) + i
		if addr < len(h.coils) {
			res[i] = h.coils[addr]
		}
	}
	return res, nil
}

func (h *simulationHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	res = make([]bool, req.Quantity)
	for i := 0; i < int(req.Quantity); i++ {
		addr := int(req.Addr) + i
		if addr < len(h.discreteInputs) {
			res[i] = h.discreteInputs[addr]
		}
	}
	return res, nil
}

func (h *simulationHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if req.IsWrite {
		for i := 0; i < int(req.Quantity); i++ {
			addr := int(req.Addr) + i
			if addr < len(h.holdingRegisters) {
				h.holdingRegisters[addr] = req.Args[i]
			}
		}
		return req.Args, nil
	}

	res = make([]uint16, req.Quantity)
	for i := 0; i < int(req.Quantity); i++ {
		addr := int(req.Addr) + i
		if addr < len(h.holdingRegisters) {
			res[i] = h.holdingRegisters[addr]
		}
	}
	return res, nil
}

func (h *simulationHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	res = make([]uint16, req.Quantity)
	for i := 0; i < int(req.Quantity); i++ {
		addr := int(req.Addr) + i
		if addr < len(h.inputRegisters) {
			res[i] = h.inputRegisters[addr]
		}
	}
	return res, nil
}

func main() {
	handler := &simulationHandler{}

	server, err := modbus.NewServer(&modbus.ServerConfiguration{
		URL:        "tcp://0.0.0.0:5020",
		Timeout:    30 * time.Second,
		MaxClients: 5,
	}, handler)

	if err != nil {
		fmt.Printf("failed to create server: %v\n", err)
		os.Exit(1)
	}

	err = server.Start()
	if err != nil {
		fmt.Printf("failed to start server: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Simulation server started on tcp://0.0.0.0:5020")

	// Update data loop
	go func() {
		var counter uint16 = 0
		var angle float64 = 0.0

		for {
			time.Sleep(500 * time.Millisecond)

			handler.lock.Lock()
			
			// Holding Registers
			handler.holdingRegisters[0] = counter
			handler.holdingRegisters[1] = (counter * 7) % 100
			sineVal := float32(math.Sin(angle) * 100)
			bits := math.Float32bits(sineVal)
			handler.holdingRegisters[2] = uint16(bits >> 16)
			handler.holdingRegisters[3] = uint16(bits & 0xFFFF)
			
			// Input Registers (Sensors)
			handler.inputRegisters[0] = uint16(2200 + (counter % 100)) // Voltage
			handler.inputRegisters[1] = uint16(5000 + (counter % 10))  // Frequency
			handler.inputRegisters[2] = uint16(150 + (counter % 50))   // Current
			
			// Discrete Inputs (Switches)
			handler.discreteInputs[0] = (counter % 2 == 0) // toggles fast
			handler.discreteInputs[1] = (counter % 10 < 5) // toggles slow
			handler.discreteInputs[2] = true               // always on
			
			// Coils (Relays)
			// Slowly toggle coil 9
			if counter % 20 == 0 {
				handler.coils[9] = !handler.coils[9]
			}

			handler.lock.Unlock()

			counter++
			angle += 0.1
		}
	}()

	select {} // Block forever
}
