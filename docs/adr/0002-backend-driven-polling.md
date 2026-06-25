# 2. Backend-Driven Polling Engine

Date: 2026-06-25

## Status

Accepted

## Context

Modbus is a polling protocol. To populate the UI (Register Browser, Watch List, etc.), the application must continuously request data from connected devices. We had to decide whether the React frontend or the Go backend should own this polling loop.

## Decision

We will implement a backend-driven polling engine in Go. The frontend will subscribe to specific registers/devices, and the Go backend will autonomously manage the Modbus polling loops (respecting timeouts, retries, and intervals) and push state changes to the frontend via Wails events.

## Consequences

- **Positive:** Leverages Go's strong concurrency model (goroutines/channels) for reliable, precise timing. Prevents UI thread blockages or complex React rendering cycles from stalling the Modbus polling. Significantly reduces the IPC (Inter-Process Communication) overhead across the Wails bridge, as only data changes or periodic updates are emitted, rather than round-trip request/responses for every poll.
- **Negative:** Increased complexity in the Go backend, which must now manage subscription state and broadcast events. The frontend must be designed to react asynchronously to incoming events rather than simply awaiting a fetch response.
