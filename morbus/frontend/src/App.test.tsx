import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi, test, expect } from 'vitest';
import App from './App';
import { ToastProvider } from './ToastContext';
import * as BackendApp from '../wailsjs/go/main/App';

vi.mock('../wailsjs/go/main/App', () => ({
  AddDeviceWithDefaults: vi.fn().mockResolvedValue(null),
}));

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
