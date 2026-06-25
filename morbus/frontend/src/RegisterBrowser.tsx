import React from 'react';

interface RegisterBrowserProps {
    data: Record<string, any>;
}

export const RegisterBrowser: React.FC<RegisterBrowserProps> = ({ data }) => {
    // Temporary hardcoded definitions until we fetch them from the backend config
    const definitions = [
        { register: 0, name: "Drive Speed Cmd", type: "UInt16" },
        { register: 1, name: "Drive Current Phase A", type: "Float32" },
    ];

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
                                <li className="flex items-center gap-2 px-2 py-1 bg-surface-container hover:bg-surface-container-highest rounded cursor-pointer text-on-surface">
                                    <span className="material-symbols-outlined text-[16px] text-secondary">memory</span>
                                    <span className="text-sm font-medium">Drive Controllers</span>
                                </li>
                            </ul>
                        </li>
                    </ul>
                </div>
            </aside>

            {/* Middle Pane: Data Table */}
            <section className="flex-1 flex flex-col h-full overflow-hidden bg-background">
                <div className="h-12 border-b border-outline-variant bg-surface flex items-center justify-between px-4 shrink-0">
                    <div className="flex items-center gap-3">
                        <span className="text-label-caps font-label-caps text-on-surface">Holding Registers: Drive Controllers</span>
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
                            {definitions.map((def) => {
                                const val = data[def.register];
                                const displayVal = val !== undefined ? val.toString() : "--";
                                return (
                                    <tr key={def.register} className="border-b border-outline-variant hover:bg-surface-container-lowest cursor-pointer">
                                        <td className="px-4 py-2.5 text-secondary">{def.register}</td>
                                        <td className="px-4 py-2.5 font-body-sm text-on-surface">{def.name}</td>
                                        <td className="px-4 py-2.5 text-on-surface-variant">0x----</td>
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
                    <div>
                        <h3 className="text-base font-semibold text-on-surface leading-tight">Drive Speed Cmd</h3>
                    </div>
                </div>
            </aside>
        </main>
    );
};
