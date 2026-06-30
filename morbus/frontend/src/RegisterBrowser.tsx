import React, { useState } from 'react';
import { decodeModbusBuffer, ByteOrder } from './utils/decoder';

import { GetDeviceConfig, AddRegisterGroup, AddRegisterDefinition, CreateRegisterDefinition, UpdateRegisterDefinition, DeleteRegisterDefinition, BulkCreateRegisterDefinitions, BulkDeleteRegisterDefinitions, BulkEditRegisterDefinitions, MoveRegisterDefinitions, DuplicateRegisterDefinitions } from '../wailsjs/go/main/App';
import { engine } from '../wailsjs/go/models';

interface RegisterBrowserProps {
    data: Record<string, any>;
    deviceID?: string;
    watchList?: string[];
    onToggleWatch?: (groupId: string, definitionId: string) => void;
}

export const RegisterBrowser: React.FC<RegisterBrowserProps> = ({ data, deviceID = "dev1", watchList = [], onToggleWatch }) => {
    const [deviceConfig, setDeviceConfig] = useState<engine.Device | null>(null);
    const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
    const [selectedDefinitionId, setSelectedDefinitionId] = useState<string | null>(null);
    const [byteOrders, setByteOrders] = useState<Record<string, ByteOrder>>({});

    const [showAddGroup, setShowAddGroup] = useState(false);
    const [newGroupName, setNewGroupName] = useState('');
    const [newGroupTable, setNewGroupTable] = useState(3);

    const [showAddDef, setShowAddDef] = useState(false);
    const [newDefAddress, setNewDefAddress] = useState('');
    const [newDefType, setNewDefType] = useState('uint16');
    const [selectedDefinitionIds, setSelectedDefinitionIds] = useState<string[]>([]);
    const [showBulkAdd, setShowBulkAdd] = useState(false);
    const [bulkStart, setBulkStart] = useState('');
    const [bulkQuantity, setBulkQuantity] = useState('');
    const [bulkNamePattern, setBulkNamePattern] = useState('Register {address}');

    const modbusTableLabels: Record<number, string> = {
        0: 'Coils (0x)',
        1: 'Discrete Inputs (1x)',
        3: 'Holding Registers (4x)',
        4: 'Input Registers (3x)'
    };

    const groupsFor = (config: any): any[] => {
        if (!config?.groups) return [];
        return Array.isArray(config.groups) ? config.groups : Object.values(config.groups);
    };

    const loadConfig = async () => {
        try {
            const config = await GetDeviceConfig(deviceID);
            setDeviceConfig(config as any);
            const groups = groupsFor(config);
            if (groups.length > 0 && !selectedGroupId) {
                setSelectedGroupId(groups[0].id || groups[0].name);
            }
        } catch (err) {
            console.error("Failed to load device config:", err);
        }
    };

    React.useEffect(() => {
        loadConfig();
    }, [deviceID]);

    const handleAddGroup = async () => {
        try {
            if (!newGroupName) return;
            await AddRegisterGroup(deviceID, newGroupName, newGroupTable);
            setShowAddGroup(false);
            setNewGroupName('');
            await loadConfig();
        } catch (err) {
            console.error("Add group error:", err);
            alert("Failed to add group: " + err);
        }
    };

    const handleAddDef = async () => {
        try {
            if (!newDefAddress || !selectedGroupId) return;
            const addr = parseInt(newDefAddress, 10);
            const count = newDefType === 'float32' ? 2 : 1;
            await CreateRegisterDefinition({ device_id: deviceID, group_id: selectedGroupId, name: `Register ${addr}`, register: addr, count, data_type: newDefType, byte_order: newDefType === 'float32' ? 'ABCD' : '' } as any);
            setShowAddDef(false);
            setNewDefAddress('');
            await loadConfig();
        } catch (err) {
            console.error("Add def error:", err);
            alert("Failed to add register: " + err);
        }
    };

    // Active definitions to render
    let definitions: any[] = [];
    let groupName = "Unknown";
    
    if (deviceConfig && selectedGroupId) {
        const group = groupsFor(deviceConfig).find((candidate: any) => candidate.id === selectedGroupId || candidate.name === selectedGroupId);
        if (group) {
            definitions = group.definitions || [];
            groupName = group.name || group.id;
        }
    }

    const displayDefinitions: any[] = definitions.map(d => ({
        ...d,
        id: d.id || `${d.register}`,
        name: d.name || (d.data_type === 'bool' ? `Address ${d.register}` : `Register ${d.register}`),
        type: d.data_type === 'bool' ? 'Bool' : d.data_type === 'float32' ? 'Float32' : 'UInt16',
        range: d.count > 1 ? `${d.register}-${d.register + d.count - 1}` : `${d.register}`,
    }));

    const selectedGroupData = selectedGroupId && data ? data[selectedGroupId] : null;
    const selectedDef = selectedDefinitionId !== null ? displayDefinitions.find(d => d.id === selectedDefinitionId) : null;
    const selectedData = selectedDefinitionId !== null && selectedGroupData ? (selectedGroupData[selectedDefinitionId] || selectedGroupData[selectedDef?.register]) : null;
    const selectedByteOrder = selectedDefinitionId !== null ? (byteOrders[selectedDefinitionId] || selectedDef?.byte_order || 'ABCD') : 'ABCD';

    const handleEditDefinition = async (def: any) => {
        const name = window.prompt('Data point name', def.name || '');
        if (name === null || !selectedGroupId) return;
        await UpdateRegisterDefinition({ device_id: deviceID, group_id: selectedGroupId, definition_id: def.id, name, register: def.register, count: def.count, data_type: def.data_type, byte_order: def.byte_order || '' } as any);
        await loadConfig();
    };

    const handleDeleteDefinition = async (def: any) => {
        if (!selectedGroupId || !window.confirm(`Delete ${def.name}?`)) return;
        await DeleteRegisterDefinition({ device_id: deviceID, group_id: selectedGroupId, definition_id: def.id } as any);
        if (selectedDefinitionId === def.id) setSelectedDefinitionId(null);
        await loadConfig();
    };

    const handleBulkDelete = async () => {
        if (!selectedGroupId || selectedDefinitionIds.length === 0 || !window.confirm(`Delete ${selectedDefinitionIds.length} selected data point(s)?`)) return;
        try {
            await BulkDeleteRegisterDefinitions({ device_id: deviceID, group_id: selectedGroupId, definition_ids: selectedDefinitionIds } as any);
            setSelectedDefinitionIds([]);
            await loadConfig();
        } catch (err) {
            alert(`Failed to bulk delete: ${err}`);
        }
    };

    const handleBulkAdd = async () => {
        if (!selectedGroupId) return;
        try {
            await BulkCreateRegisterDefinitions({ device_id: deviceID, group_id: selectedGroupId, start_register: parseInt(bulkStart, 10), quantity: parseInt(bulkQuantity, 10), data_type: newDefType, name_pattern: bulkNamePattern } as any);
            setShowBulkAdd(false);
            setBulkStart('');
            setBulkQuantity('');
            await loadConfig();
        } catch (err) {
            alert(`Failed to bulk add: ${err}`);
        }
    };

    const handleBulkEdit = async () => {
        if (!selectedGroupId || selectedDefinitionIds.length === 0) return;
        const dataType = window.prompt('Bulk data type (bool, uint16, float32)', newDefType);
        if (!dataType) return;
        try {
            await BulkEditRegisterDefinitions({ device_id: deviceID, group_id: selectedGroupId, definition_ids: selectedDefinitionIds, data_type: dataType, count: dataType === 'float32' ? 2 : 1, byte_order: dataType === 'float32' ? 'ABCD' : '' } as any);
            await loadConfig();
        } catch (err) {
            alert(`Failed to bulk edit: ${err}`);
        }
    };

    const sameTableTargets = deviceConfig && selectedGroupId
        ? groupsFor(deviceConfig).filter((group: any) => group.id !== selectedGroupId && groupsFor(deviceConfig).find((candidate: any) => candidate.id === selectedGroupId)?.modbus_table === group.modbus_table)
        : [];

    const handleMoveSelected = async () => {
        if (!selectedGroupId || selectedDefinitionIds.length === 0 || sameTableTargets.length === 0) return;
        const targetName = window.prompt('Move to group', sameTableTargets[0].name || sameTableTargets[0].id);
        const target = sameTableTargets.find((group: any) => group.name === targetName || group.id === targetName);
        if (!target) return;
        try {
            await MoveRegisterDefinitions({ device_id: deviceID, source_group_id: selectedGroupId, target_group_id: target.id, definition_ids: selectedDefinitionIds } as any);
            setSelectedDefinitionIds([]);
            await loadConfig();
        } catch (err) {
            alert(`Failed to move: ${err}`);
        }
    };

    const handleDuplicateSelected = async () => {
        if (!selectedGroupId || selectedDefinitionIds.length === 0) return;
        const offset = parseInt(window.prompt('Address offset', '1') || '', 10);
        if (Number.isNaN(offset)) return;
        try {
            await DuplicateRegisterDefinitions({ device_id: deviceID, source_group_id: selectedGroupId, target_group_id: selectedGroupId, definition_ids: selectedDefinitionIds, address_offset: offset, name_pattern: '{name} Copy' } as any);
            await loadConfig();
        } catch (err) {
            alert(`Failed to duplicate: ${err}`);
        }
    };

    const allSelected = displayDefinitions.length > 0 && displayDefinitions.every(def => selectedDefinitionIds.includes(def.id));

    if (!deviceConfig) {
        return (
            <div className="flex flex-col items-center justify-center h-full w-full bg-background text-on-surface-variant">
                <span className="material-symbols-outlined text-5xl mb-4 opacity-50">electrical_services</span>
                <h2 className="text-xl font-bold text-on-surface mb-2">No Device Connected</h2>
                <p className="text-sm max-w-md text-center opacity-80">
                    You need to connect to a Modbus device before you can configure its register map. Go to the Device Manager to add a device.
                </p>
            </div>
        );
    }

    return (
        <main className="flex-1 flex overflow-hidden bg-background h-full w-full text-on-background font-body-sm">
            {/* Left Pane: Register Map Explorer */}
            <aside className="w-[260px] flex-shrink-0 border-r border-outline-variant bg-surface-container-lowest flex flex-col h-full">
                <div className="px-4 py-3 border-b border-outline-variant bg-surface-container">
                    <h2 className="text-label-caps font-label-caps text-on-surface-variant uppercase">Register Map</h2>
                </div>
                <div className="flex-1 overflow-y-auto p-2">
                    <ul className="space-y-4">
                        {[0, 1, 3, 4].map((tableType) => {
                            const tableGroups = deviceConfig ? groupsFor(deviceConfig).filter((g: any) => g.modbus_table === tableType) : [];
                            if (tableGroups.length === 0) return null;
                            
                            return (
                                <li key={tableType}>
                                    <div className="flex items-center gap-2 px-2 py-1.5 bg-surface-container-high border-l-2 border-primary rounded-r text-primary mb-1">
                                        <span className="material-symbols-outlined text-[18px]">folder_open</span>
                                        <span className="text-sm font-semibold">{modbusTableLabels[tableType]}</span>
                                    </div>
                                    <ul className="pl-6 space-y-1 border-l border-outline-variant ml-3">
                                        {tableGroups.map((group: any) => (
                                            <li 
                                                key={group.id}
                                                onClick={() => setSelectedGroupId(group.id)}
                                                className={`flex items-center gap-2 px-2 py-1 bg-surface-container hover:bg-surface-container-highest rounded cursor-pointer ${selectedGroupId === group.id ? 'text-primary' : 'text-on-surface'}`}
                                            >
                                                <span className="material-symbols-outlined text-[16px] text-secondary">memory</span>
                                                <span className="text-sm font-medium">{group.name || group.id}</span>
                                            </li>
                                        ))}
                                    </ul>
                                </li>
                            );
                        })}
                    </ul>

                    {showAddGroup ? (
                        <div className="mt-4 p-3 bg-surface-container-highest rounded border border-outline-variant shadow-sm">
                            <select 
                                value={newGroupTable}
                                onChange={(e) => setNewGroupTable(parseInt(e.target.value, 10))}
                                className="w-full bg-background border border-outline-variant rounded px-2 py-1 text-sm text-on-surface outline-none focus:border-primary mb-2"
                            >
                                <option value={0}>Coils (0x)</option>
                                <option value={1}>Discrete Inputs (1x)</option>
                                <option value={3}>Holding Registers (4x)</option>
                                <option value={4}>Input Registers (3x)</option>
                            </select>
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
                        <span className="text-label-caps font-label-caps text-on-surface">Data Points: {groupName}</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <button onClick={handleBulkDelete} disabled={selectedDefinitionIds.length === 0} className="border border-outline-variant text-xs px-3 py-1.5 rounded disabled:opacity-50">Bulk Delete</button>
                        <button onClick={handleBulkEdit} disabled={selectedDefinitionIds.length === 0} className="border border-outline-variant text-xs px-3 py-1.5 rounded disabled:opacity-50">Bulk Edit</button>
                        <button onClick={handleMoveSelected} disabled={selectedDefinitionIds.length === 0 || sameTableTargets.length === 0} className="border border-outline-variant text-xs px-3 py-1.5 rounded disabled:opacity-50">Move</button>
                        <button onClick={handleDuplicateSelected} disabled={selectedDefinitionIds.length === 0} className="border border-outline-variant text-xs px-3 py-1.5 rounded disabled:opacity-50">Duplicate</button>
                        <button onClick={() => setShowBulkAdd(true)} disabled={!selectedGroupId} className="border border-outline-variant text-xs px-3 py-1.5 rounded disabled:opacity-50">Bulk Add</button>
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
                                    <option value="bool">Bool</option>
                                    <option value="uint16">UInt16</option>
                                    <option value="float32">Float32</option>
                                </select>
                                <button aria-label="Save Register / Save Data Point" onClick={handleAddDef} className="bg-primary text-on-primary text-xs font-bold px-3 py-1.5 rounded">Save Data Point</button>
                                <button onClick={() => setShowAddDef(false)} className="border border-outline-variant text-on-surface-variant text-xs px-2 py-1.5 rounded">Cancel</button>
                            </div>
                        ) : (
                            <button 
                                aria-label="Add Register / Add Data Point"
                                onClick={() => setShowAddDef(true)}
                                disabled={!selectedGroupId}
                                className="flex items-center gap-1 bg-primary text-on-primary hover:brightness-110 transition-all text-xs font-bold px-3 py-1.5 rounded disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                <span className="material-symbols-outlined text-[16px]">add</span>
                                Add Data Point
                            </button>
                        )}
                    </div>
                </div>
                {showBulkAdd && (
                    <div className="px-4 py-2 border-b border-outline-variant bg-surface-container flex items-center gap-2">
                        <span className="text-xs text-on-surface-variant">Bulk add preview: {bulkQuantity || 0} data point(s)</span>
                        <input aria-label="Bulk start address" placeholder="Start" value={bulkStart} onChange={(e) => setBulkStart(e.target.value)} className="w-20 bg-background border border-outline-variant rounded px-2 py-1 text-sm" />
                        <input aria-label="Bulk quantity" placeholder="Qty" value={bulkQuantity} onChange={(e) => setBulkQuantity(e.target.value)} className="w-20 bg-background border border-outline-variant rounded px-2 py-1 text-sm" />
                        <input aria-label="Bulk name pattern" placeholder="Register {address}" value={bulkNamePattern} onChange={(e) => setBulkNamePattern(e.target.value)} className="w-44 bg-background border border-outline-variant rounded px-2 py-1 text-sm" />
                        <button onClick={handleBulkAdd} className="bg-primary text-on-primary text-xs font-bold px-3 py-1.5 rounded">Commit Bulk Add</button>
                        <button onClick={() => setShowBulkAdd(false)} className="border border-outline-variant text-xs px-3 py-1.5 rounded">Cancel</button>
                    </div>
                )}
                <p className="px-4 py-2 text-xs text-on-surface-variant border-b border-outline-variant">Protocol offsets are 0-based, not 40001-style references.</p>
                <div className="flex-1 overflow-auto">
                    <table className="w-full text-left border-collapse whitespace-nowrap">
                        <thead className="sticky top-0 bg-surface-container z-10 border-b border-outline-variant shadow-sm text-label-caps font-label-caps text-on-surface-variant uppercase">
                            <tr>
                                <th className="px-4 py-2 font-normal w-10"><input aria-label="Select all data points" type="checkbox" checked={allSelected} onChange={(e) => setSelectedDefinitionIds(e.target.checked ? displayDefinitions.map(def => def.id) : [])} /></th>
                                <th className="px-4 py-2 font-normal">Address / Range</th>
                                <th className="px-4 py-2 font-normal">Count</th>
                                <th className="px-4 py-2 font-normal">Name</th>
                                <th className="px-4 py-2 font-normal">Type</th>
                                <th className="px-4 py-2 font-normal">Raw Summary</th>
                                <th className="px-4 py-2 font-normal">Value Summary</th>
                                <th className="px-4 py-2 font-normal">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="text-sm font-data-mono">
                            {displayDefinitions.map((def) => {
                                const pollResult = selectedGroupData ? (selectedGroupData[def.id] || selectedGroupData[def.register]) : null;
                                const bo = byteOrders[def.id] || def.byte_order || 'ABCD';
                                let displayVal = "--";
                                if (pollResult) {
                                    if (def.type === 'Bool') {
                                        displayVal = pollResult.value ? "TRUE" : "FALSE";
                                    } else {
                                        const decodedVal = pollResult.raw ? decodeModbusBuffer(pollResult.raw, def.type, bo) : undefined;
                                        displayVal = decodedVal !== undefined ? decodedVal.toString() : "--";
                                    }
                                }
                                
                                const rawHex = pollResult?.raw 
                                    ? pollResult.raw.map((w: number) => "0x" + w.toString(16).padStart(4, '0').toUpperCase()).join(" ") 
                                    : (def.type === 'Bool' ? (pollResult ? (pollResult.value ? "1" : "0") : "-") : "0x----");

                                const watchId = `${selectedGroupId}:${def.id}`;

                                return (
                                    <tr 
                                        key={def.id} 
                                        onClick={() => setSelectedDefinitionId(def.id)}
                                        className={`border-b border-outline-variant hover:bg-surface-container-lowest cursor-pointer ${selectedDefinitionId === def.id ? 'bg-surface-container-highest border-l-2 border-l-primary' : ''}`}
                                    >
                                        <td className="px-4 py-2.5 flex items-center gap-2">
                                            <input aria-label={`Select ${def.name}`} type="checkbox" checked={selectedDefinitionIds.includes(def.id)} onClick={(e) => e.stopPropagation()} onChange={(e) => setSelectedDefinitionIds(prev => e.target.checked ? [...prev, def.id] : prev.filter(id => id !== def.id))} />
                                            {onToggleWatch && (
                                                <button 
                                                    onClick={(e) => { e.stopPropagation(); onToggleWatch(selectedGroupId!, def.id); }}
                                                    className="text-on-surface-variant hover:text-primary transition-colors flex items-center justify-center"
                                                    title={watchList?.includes(watchId) ? "Remove from Watch List" : "Add to Watch List"}
                                                >
                                                    <span className="material-symbols-outlined text-[18px]" style={watchList?.includes(watchId) ? { fontVariationSettings: "'FILL' 1", color: 'var(--primary)' } : {}}>
                                                        star
                                                    </span>
                                                </button>
                                            )}
                                        </td>
                                        <td className="px-4 py-2.5 text-secondary">{def.range}</td>
                                        <td className="px-4 py-2.5 text-on-surface-variant">{def.count}</td>
                                        <td className="px-4 py-2.5 font-body-sm text-on-surface">{def.name}</td>
                                        <td className="px-4 py-2.5"><span className="px-1.5 py-0.5 rounded bg-surface-container-highest text-on-surface-variant text-[11px] border border-outline-variant">{def.type}</span></td>
                                        <td className="px-4 py-2.5 text-on-surface-variant">{rawHex}</td>
                                        <td className="px-4 py-2.5 font-bold text-primary">{displayVal}</td>
                                        <td className="px-4 py-2.5 flex gap-2">
                                            <button onClick={(e) => { e.stopPropagation(); handleEditDefinition(def); }} className="text-primary hover:underline">Edit</button>
                                            <button onClick={(e) => { e.stopPropagation(); handleDeleteDefinition(def); }} className="text-error hover:underline">Delete</button>
                                        </td>
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
                                <p className="text-body-sm text-on-surface-variant">Address range: {selectedDef.range}</p>
                            </div>
                            
                            <div>
                                <span className="text-label-caps font-label-caps text-on-surface-variant block mb-1">RAW HEX BUFFER</span>
                                <div className="bg-surface-container-lowest border border-outline-variant rounded p-2 text-data-mono font-data-mono text-primary text-sm tracking-wider">
                                    {selectedData?.raw 
                                        ? selectedData.raw.map((w: number) => "0x" + w.toString(16).padStart(4, '0').toUpperCase()).join(" ") 
                                        : (selectedDef.type === 'Bool' ? (selectedData ? (selectedData.value ? "1" : "0") : "-") : "----")}
                                </div>
                            </div>

                            <div>
                                <span className="text-label-caps font-label-caps text-on-surface-variant block mb-1">DECODED VALUE</span>
                                <div className="text-headline-md font-headline-md text-on-surface">
                                    {selectedData 
                                        ? (selectedDef.type === 'Bool' 
                                            ? (selectedData.value ? "TRUE" : "FALSE") 
                                            : (selectedData.raw ? decodeModbusBuffer(selectedData.raw, selectedDef.type, selectedByteOrder).toString() : "--")) 
                                        : "--"}
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
                                    onChange={(e) => setByteOrders({ ...byteOrders, [selectedDefinitionId!]: e.target.value as ByteOrder })}
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
                            <p className="text-center text-sm">Select a data point<br/>to inspect its data.</p>                        </div>
                    )}
                </div>
            </aside>
        </main>
    );
};
