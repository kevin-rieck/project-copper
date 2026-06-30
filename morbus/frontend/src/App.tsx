import { useState, useEffect } from 'react';
import { AddDeviceWithDefaults, LoadConfig, SaveConfig } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { RegisterBrowser } from './RegisterBrowser';
import { AddDeviceModal } from './AddDeviceModal';
import { useToast } from './ToastContext';

export default function App() {
  const [modbusData, setModbusData] = useState<any>(null);
  const [deviceStatus, setDeviceStatus] = useState<"Connected" | "Offline" | "Error">("Offline");
  const [lastError, setLastError] = useState<string>("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [activeTab, setActiveTab] = useState("device_manager");
  const [watchList, setWatchList] = useState<string[]>([]);
  const [activeDeviceID, setActiveDeviceID] = useState<string | undefined>(undefined);
  const [configRevision, setConfigRevision] = useState(0);
  const { addToast } = useToast();

  const toggleWatch = (groupId: string, definitionId: string) => {
    const id = `${groupId}:${definitionId}`;
    setWatchList(prev => prev.includes(id) ? prev.filter(r => r !== id) : [...prev, id]);
  };

  useEffect(() => {
    const cancelData = EventsOn("modbusData", (eventData: any) => {
      // Don't log every single poll payload to avoid console spam
      setModbusData(eventData.data);
      setDeviceStatus("Connected");
      setLastError("");
    });

    const cancelError = EventsOn("modbusError", (errData: any) => {
      console.error("Modbus Error:", errData);
      setDeviceStatus("Error");
      setLastError(errData.error);
    });

    return () => {
      cancelData();
      cancelError();
    };
  }, []);

  const handleAddDevice = async (uri: string, slaveId: number) => {
    // The modal now handles try/catch and error display. 
    // We just need to throw if it fails, or close if it succeeds.
    await AddDeviceWithDefaults(uri, slaveId);
    
    const deviceID = `${uri}_${slaveId}`;
    setActiveDeviceID(deviceID);
    
    // If we reach here, it succeeded
    console.log(`Connected to ${uri} with slave ID ${slaveId}`);
    setIsModalOpen(false);
    
    addToast(`Connected to ${uri} (Slave ${slaveId}) successfully`, 'success');
  };

  const handleSaveConfig = async () => {
    try {
      await SaveConfig();
      addToast('Configuration saved successfully', 'success');
    } catch (err) {
      console.error('Failed to save configuration:', err);
      addToast(`Failed to save configuration: ${err}`, 'error');
    }
  };

  const handleLoadConfig = async () => {
    try {
      const result = await LoadConfig();
      if (!result || result.loaded === false) {
        return;
      }

      setModbusData(null);
      setWatchList([]);
      setActiveDeviceID(result.activeDeviceID || undefined);
      setConfigRevision((revision) => revision + 1);
      setDeviceStatus('Offline');
      setLastError('');
      setActiveTab(result.activeDeviceID ? 'register_browser' : 'device_manager');
      addToast('Configuration loaded successfully', 'success');
    } catch (err) {
      console.error('Failed to load configuration:', err);
      addToast(`Failed to load configuration: ${err}`, 'error');
    }
  };
  return (
    <>
      <AddDeviceModal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        onAddDevice={handleAddDevice} 
      />

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
            <li>
              <button 
                onClick={() => setActiveTab('device_manager')}
                className={`w-full flex items-center gap-standard-gap px-4 py-2 transition-all ${activeTab === 'device_manager' ? 'text-primary border-r-2 border-primary bg-surface-container-high opacity-80 scale-[0.99]' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-highest'}`}
              >
                <span className="material-symbols-outlined" style={activeTab === 'device_manager' ? { fontVariationSettings: "'FILL' 1" } : {}}>settings_remote</span>
                <span className="text-label-caps font-label-caps">Device Manager</span>
              </button>
            </li>
            <li>
              <button 
                onClick={() => setActiveTab('register_browser')}
                className={`w-full flex items-center gap-standard-gap px-4 py-2 transition-all ${activeTab === 'register_browser' ? 'text-primary border-r-2 border-primary bg-surface-container-high opacity-80 scale-[0.99]' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-highest'}`}
              >
                <span className="material-symbols-outlined" style={activeTab === 'register_browser' ? { fontVariationSettings: "'FILL' 1" } : {}}>table_chart</span>
                <span className="text-label-caps font-label-caps">Register Browser</span>
              </button>
            </li>
            <li>
              <button 
                onClick={() => setActiveTab('watch_list')}
                className={`w-full flex items-center gap-standard-gap px-4 py-2 transition-all ${activeTab === 'watch_list' ? 'text-primary border-r-2 border-primary bg-surface-container-high opacity-80 scale-[0.99]' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-highest'}`}
              >
                <span className="material-symbols-outlined" style={activeTab === 'watch_list' ? { fontVariationSettings: "'FILL' 1" } : {}}>visibility</span>
                <span className="text-label-caps font-label-caps">Watch List</span>
              </button>
            </li>
            <li>
              <button 
                onClick={() => setActiveTab('traffic_inspector')}
                className={`w-full flex items-center gap-standard-gap px-4 py-2 transition-all ${activeTab === 'traffic_inspector' ? 'text-primary border-r-2 border-primary bg-surface-container-high opacity-80 scale-[0.99]' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-highest'}`}
              >
                <span className="material-symbols-outlined" style={activeTab === 'traffic_inspector' ? { fontVariationSettings: "'FILL' 1" } : {}}>terminal</span>
                <span className="text-label-caps font-label-caps">Traffic Inspector</span>
              </button>
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
              <button className="w-full flex items-center gap-standard-gap text-on-surface-variant hover:text-on-surface px-4 py-2 hover:bg-surface-container-highest transition-colors duration-150">
                <span className="material-symbols-outlined">settings</span>
                <span className="text-label-caps font-label-caps">Settings</span>
              </button>
            </li>
            <li>
              <button className="w-full flex items-center gap-standard-gap text-on-surface-variant hover:text-on-surface px-4 py-2 hover:bg-surface-container-highest transition-colors duration-150">
                <span className="material-symbols-outlined">help</span>
                <span className="text-label-caps font-label-caps">Support</span>
              </button>
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
            
            {/* Status Indicator */}
            <div 
              className={`flex items-center gap-2 ml-4 px-3 py-1 rounded-full border cursor-help transition-all ${
                deviceStatus === 'Connected' ? 'bg-green-500/10 border-green-500/20 text-green-500' :
                deviceStatus === 'Error' ? 'bg-error/10 border-error/20 text-error' : 'bg-surface-container-highest border-outline-variant text-on-surface-variant'
              }`} 
              title={lastError || "No active polling errors"}
            >
              <div className={`w-2.5 h-2.5 rounded-full ${
                deviceStatus === 'Connected' ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)] animate-pulse' :
                deviceStatus === 'Error' ? 'bg-error shadow-[0_0_8px_rgba(239,68,68,0.6)]' : 'bg-outline'
              }`}></div>
              <span className="text-label-sm font-label-sm">
                {deviceStatus}
              </span>
            </div>

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
            <button
              onClick={handleLoadConfig}
              className="text-label-caps font-label-caps border border-outline-variant text-on-surface hover:border-primary hover:text-primary px-3 py-1.5 rounded transition-all"
            >
              Load
            </button>
            <button
              onClick={handleSaveConfig}
              className="text-label-caps font-label-caps border border-outline-variant text-on-surface hover:border-primary hover:text-primary px-3 py-1.5 rounded transition-all"
            >
              Save
            </button>
            <button 
              className="text-label-caps font-label-caps bg-primary text-on-primary hover:brightness-110 px-3 py-1.5 rounded transition-all" 
              style={{ WebkitAppRegion: 'no-drag' } as React.CSSProperties}
              onClick={() => setIsModalOpen(true)}
            >
              Add Device
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
                <h2 className="text-display-lg font-display-lg text-on-surface mb-1">
                  {activeTab === 'device_manager' && 'Device Manager'}
                  {activeTab === 'register_browser' && 'Register Browser'}
                  {activeTab === 'watch_list' && 'Watch List'}
                  {activeTab === 'traffic_inspector' && 'Traffic Inspector'}
                </h2>
                <p className="text-body-sm font-body-sm text-on-surface-variant">
                  {activeTab === 'device_manager' && 'Monitor and manage connected Modbus nodes.'}
                  {activeTab === 'register_browser' && 'Explore and edit register maps for selected devices.'}
                  {activeTab === 'watch_list' && 'Track specific registers across multiple devices in real-time.'}
                  {activeTab === 'traffic_inspector' && 'View raw Modbus frames and decode network traffic.'}
                </p>
              </div>
            </div>

            {/* Dynamic Content */}
            <div className="h-[calc(100vh-140px)] w-full border border-outline-variant rounded overflow-hidden bg-surface-container-lowest">
              {activeTab === 'register_browser' && (
                <RegisterBrowser 
                  key={`${activeDeviceID ?? 'no-device'}:${configRevision}`}
                  data={modbusData || {}} 
                  deviceID={activeDeviceID}
                  watchList={watchList} 
                  onToggleWatch={toggleWatch} 
                />
              )}
              {activeTab === 'device_manager' && (
                <div className="p-6">
                  {!activeDeviceID ? (
                    <div className="flex flex-col items-center justify-center h-full text-on-surface-variant opacity-60 pt-20">
                      <span className="material-symbols-outlined text-4xl mb-4">settings_input_component</span>
                      <p className="text-center text-lg mb-4">No devices connected.</p>
                      <button 
                        onClick={() => setIsModalOpen(true)}
                        className="bg-primary text-on-primary hover:brightness-110 transition-all px-4 py-2 rounded font-label-large"
                      >
                        Add Your First Device
                      </button>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      <div className="bg-surface-container border border-outline-variant rounded-xl p-4 flex flex-col gap-4 shadow-sm hover:shadow-md transition-all group">
                        <div className="flex justify-between items-start">
                          <div className="flex items-center gap-3">
                            <div className="p-2 bg-surface-container-highest rounded-lg group-hover:bg-primary/10 transition-colors">
                              <span className="material-symbols-outlined text-primary text-2xl">developer_board</span>
                            </div>
                            <div>
                              <h3 className="text-title-md font-bold text-on-surface truncate max-w-[120px]" title={activeDeviceID}>{activeDeviceID?.split('_')[0] || 'Unknown'}</h3>
                              <p className="text-body-sm text-on-surface-variant">Slave {activeDeviceID?.split('_')[1] || '1'}</p>
                            </div>
                          </div>
                          <div className={`px-2 py-0.5 rounded text-[11px] font-bold ${
                            deviceStatus === 'Connected' ? 'bg-green-500/10 text-green-500 border border-green-500/20' : 
                            'bg-error/10 text-error border border-error/20'
                          }`}>
                            {deviceStatus}
                          </div>
                        </div>
                        <div className="text-body-sm text-on-surface-variant space-y-1 bg-surface-container-lowest p-3 rounded border border-outline-variant/50">
                          <p className="flex justify-between"><span>Slave ID</span> <span className="text-on-surface font-mono">{activeDeviceID?.split('_')[1] || '1'}</span></p>
                          <p className="flex justify-between"><span>Status</span> <span className="text-on-surface truncate ml-2" title={lastError}>{lastError || 'Polling OK'}</span></p>
                        </div>
                        <button 
                          onClick={() => setActiveTab('register_browser')}
                          className="mt-2 w-full py-1.5 border border-outline-variant rounded text-sm hover:border-primary hover:text-primary hover:bg-primary/5 transition-all text-on-surface-variant font-label-large"
                        >
                          View Registers
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              )}
              {activeTab === 'watch_list' && (
                <div className="flex flex-col h-full bg-background">
                  <div className="h-12 border-b border-outline-variant bg-surface flex items-center px-4 shrink-0">
                    <span className="text-label-caps font-label-caps text-on-surface">Watched Registers ({watchList.length})</span>
                  </div>
                  <div className="flex-1 overflow-auto">
                    {watchList.length === 0 ? (
                      <div className="flex flex-col items-center justify-center h-full text-on-surface-variant opacity-60">
                        <span className="material-symbols-outlined text-4xl mb-4">visibility_off</span>
                        <p className="text-center text-lg">No registers watched.</p>
                        <p className="text-center text-sm mt-1">Go to the Register Browser and click the star icon next to a register.</p>
                        <button 
                          onClick={() => setActiveTab('register_browser')}
                          className="mt-4 border border-outline-variant px-4 py-2 rounded font-label-large hover:text-primary hover:border-primary transition-all"
                        >
                          Go to Register Browser
                        </button>
                      </div>
                    ) : (
                      <table className="w-full text-left border-collapse whitespace-nowrap">
                        <thead className="sticky top-0 bg-surface-container z-10 border-b border-outline-variant text-label-caps font-label-caps text-on-surface-variant uppercase">
                          <tr>
                            <th className="px-4 py-2 font-normal w-10"></th>
                            <th className="px-4 py-2 font-normal">Device</th>
                            <th className="px-4 py-2 font-normal">Group</th>
                            <th className="px-4 py-2 font-normal">Address</th>
                            <th className="px-4 py-2 font-normal">Live Value</th>
                          </tr>
                        </thead>
                        <tbody className="text-sm font-data-mono">
                          {watchList.map(watchId => {
                            const [groupId, definitionId] = watchId.split(':');
                            const pollResult = modbusData && modbusData[groupId] ? modbusData[groupId][definitionId] : null;
                            
                            let displayVal = "--";
                            if (pollResult && pollResult.value !== undefined) {
                              if (typeof pollResult.value === 'boolean') {
                                displayVal = pollResult.value ? "TRUE" : "FALSE";
                              } else {
                                displayVal = pollResult.value.toString();
                              }
                            }

                            return (
                              <tr key={watchId} className="border-b border-outline-variant hover:bg-surface-container-lowest">
                                <td className="px-4 py-2.5">
                                  <button onClick={() => toggleWatch(groupId, definitionId)} className="text-primary hover:text-error transition-colors flex items-center justify-center" title="Remove from Watch List">
                                    <span className="material-symbols-outlined text-[18px]" style={{ fontVariationSettings: "'FILL' 1" }}>star</span>
                                  </button>
                                </td>
                                <td className="px-4 py-2.5 text-on-surface font-body-sm truncate max-w-[100px]" title={activeDeviceID}>{activeDeviceID?.split('_')[0] || 'Unknown'}</td>
                                <td className="px-4 py-2.5 text-on-surface font-body-sm">{groupId}</td>
                                <td className="px-4 py-2.5 text-secondary">{definitionId}</td>
                                <td className="px-4 py-2.5 font-bold text-primary">{displayVal}</td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </table>
                    )}
                  </div>
                </div>
              )}
              {activeTab === 'traffic_inspector' && (
                <div className="flex flex-col items-center justify-center h-full text-on-surface-variant opacity-60">
                  <span className="material-symbols-outlined text-4xl mb-2">construction</span>
                  <p className="text-center">Traffic Inspector coming soon.</p>
                </div>
              )}
            </div>
          </div>
        </main>
      </div>
    </>
  );
}
