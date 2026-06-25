# 3. Optimized Block Reads for Register Groups

Date: 2026-06-25

## Status

Accepted

## Context

Our Modbus utility allows users to define a logical `RegisterGroup` (e.g., "Drive Controllers") containing multiple `RegisterDefinition` items (e.g., a Float32 at address 40001, an Int16 at 40005). 

The initial polling engine iterated over every `RegisterDefinition` sequentially, issuing a dedicated Modbus network request for each one (e.g., `ReadRegisters(40001, 2)`, then `ReadRegisters(40005, 1)`). While simple to implement, this approach scales terribly. Over high-latency connections like serial RTU, issuing 50 sequential read requests for a single group can take seconds, destroying the real-time responsiveness required for an engineering utility.

Furthermore, we decided to allow overlapping `RegisterDefinitions` (e.g., interpreting the same 16-bit word as an Int16 and simultaneously as part of a Float32). If we read them individually, we would fetch the exact same bytes from the network multiple times.

## Decision

We will implement **Optimized Block Reads** at the `RegisterGroup` level.

The polling engine will:
1. Examine all `RegisterDefinitions` within a `RegisterGroup` (which strictly belong to a single Modbus Table).
2. Calculate the minimum start address and the maximum end address to determine the contiguous span of memory required.
3. Issue a single bulk Modbus request (or chunked requests if the span exceeds the Modbus maximum of ~125 registers) to fetch the entire memory block into a raw byte buffer.
4. Iterate over the `RegisterDefinitions` in memory, decoding their specific data types directly from the cached byte buffer.

## Consequences

**Positive:**
- Drastically reduces network traffic and latency.
- Ensures all registers in a group are sampled at the exact same point in time (data coherence).
- Overlapping definitions are highly efficient, as the bytes are only fetched once.

**Negative:**
- The engine logic becomes significantly more complex, requiring buffer arithmetic and chunking algorithms.
- If a user configures a `RegisterGroup` with a massive gap (e.g., defining a register at 40001 and another at 49999 in the same group), the engine could theoretically attempt to block-read 10,000 registers. We will need to implement a chunking/gap-optimization algorithm in the future to split reads when gaps are unacceptably large.
