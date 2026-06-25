import React, { useState } from 'react';
import { decodeModbusBuffer, ByteOrder } from './utils/decoder';

import { GetDeviceConfig, AddRegisterGroup, AddRegisterDefinition } from '../wailsjs/go/main/App';
import { engine } from '../wailsjs/go/models';

interface RegisterBrowserProps {
    data: Record<string, any>;
    deviceID?: string;
}

export const RegisterBrowser: React.FC<RegisterBrowserProps> = ({ data, deviceID = "dev1" }) => {
    const [deviceConfig, setDeviceConfig] = useState<engine.Device | null>(null);
    const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
    const [selectedReg, setSelectedReg] = useState<number | null>(null);
    const [byteOrders, setByteOrders] = useState<Record<number, ByteOrder>>({});

    const [showAddGroup, setShowAddGroup] = useState(false);
    const [newGroupName, setNewGroupName] = useState('');

    const [showAddDef, setShowAddDef] = useState(false);
    const [newDefAddress, setNewDefAddress] = useState('');
    const [newDefType, setNewDefType] = useState('uint16');

    const loadConfig = async () => {
        try {
            const config = await GetDeviceConfig(deviceID);
            setDeviceConfig(config as any);
            if (config && config.groups) {
                const groupIds = Object.keys(config.groups);
                if (groupIds.length > 0 && !selectedGroupId) {
                    setSelectedGroupId(groupIds[0]);
                }
            }
        } catch (err) {
            console.error("Failed to load device config:", err);
        }
    };

    React.useEffect(() => {
        loadConfig();
    }, [deviceID]);

    const handleAddGroup = async () => {
        if (!newGroupName) return;
        await AddRegisterGroup(deviceID, newGroupName, 3);
        setShowAddGroup(false);
        setNewGroupName('');
        await loadConfig();
    };

    const handleAddDef = async () => {
        if (!newDefAddress || !selectedGroupId) return;
        const addr = parseInt(newDefAddress, 10);
        const count = newDefType === 'float32' ? 2 : 1;
        await AddRegisterDefinition(deviceID, selectedGroupId, addr, count, newDefType);
        setShowAddDef(false);
        setNewDefAddress('');
        await loadConfig();
    };

    // Active definitions to render
    let definitions: any[] = [];
    let groupName = "Unknown";
    
    if (deviceConfig && selectedGroupId && deviceConfig.groups[selectedGroupId]) {
        const group = deviceConfig.groups[selectedGroupId];
        definitions = group.definitions || [];
        groupName = group.id;
    }

    // Add UI mock names to definitions since our backend currently doesn't store 'name'
    const displayDefinitions = definitions.map(d => ({
        register: d.register,
        type: d.data_type === 'uint16' ? 'UInt16' : (d.data_type === 'float32' ? 'Float32' : d.data_type),
        name: `Register ${d.register}` // placeholder until we add 'name' to the DB/config
    }));

    const selectedDef = selectedReg !== null ? displayDefinitions.find(d => d.register === selectedReg) : null;
    const selectedData = selectedReg !== null ? data[selectedReg] : null;
    const selectedByteOrder = selectedReg !== null ? (byteOrders[selectedReg] || 'ABCD') : 'ABCD';

    return (
        <main className="flex-1 flex overflow-hidden bg-background h-full w-full text-on-background font-body-sm">
            {/* Left Pane: Register Map Explorer */}
            <aside className="w-[260px] flex-shrink-0 border-r border-outline-variant bg-surface-container-lowest flex flex-col h-full">
                <div className="px-4 py-3 border-b border-outline-variant bg-surface-container">
                    <h2 className="text-label-caps font-label-caps text-on-surface-variant uppercase">Register Map</h2>
                </div>
                <div className="flex-1 overflow-y-auto p-2">
                    <ul className="space-y-1">
                        <li className="mt-2">
                            <div className="flex items-center gap-2 px-2 py-1.5 bg-surface-container-high border-l-2 border-primary rounded-r cursor-pointer text-primary">
                                <span className="material-symbols-outlined text-[18px]">folder_open</span>
                                <span className="text-sm font-semibold">Holding Registers (4x)</span>
                            </div>
                            <ul className="pl-6 mt-1 space-y-1 border-l border-outline-variant ml-3">
                                {deviceConfig && Object.values(deviceConfig.groups).map((group: any) => (
                                    <li 
                                        key={group.id}
                                        onClick={() => setSelectedGroupId(group.id)}
                                        className={`flex items-center gap-2 px-2 py-1 bg-surface-container hover:bg-surface-container-highest rounded cursor-pointer ${selectedGroupId === group.id ? 'text-primary' : 'text-on-surface'}`}
                                    >
                                        <span className="material-symbols-outlined text-[16px] text-secondary">memory</span>
                                        <span className="text-sm font-medium">{group.id}</span>
                                    </li>
                                ))}
                            </ul>
                        </li>
                    </ul>

                    {showAddGroup ? (
                        <div className="mt-4 p-3 bg-surface-container-highest rounded border border-outline-variant shadow-sm">
                            <input 
                                type="text" 
                                placeholder="Group Name (e.g. settings)" 
                                value={newGroupName}
                                onChange={(e) => setNewGroupName(e.target.value)}
                                className="w-full bg-background border border-outline-variant rounded px-2 py-1 text-sm text-on-surface outline-none focus:border-primary mb-2"
                            />
                            <div className="flex gap-2">
                                <button onClick={handleAddGroup} className="flex-1 bg-primary text-on-primary text-[11px] font-bold py-1 rounded">Save Group</button>
                                <button onClick={() => setShowAddGroup(false)} className="flex-1 border border-outline-variant text-on-surface-variant text-[11px] py-1 rounded">Cancel</button>
                            </div>
                        </div>
                    ) : (
                        <button 
                            onClick={() => setShowAddGroup(true)}
                            className="mt-4 w-full flex items-center justify-center gap-1 py-1.5 border border-dashed border-outline-variant rounded text-on-surface-variant hover:text-primary hover:border-primary transition-colors text-sm"
                        >
                            <span className="material-symbols-outlined text-[16px]">add</span>
                            <span>Add Group</span>
                        </button>
                    )}
                </div>
            </aside>

            {/* Middle Pane: Data Table */}
            <section className="flex-1 flex flex-col h-full overflow-hidden bg-background">
                <div className="h-12 border-b border-outline-variant bg-surface flex items-center justify-between px-4 shrink-0">
                    <div className="flex items-center gap-3">
                        <span className="text-label-caps font-label-caps text-on-surface">Holding Registers: {groupName}</span>
                    </div>
                    <div className="flex items-center gap-2">
                        {showAddDef ? (
                            <div className="flex items-center gap-2">
                                <input 
                                    type="number" 
                                    placeholder="Address (e.g. 0)" 
                                    value={newDefAddress}
                                    onChange={(e) => setNewDefAddress(e.target.value)}
                                    className="w-32 bg-background border border-outline-variant rounded px-2 py-1 text-sm text-on-surface outline-none focus:border-primary"
                                />
                                <select 
                                    aria-label="Data Type"
                                    value={newDefType}
                                    onChange={(e) => setNewDefType(e.target.value)}
                                    className="bg-background border border-outline-variant rounded px-2 py-1 text-sm text-on-surface outline-none focus:border-primary"
                                >
                                    <option value="uint16">UInt16</option>
                                    <option value="float32">Float32</option>
                                </select>
                                <button onClick={handleAddDef} className="bg-primary text-on-primary text-xs font-bold px-3 py-1.5 rounded">Save Register</button>
                                <button onClick={() => setShowAddDef(false)} className="border border-outline-variant text-on-surface-variant text-xs px-2 py-1.5 rounded">Cancel</button>
                            </div>
                        ) : (
                            <button 
                                onClick={() => setShowAddDef(true)}
                                disabled={!selectedGroupId}
                                className="flex items-center gap-1 bg-primary text-on-primary hover:brightness-110 transition-all text-xs font-bold px-3 py-1.5 rounded disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                <span className="material-symbols-outlined text-[16px]">add</span>
                                Add Register
                            </button>
                        )}
                    </div>
                </div>
                <div className="flex-1 overflow-auto">
                    <table className="w-full text-left border-collapse whitespace-nowrap">
                        <thead className="sticky top-0 bg-surface-container z-10 border-b border-outline-variant shadow-sm text-label-caps font-label-caps text-on-surface-variant uppercase">
                            <tr>
                                <th className="px-4 py-2 font-normal">Address</th>
                                <th className="px-4 py-2 font-normal">Name</th>
                                <th className="px-4 py-2 font-normal">Raw (Hex)</th>
                                <th className="px-4 py-2 font-normal">Value</th>
                                <th className="px-4 py-2 font-normal">Type</th>
                            </tr>
                        </thead>
                        <tbody className="text-sm font-data-mono">
                            {displayDefinitions.map((def) => {
                                const pollResult = data[def.register];
                                const bo = byteOrders[def.register] || 'ABCD';
                                const decodedVal = pollResult?.raw ? decodeModbusBuffer(pollResult.raw, def.type, bo) : undefined;
                                const displayVal = decodedVal !== undefined ? decodedVal.toString() : "--";
                                const rawHex = pollResult?.raw 
                                    ? pollResult.raw.map((w: number) => "0x" + w.toString(16).padStart(4, '0').toUpperCase()).join(" ") 
                                    : "0x----";

                                return (
                                    <tr 
                                        key={def.register} 
                                        onClick={() => setSelectedReg(def.register)}
                                        className={`border-b border-outline-variant hover:bg-surface-container-lowest cursor-pointer ${selectedReg === def.register ? 'bg-surface-container-highest border-l-2 border-l-primary' : ''}`}
                                    >
                                        <td className="px-4 py-2.5 text-secondary">{def.register}</td>
                                        <td className="px-4 py-2.5 font-body-sm text-on-surface">{def.name}</td>
                                        <td className="px-4 py-2.5 text-on-surface-variant">{rawHex}</td>
                                        <td className="px-4 py-2.5 font-bold text-primary">{displayVal}</td>
                                        <td className="px-4 py-2.5"><span className="px-1.5 py-0.5 rounded bg-surface-container-highest text-on-surface-variant text-[11px] border border-outline-variant">{def.type}</span></td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </section>

            {/* Right Pane: Data Lab Inspector */}
            <aside className="w-[320px] flex-shrink-0 border-l border-outline-variant bg-surface-container flex flex-col h-full overflow-y-auto">
                <div className="px-4 py-3 border-b border-outline-variant bg-surface-container sticky top-0 z-10 flex justify-between items-center">
                    <div className="flex items-center gap-2">
                        <span className="material-symbols-outlined text-primary text-[20px]">science</span>
                        <h2 className="text-label-caps font-label-caps text-on-surface uppercase">Data Lab</h2>
                    </div>
                </div>
                <div className="p-4 space-y-6">
                    {selectedDef ? (
                        <>
                            <div>
                                <h3 className="text-base font-semibold text-on-surface leading-tight">{selectedDef.name}</h3>
                                <p className="text-body-sm text-on-surface-variant">Address: {selectedDef.register}</p>
                            </div>
                            
                            <div>
                                <span className="text-label-caps font-label-caps text-on-surface-variant block mb-1">RAW HEX BUFFER</span>
                                <div className="bg-surface-container-lowest border border-outline-variant rounded p-2 text-data-mono font-data-mono text-primary text-sm tracking-wider">
                                    {selectedData?.raw 
                                        ? selectedData.raw.map((w: number) => "0x" + w.toString(16).padStart(4, '0').toUpperCase()).join(" ") 
                                        : "----"}
                                </div>
                            </div>

                            <div>
                                <span className="text-label-caps font-label-caps text-on-surface-variant block mb-1">DECODED VALUE</span>
                                <div className="text-headline-md font-headline-md text-on-surface">
                                    {selectedData?.raw ? decodeModbusBuffer(selectedData.raw, selectedDef.type, selectedByteOrder).toString() : "--"}
                                </div>
                                <span className="inline-block mt-1 px-1.5 py-0.5 rounded bg-surface-container-highest text-on-surface-variant text-[11px] border border-outline-variant">
                                    {selectedDef.type}
                                </span>
                            </div>

                            <div>
                                <span className="text-label-caps font-label-caps text-on-surface-variant block mb-1">BYTE ORDER</span>
                                <select 
                                    className="w-full bg-surface-container-lowest border border-outline-variant rounded px-2 py-1.5 text-sm text-on-surface outline-none focus:border-primary"
                                    value={selectedByteOrder}
                                    onChange={(e) => setByteOrders({ ...byteOrders, [selectedReg!]: e.target.value as ByteOrder })}
                                >
                                    <option value="ABCD">Big Endian (ABCD)</option>
                                    <option value="DCBA">Little Endian (DCBA)</option>
                                    <option value="BADC">Byte Swap (BADC)</option>
                                    <option value="CDAB">Word Swap (CDAB)</option>
                                </select>
                            </div>
                        </>
                    ) : (
                        <div className="flex flex-col items-center justify-center h-full text-on-surface-variant opacity-60 pt-10">
                            <span className="material-symbols-outlined text-4xl mb-2">touch_app</span>
                            <p className="text-center text-sm">Select a register<br/>to inspect its data.</p>
                        </div>
                    )}
                </div>
            </aside>
        </main>
    );
};
