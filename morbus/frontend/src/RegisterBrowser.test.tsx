import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RegisterBrowser } from './RegisterBrowser';

describe('RegisterBrowser Component', () => {
    it('renders the core layout panes (Register Map, Data Table, Data Lab)', () => {
        render(<RegisterBrowser data={{}} />);
        
        // Left Pane: Register Map
        expect(screen.getByText(/Register Map/i)).toBeInTheDocument();
        
        // Middle Pane: Data Table (check for specific column headers)
        expect(screen.getByText('Address')).toBeInTheDocument();
        expect(screen.getByText('Raw (Hex)')).toBeInTheDocument();
        
        // Right Pane: Data Lab
        expect(screen.getByText(/Data Lab/i)).toBeInTheDocument();
    });

    it('displays live data values passed from the engine', async () => {
        render(<RegisterBrowser data={{ 1: { value: 1234, raw: [17562, 16384] } }} />);
        
        // It should display 1234 in the table
        expect(screen.getByText('1234')).toBeInTheDocument();
        // It should display the raw hex in the table
        expect(screen.getByText('0x449A 0x4000')).toBeInTheDocument();

        // Click on the row
        await userEvent.click(screen.getByText('1234'));

        // The Data Lab should now show the raw hex and name
        expect(screen.getByText('RAW HEX BUFFER')).toBeInTheDocument();
        expect(screen.getAllByText('0x449A 0x4000').length).toBeGreaterThan(0);

        // Change byte order to CDAB
        const select = screen.getByRole('combobox');
        await userEvent.selectOptions(select, 'CDAB');

        // The value should decode to ~2.004...
        expect(screen.getAllByText(/2\.004/).length).toBeGreaterThan(0);
    });
});
