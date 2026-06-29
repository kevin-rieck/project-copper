import React, { useState } from 'react';

interface AddDeviceModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAddDevice: (uri: string, slaveId: number) => Promise<void>;
}

export const AddDeviceModal: React.FC<AddDeviceModalProps> = ({ isOpen, onClose, onAddDevice }) => {
  const [uri, setUri] = useState('tcp://127.0.0.1:5020');
  const [slaveId, setSlaveId] = useState('1');
  
  const [isLoading, setIsLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg('');
    setIsLoading(true);
    
    try {
      await onAddDevice(uri, parseInt(slaveId, 10));
    } catch (err: any) {
      setErrorMsg(err.toString());
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="bg-surface-container w-full max-w-md rounded-xl shadow-2xl border border-outline-variant overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-outline-variant flex justify-between items-center bg-surface-container-high">
          <h2 className="text-title-lg font-title-lg text-on-surface">Add Device</h2>
          <button 
            onClick={onClose}
            disabled={isLoading}
            className="text-on-surface-variant hover:text-on-surface hover:bg-surface-container-highest rounded-full p-1 transition-colors disabled:opacity-50"
          >
            <span className="material-symbols-outlined">close</span>
          </button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit} className="p-6 flex flex-col gap-6">
          {errorMsg && (
            <div className="bg-error/10 text-error p-3 rounded border border-error/20 text-body-sm font-body-sm break-all">
              {errorMsg}
            </div>
          )}

          <div className="flex flex-col gap-1.5">
            <label className="text-label-md font-label-md text-on-surface-variant">Connection URI</label>
            <input 
              type="text" 
              value={uri}
              onChange={e => setUri(e.target.value)}
              disabled={isLoading}
              className="bg-surface-container-highest border border-outline-variant rounded-md px-3 py-2 text-body-md text-on-surface focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all disabled:opacity-50"
              placeholder="e.g. tcp://192.168.1.50:502"
              required
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-label-md font-label-md text-on-surface-variant">Slave ID</label>
            <input 
              type="number" 
              value={slaveId}
              onChange={e => setSlaveId(e.target.value)}
              disabled={isLoading}
              className="bg-surface-container-highest border border-outline-variant rounded-md px-3 py-2 text-body-md text-on-surface focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all disabled:opacity-50"
              min="1"
              max="255"
              required
            />
          </div>

          {/* Footer Actions */}
          <div className="flex justify-end gap-3 mt-2">
            <button 
              type="button" 
              onClick={onClose}
              disabled={isLoading}
              className="px-4 py-2 rounded-md text-label-large font-label-large text-primary hover:bg-primary/10 transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button 
              type="submit" 
              disabled={isLoading}
              className="flex items-center gap-2 px-4 py-2 rounded-md text-label-large font-label-large bg-primary text-on-primary hover:brightness-110 shadow-sm transition-all disabled:opacity-50 disabled:hover:brightness-100"
            >
              {isLoading ? (
                <>
                  <span className="material-symbols-outlined animate-spin text-sm">progress_activity</span>
                  Connecting...
                </>
              ) : 'Add Device'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
