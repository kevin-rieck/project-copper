export type ByteOrder = 'ABCD' | 'BADC' | 'CDAB' | 'DCBA';

export function decodeModbusBuffer(raw: number[], dataType: string, byteOrder: ByteOrder = 'ABCD'): number {
    if (!raw || raw.length === 0) return 0;
    
    if (dataType === 'UInt16') {
        let word = raw[0];
        if (byteOrder === 'BADC' || byteOrder === 'DCBA') {
            word = ((word & 0xFF) << 8) | ((word >> 8) & 0xFF);
        }
        return word;
    }
    
    if (dataType === 'Float32' && raw.length >= 2) {
        const w0 = raw[0];
        const w1 = raw[1];
        
        let b0 = 0, b1 = 0, b2 = 0, b3 = 0;
        
        switch (byteOrder) {
            case 'ABCD':
                b0 = (w0 >> 8) & 0xFF; b1 = w0 & 0xFF; 
                b2 = (w1 >> 8) & 0xFF; b3 = w1 & 0xFF;
                break;
            case 'BADC':
                b0 = w0 & 0xFF; b1 = (w0 >> 8) & 0xFF; 
                b2 = w1 & 0xFF; b3 = (w1 >> 8) & 0xFF;
                break;
            case 'CDAB':
                b0 = (w1 >> 8) & 0xFF; b1 = w1 & 0xFF; 
                b2 = (w0 >> 8) & 0xFF; b3 = w0 & 0xFF;
                break;
            case 'DCBA':
                b0 = w1 & 0xFF; b1 = (w1 >> 8) & 0xFF; 
                b2 = w0 & 0xFF; b3 = (w0 >> 8) & 0xFF;
                break;
        }
        
        const buffer = new ArrayBuffer(4);
        const view = new DataView(buffer);
        view.setUint8(0, b0);
        view.setUint8(1, b1);
        view.setUint8(2, b2);
        view.setUint8(3, b3);
        
        return view.getFloat32(0, false); // Read back as Big Endian
    }
    
    return 0;
}
