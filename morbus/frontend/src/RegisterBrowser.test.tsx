import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RegisterBrowser } from './RegisterBrowser';
import * as AppBindings from '../wailsjs/go/main/App';

vi.mock('../wailsjs/go/main/App', () => ({
    GetDeviceConfig: vi.fn(),
    AddRegisterGroup: vi.fn(),
    AddRegisterDefinition: vi.fn()
}));

describe('RegisterBrowser Component', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        // Setup default mock response
        (AppBindings.GetDeviceConfig as any).mockResolvedValue({
            id: "dev1",
            groups: {
                "group1": {
                    id: "group1",
                    modbus_table: 3,
                    definitions: [
                        { register: 0, count: 1, data_type: "uint16" },
                        { register: 1, count: 2, data_type: "float32" }
                    ]
                }
            }
        });
    });
    it('renders the structure without data', async () => {
        render(<RegisterBrowser data={{}} deviceID="dev1" />);
        
        await waitFor(() => {
            expect(AppBindings.GetDeviceConfig).toHaveBeenCalledWith('dev1');
            expect(screen.getByText(/Register Map/i)).toBeInTheDocument();
            expect(screen.getByText(/Data Lab/i)).toBeInTheDocument();
        });
    });

    it('displays live data values passed from the engine', async () => {
        render(<RegisterBrowser data={{ group1: { 0: { value: 1234, raw: [1234] }, 1: { value: 1234, raw: [17562, 16384] } } }} deviceID="dev1" />);
        
        // Wait for config to load
        await waitFor(() => {
            expect(screen.getAllByText('1234').length).toBeGreaterThan(0);
        });
        // It should display the raw hex in the table
        expect(screen.getByText('0x449A 0x4000')).toBeInTheDocument();

        // Click on the float row
        await userEvent.click(screen.getByText('0x449A 0x4000'));

        // The Data Lab should now show the raw hex and name
        expect(screen.getByText('RAW HEX BUFFER')).toBeInTheDocument();
        expect(screen.getAllByText('0x449A 0x4000').length).toBeGreaterThan(0);

        // Change byte order to CDAB
        const select = screen.getByRole('combobox');
        await userEvent.selectOptions(select, 'CDAB');

        // The value should decode to ~2.004...
        expect(screen.getAllByText(/2\.004/).length).toBeGreaterThan(0);
    });

    it('allows adding a new register group', async () => {
        render(<RegisterBrowser data={{}} deviceID="dev1" />);
        await waitFor(() => expect(AppBindings.GetDeviceConfig).toHaveBeenCalledWith('dev1'));

        // Click Add Group button
        const addGroupBtn = screen.getByRole('button', { name: /Add Group/i });
        await userEvent.click(addGroupBtn);

        // Fill form
        await userEvent.type(screen.getByPlaceholderText('Group Name (e.g. settings)'), 'settings');
        // Submit
        await userEvent.click(screen.getByRole('button', { name: /Save Group/i }));

        expect(AppBindings.AddRegisterGroup).toHaveBeenCalledWith('dev1', 'settings', 3);
        // It should reload config
        expect(AppBindings.GetDeviceConfig).toHaveBeenCalledTimes(2);
    });

    it('allows adding a new register definition to a group', async () => {
        render(<RegisterBrowser data={{}} deviceID="dev1" />);
        await waitFor(() => expect(AppBindings.GetDeviceConfig).toHaveBeenCalledWith('dev1'));

        // Click Add Register button
        const addRegBtn = screen.getByRole('button', { name: /Add Register/i });
        await userEvent.click(addRegBtn);

        // Fill form
        await userEvent.type(screen.getByPlaceholderText('Address (e.g. 0)'), '2');
        await userEvent.selectOptions(screen.getByRole('combobox', { name: /Data Type/i }), 'uint16');
        
        // Submit
        await userEvent.click(screen.getByRole('button', { name: /Save Register/i }));

        // count=1 for uint16
        expect(AppBindings.AddRegisterDefinition).toHaveBeenCalledWith('dev1', 'group1', 2, 1, 'uint16');
        expect(AppBindings.GetDeviceConfig).toHaveBeenCalledTimes(2);
    });
});
