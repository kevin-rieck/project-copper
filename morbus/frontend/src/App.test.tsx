import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi, test, expect } from 'vitest';
import App from './App';
import * as BackendApp from '../wailsjs/go/main/App';

vi.mock('../wailsjs/go/main/App', () => ({
  AddConnection: vi.fn().mockResolvedValue(null),
  AddDevice: vi.fn().mockResolvedValue(null),
  StartPolling: vi.fn().mockResolvedValue(null),
}));

test('clicking Connect button initializes backend connections', async () => {
  render(<App />);
  const connectBtn = screen.getByText('Connect');
  await userEvent.click(connectBtn);

  expect(BackendApp.AddConnection).toHaveBeenCalledWith("conn1", "tcp://192.168.1.50:502");
  expect(BackendApp.AddDevice).toHaveBeenCalledWith("dev1", "conn1", 1);
  expect(BackendApp.StartPolling).toHaveBeenCalled();
});
