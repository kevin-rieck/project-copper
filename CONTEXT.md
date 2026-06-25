# Domain Glossary

- **Connection**: The physical or network transport layer used to reach hardware. Examples include a Serial Port (e.g., `COM3`, `/dev/ttyUSB0`) or a TCP socket (e.g., `192.168.1.50:502`).
- **Device**: The logical entity communicated with over a Connection. Uniquely identified by a `Slave ID` or `Unit ID`. Multiple Devices can share a single Connection (common in Modbus RTU).
