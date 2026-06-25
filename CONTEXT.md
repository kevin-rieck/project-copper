# Domain Glossary

- **Connection**: The physical or network transport layer used to reach hardware. Examples include a Serial Port (e.g., `COM3`, `/dev/ttyUSB0`) or a TCP socket (e.g., `192.168.1.50:502`).
- **Device**: The logical entity communicated with over a Connection. Uniquely identified by a `Slave ID` or `Unit ID`. Contains the default `ByteOrder` (Endianness) for decoding data. Multiple Devices can share a single Connection (common in Modbus RTU).
- **ModbusTable**: One of the four primary Modbus memory regions: Coils (0x), Discrete Inputs (1x), Input Registers (3x), or Holding Registers (4x).
- **RegisterGroup**: A logical collection of registers (e.g. "Drive Controllers") that are typically polled and managed together. Forms the leaf nodes of the Register Map tree. A RegisterGroup strictly belongs to exactly one ModbusTable.
- **RegisterDefinition**: The configuration for an individual data point. It defines the starting address, name, data type (e.g., Float32, UInt16), and how to decode the raw bytes. It can optionally override the Device's default `ByteOrder`. Overlapping RegisterDefinitions are permitted to allow multiple interpretations of the same memory block. Note: We use this instead of "Watch" to avoid collision with the UI's "Watch List" feature.
