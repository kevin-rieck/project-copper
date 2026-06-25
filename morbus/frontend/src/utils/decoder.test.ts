import { describe, it, expect } from 'vitest';
import { decodeModbusBuffer } from './decoder';

describe('decodeModbusBuffer', () => {
    it('decodes Float32 ABCD (Big Endian)', () => {
        // 1234.0 in hex is 0x449A 0x4000
        const raw = [0x449A, 0x4000];
        expect(decodeModbusBuffer(raw, 'Float32', 'ABCD')).toBeCloseTo(1234, 2);
    });

    it('decodes Float32 CDAB (Word Swap)', () => {
        // Swapped words: 0x4000 0x449A
        const raw = [0x4000, 0x449A];
        expect(decodeModbusBuffer(raw, 'Float32', 'CDAB')).toBeCloseTo(1234, 2);
    });

    it('decodes Float32 BADC (Byte Swap)', () => {
        // Swapped bytes: 0x9A44 0x0040
        const raw = [0x9A44, 0x0040];
        expect(decodeModbusBuffer(raw, 'Float32', 'BADC')).toBeCloseTo(1234, 2);
    });

    it('decodes Float32 DCBA (Little Endian)', () => {
        // Swapped both: 0x0040 0x9A44
        const raw = [0x0040, 0x9A44];
        expect(decodeModbusBuffer(raw, 'Float32', 'DCBA')).toBeCloseTo(1234, 2);
    });

    it('decodes UInt16 AB (Big Endian)', () => {
        const raw = [0x05DC]; // 1500
        expect(decodeModbusBuffer(raw, 'UInt16', 'ABCD')).toBe(1500);
    });

    it('decodes UInt16 BA (Byte Swap)', () => {
        const raw = [0xDC05]; // 1500 byte swapped
        expect(decodeModbusBuffer(raw, 'UInt16', 'BADC')).toBe(1500);
    });
});
