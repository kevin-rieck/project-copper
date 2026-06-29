import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, vi, test, expect } from 'vitest';
import App from './App';
import { ToastProvider } from './ToastContext';
import * as BackendApp from '../wailsjs/go/main/App';

vi.mock('../wailsjs/go/main/App', () => ({
  AddDeviceWithDefaults: vi.fn().mockResolvedValue(null),
  SaveConfig: vi.fn().mockResolvedValue(null),
  LoadConfig: vi.fn().mockResolvedValue({ activeDeviceID: 'loaded_device', deviceIDs: ['loaded_device'] }),
  GetDeviceConfig: vi.fn().mockResolvedValue({
    id: 'loaded_device',
    conn_id: 'loaded_conn',
    slave_id: 3,
    groups: {
      holding_regs: { id: 'holding_regs', modbus_table: 3, definitions: [] },
    },
  }),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

test('clicking Save exports the current configuration through the backend file dialog', async () => {
  render(
    <ToastProvider>
      <App />
    </ToastProvider>
  );

  await userEvent.click(screen.getByRole('button', { name: /Save/i }));

  await waitFor(() => {
    expect(BackendApp.SaveConfig).toHaveBeenCalledOnce();
  });
});

test('clicking Load imports a configuration and refreshes the Register Browser device', async () => {
  render(
    <ToastProvider>
      <App />
    </ToastProvider>
  );

  await userEvent.click(screen.getByRole('button', { name: /Load/i }));

  await waitFor(() => {
    expect(BackendApp.LoadConfig).toHaveBeenCalledOnce();
    expect(BackendApp.GetDeviceConfig).toHaveBeenCalledWith('loaded_device');
  });
  expect(screen.getByRole('heading', { name: /Register Browser/i })).toBeInTheDocument();
});

test('loading a configuration refreshes the Register Browser even when the active device id is unchanged', async () => {
  const activeDeviceID = 'tcp://127.0.0.1:5020_1';
  vi.mocked(BackendApp.LoadConfig).mockResolvedValueOnce({
    activeDeviceID,
    deviceIDs: [activeDeviceID],
  });

  render(
    <ToastProvider>
      <App />
    </ToastProvider>
  );

  await userEvent.click(screen.getByRole('button', { name: /Add Device/i }));
  const addDeviceButtons = screen.getAllByRole('button', { name: /Add Device/i });
  await userEvent.click(addDeviceButtons[0]);
  await userEvent.click(screen.getByRole('button', { name: /Register Browser/i }));

  await waitFor(() => {
    expect(BackendApp.GetDeviceConfig).toHaveBeenCalledWith(activeDeviceID);
  });
  vi.mocked(BackendApp.GetDeviceConfig).mockClear();

  await userEvent.click(screen.getByRole('button', { name: /Load/i }));

  await waitFor(() => {
    expect(BackendApp.GetDeviceConfig).toHaveBeenCalledWith(activeDeviceID);
  });
});

test('adding a device initializes backend defaults from the connection form', async () => {
  render(
    <ToastProvider>
      <App />
    </ToastProvider>
  );

  await userEvent.click(screen.getByRole('button', { name: /Add Device/i }));
  const addDeviceButtons = screen.getAllByRole('button', { name: /Add Device/i });
  await userEvent.click(addDeviceButtons[0]);

  await waitFor(() => {
    expect(BackendApp.AddDeviceWithDefaults).toHaveBeenCalledWith('tcp://127.0.0.1:5020', 1);
  });
});
