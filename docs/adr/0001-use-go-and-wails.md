# 1. Use Go and Wails for Application Architecture

Date: 2026-06-25

## Status

Accepted

## Context

We are building a Modbus Control Studio that needs to communicate with local hardware via Modbus TCP/RTU. We need a rich UI based on our HTML/CSS mockups, but also need low-level system access to communicate with these devices reliably. The options considered were a pure web app with a local Node.js backend, an Electron app, or a Wails app.

## Decision

We will use Go and Wails to build the application as a standalone desktop executable.

## Consequences

- **Positive:** Go provides excellent concurrency and robust native Modbus libraries for the backend logic. Wails produces much smaller, more memory-efficient binaries compared to Electron. We can still use modern web technologies (HTML/CSS/JS/Vite) for the frontend.
- **Negative:** We trade away the massive Node.js/npm ecosystem for backend logic, and may need to write more bridging code between the Go backend and the web frontend compared to an all-JavaScript/TypeScript environment.
