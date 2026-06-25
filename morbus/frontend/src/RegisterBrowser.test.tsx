import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
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

    it('displays live data values passed from the engine', () => {
        render(<RegisterBrowser data={{ 1: 1234.5 }} />);
        
        // It should display 1234.5 in the table
        expect(screen.getByText('1234.5')).toBeInTheDocument();
    });
});
