import { useState, useEffect } from 'react';
import { AddConnection, AddDevice, StartPolling } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

export default function App() {
  const [modbusData, setModbusData] = useState<any>(null);
  const [deviceStatus, setDeviceStatus] = useState("Offline");

  useEffect(() => {
    const cancelData = EventsOn("modbusData", (eventData: any) => {
      console.log("Received data:", eventData);
      setModbusData(eventData.data);
      setDeviceStatus("Connected");
    });

    const cancelError = EventsOn("modbusError", (errData: any) => {
      console.error("Modbus Error:", errData);
      setDeviceStatus("Error");
    });

    return () => {
      cancelData();
      cancelError();
    };
  }, []);

  const handleConnect = async () => {
    try {
      await AddConnection("conn1", "tcp://192.168.1.50:502");
      await AddDevice("dev1", "conn1", 1);
      await StartPolling();
      console.log("Connected and polling started");
    } catch (err) {
      console.error("Failed to connect", err);
    }
  };
  return (
    <>
      {/* SideNavBar */}
      <nav className="fixed left-0 top-0 h-full w-[240px] flex flex-col z-50 bg-background border-r border-outline-variant">
        {/* Header */}
        <div className="p-4 border-b border-outline-variant">
          <h1 className="text-headline-md font-headline-md font-bold text-primary">Morbus Utility</h1>
          <p className="text-label-caps font-label-caps text-on-surface-variant mt-1">Modbus TCP/RTU</p>
        </div>
        
        {/* Main Navigation */}
        <div className="flex-1 overflow-y-auto py-4">
          <ul className="space-y-1">
            {/* Active Tab */}
            <li>
              <a className="flex items-center gap-standard-gap text-primary border-r-2 border-primary bg-surface-container-high px-4 py-2 opacity-80 scale-[0.99] transition-all" href="#">
                <span className="material-symbols-outlined" style={{ fontVariationSettings: "'FILL' 1" }}>settings_remote</span>
                <span className="text-label-caps font-label-caps">Device Manager</span>
              </a>
            </li>
            {/* Inactive Tabs */}
            <li>
              <a className="flex items-center gap-standard-gap text-on-surface-variant hover:text-on-surface px-4 py-2 hover:bg-surface-container-highest transition-colors duration-150" href="#">
                <span className="material-symbols-outlined">table_chart</span>
                <span className="text-label-caps font-label-caps">Register Browser</span>
              </a>
            </li>
            <li>
              <a className="flex items-center gap-standard-gap text-on-surface-variant hover:text-on-surface px-4 py-2 hover:bg-surface-container-highest transition-colors duration-150" href="#">
                <span className="material-symbols-outlined">visibility</span>
                <span className="text-label-caps font-label-caps">Watch List</span>
              </a>
            </li>
            <li>
              <a className="flex items-center gap-standard-gap text-on-surface-variant hover:text-on-surface px-4 py-2 hover:bg-surface-container-highest transition-colors duration-150" href="#">
                <span className="material-symbols-outlined">terminal</span>
                <span className="text-label-caps font-label-caps">Traffic Inspector</span>
              </a>
            </li>
          </ul>
        </div>
        
        {/* CTA & Footer */}
        <div className="p-4 border-t border-outline-variant space-y-4">
          <button className="w-full bg-primary-container text-on-primary-container hover:brightness-110 transition-all font-label-caps text-label-caps py-2 rounded">
            Scan Network
          </button>
          <ul className="space-y-1">
            <li>
              <a className="flex items-center gap-standard-gap text-on-surface-variant hover:text-on-surface px-4 py-2 hover:bg-surface-container-highest transition-colors duration-150" href="#">
                <span className="material-symbols-outlined">settings</span>
                <span className="text-label-caps font-label-caps">Settings</span>
              </a>
            </li>
            <li>
              <a className="flex items-center gap-standard-gap text-on-surface-variant hover:text-on-surface px-4 py-2 hover:bg-surface-container-highest transition-colors duration-150" href="#">
                <span className="material-symbols-outlined">help</span>
                <span className="text-label-caps font-label-caps">Support</span>
              </a>
            </li>
          </ul>
        </div>
      </nav>

      {/* Main Workspace Area */}
      <div className="flex-1 ml-[240px] flex flex-col w-full">
        {/* TopAppBar */}
        <header className="flex justify-between items-center w-full px-container-padding h-14 bg-surface-container border-b border-outline-variant z-40">
          {/* Brand / Search Area */}
          <div className="flex items-center gap-4">
            <span className="text-headline-md font-headline-md font-black text-primary">Morbus</span>
            <div className="relative ml-4">
              <span className="material-symbols-outlined absolute left-2 top-1/2 -translate-y-1/2 text-on-surface-variant text-sm">search</span>
              <input className="bg-surface-container-highest border border-outline-variant rounded pl-8 pr-3 py-1 text-data-mono font-data-mono text-on-surface focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all w-64 placeholder-on-surface-variant" placeholder="Search devices..." type="text" />
            </div>
          </div>
          {/* Actions */}
          <div className="flex items-center gap-gutter">
            <button className="text-on-surface-variant hover:text-primary hover:bg-surface-container-highest rounded-lg p-1.5 transition-colors">
              <span className="material-symbols-outlined">notifications</span>
            </button>
            <button className="text-on-surface-variant hover:text-primary hover:bg-surface-container-highest rounded-lg p-1.5 transition-colors">
              <span className="material-symbols-outlined">wifi</span>
            </button>
            <div className="h-6 w-px bg-outline-variant mx-1"></div>
            <button className="text-label-caps font-label-caps border border-outline-variant text-on-surface hover:border-primary hover:text-primary px-3 py-1.5 rounded transition-all">
              Export
            </button>
            <button 
              className="text-label-caps font-label-caps bg-primary text-on-primary hover:brightness-110 px-3 py-1.5 rounded transition-all" 
              style={{ WebkitAppRegion: 'no-drag' } as React.CSSProperties}
              onClick={handleConnect}
            >
              Connect
            </button>
          </div>
        </header>

        {/* Content Canvas */}
        <main className="flex-1 overflow-y-auto p-container-padding bg-background relative">
          {/* Background Decorative Grid */}
          <div className="absolute inset-0 pointer-events-none" style={{ backgroundSize: '40px 40px', backgroundImage: 'linear-gradient(to right, rgba(70, 69, 85, 0.1) 1px, transparent 1px), linear-gradient(to bottom, rgba(70, 69, 85, 0.1) 1px, transparent 1px)' }}></div>
          
          <div className="relative z-10 max-w-7xl mx-auto">
            {/* Page Header */}
            <div className="flex justify-between items-end mb-6">
              <div>
                <h2 className="text-display-lg font-display-lg text-on-surface mb-1">Device Manager</h2>
                <p className="text-body-sm font-body-sm text-on-surface-variant">Monitor and manage connected Modbus nodes.</p>
              </div>
            </div>

            {/* Device Cards Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-gutter">
              
              {/* Connected Device Card 1 */}
              <div className="bg-surface-container border border-outline-variant rounded p-4 flex flex-col hover:border-primary transition-colors group relative overflow-hidden">
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h3 className="text-headline-md font-headline-md text-on-surface mb-1 truncate" title="Main PLC Rack 1">Main PLC Rack 1</h3>
                    <div className="flex items-center gap-2">
                      <span className={`w-2 h-2 rounded-full ${deviceStatus === 'Connected' ? 'bg-green-500 status-glow-active' : 'bg-red-500 status-glow-inactive'}`}></span>
                      <span className={`text-label-caps font-label-caps ${deviceStatus === 'Connected' ? 'text-on-surface-variant' : 'text-error'}`}>{deviceStatus}</span>
                    </div>
                  </div>
                  <button className="text-on-surface-variant hover:text-primary p-1 rounded hover:bg-surface-container-highest transition-colors">
                    <span className="material-symbols-outlined">more_vert</span>
                  </button>
                </div>
                <div className="space-y-3 mb-6 flex-1">
                  <div>
                    <span className="text-label-caps font-label-caps text-on-surface-variant block mb-0.5">IP ADDRESS</span>
                    <span className="text-data-mono font-data-mono text-primary">192.168.1.50</span>
                  </div>
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <span className="text-label-caps font-label-caps text-on-surface-variant block mb-0.5">PROTOCOL</span>
                      <span className="text-body-sm font-body-sm text-on-surface">Modbus TCP</span>
                    </div>
                    <div>
                      <span className="text-label-caps font-label-caps text-on-surface-variant block mb-0.5">REGISTER DATA</span>
                      <span className="text-data-mono font-data-mono text-on-surface">
                        {modbusData ? JSON.stringify(modbusData) : "Waiting..."}
                      </span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center justify-between border-t border-outline-variant pt-3 mt-auto">
                  <div className="flex gap-2">
                    <button className="text-on-surface-variant hover:text-primary p-1 rounded bg-surface-container-highest transition-colors" title="Settings">
                      <span className="material-symbols-outlined text-sm">settings</span>
                    </button>
                    <button className="text-on-surface-variant hover:text-primary p-1 rounded bg-surface-container-highest transition-colors" title="Browser">
                      <span className="material-symbols-outlined text-sm">table_chart</span>
                    </button>
                  </div>
                  <button className="text-label-caps font-label-caps text-error border border-error/50 hover:bg-error/10 px-2 py-1 rounded transition-colors">
                    Disconnect
                  </button>
                </div>
              </div>

              {/* Disconnected Device Card */}
              <div className="bg-surface-container border border-outline-variant rounded p-4 flex flex-col hover:border-outline transition-colors group relative overflow-hidden opacity-70">
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h3 className="text-headline-md font-headline-md text-on-surface mb-1 truncate" title="Sensor Array 04">Sensor Array 04</h3>
                    <div className="flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full bg-red-500 status-glow-inactive"></span>
                      <span className="text-label-caps font-label-caps text-error">Offline</span>
                    </div>
                  </div>
                  <button className="text-on-surface-variant hover:text-on-surface p-1 rounded hover:bg-surface-container-highest transition-colors">
                    <span className="material-symbols-outlined">more_vert</span>
                  </button>
                </div>
                <div className="space-y-3 mb-6 flex-1">
                  <div>
                    <span className="text-label-caps font-label-caps text-on-surface-variant block mb-0.5">SERIAL PORT</span>
                    <span className="text-data-mono font-data-mono text-on-surface-variant">COM3 (9600, 8, N, 1)</span>
                  </div>
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <span className="text-label-caps font-label-caps text-on-surface-variant block mb-0.5">PROTOCOL</span>
                      <span className="text-body-sm font-body-sm text-on-surface-variant">Modbus RTU</span>
                    </div>
                    <div>
                      <span className="text-label-caps font-label-caps text-on-surface-variant block mb-0.5">LAST SEEN</span>
                      <span className="text-data-mono font-data-mono text-on-surface-variant">2h ago</span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center justify-between border-t border-outline-variant pt-3 mt-auto">
                  <div className="flex gap-2">
                    <button className="text-on-surface-variant hover:text-on-surface p-1 rounded bg-surface-container-highest transition-colors" title="Settings">
                      <span className="material-symbols-outlined text-sm">settings</span>
                    </button>
                  </div>
                  <button className="text-label-caps font-label-caps text-primary border border-primary/50 hover:bg-primary/10 px-2 py-1 rounded transition-colors">
                    Reconnect
                  </button>
                </div>
              </div>

              {/* Add New Device */}
              <button className="bg-surface-container-low border-2 border-dashed border-outline-variant hover:border-primary hover:bg-surface-container rounded p-4 flex flex-col items-center justify-center min-h-[220px] transition-all group">
                <div className="w-12 h-12 rounded-full bg-surface-container-highest flex items-center justify-center mb-3 group-hover:bg-primary/20 group-hover:text-primary transition-colors">
                  <span className="material-symbols-outlined text-3xl">add</span>
                </div>
                <span className="text-headline-md font-headline-md text-on-surface group-hover:text-primary transition-colors">Add Device</span>
                <span className="text-body-sm font-body-sm text-on-surface-variant text-center mt-1">Configure a new Modbus TCP or RTU connection.</span>
              </button>

            </div>
          </div>
        </main>
      </div>
    </>
  );
}
