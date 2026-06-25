---
name: Engineering Precision
colors:
  surface: '#13121b'
  surface-dim: '#13121b'
  surface-bright: '#393842'
  surface-container-lowest: '#0e0d16'
  surface-container-low: '#1b1b24'
  surface-container: '#1f1f28'
  surface-container-high: '#2a2933'
  surface-container-highest: '#35343e'
  on-surface: '#e4e1ee'
  on-surface-variant: '#c7c4d8'
  inverse-surface: '#e4e1ee'
  inverse-on-surface: '#302f39'
  outline: '#918fa1'
  outline-variant: '#464555'
  surface-tint: '#c3c0ff'
  primary: '#c3c0ff'
  on-primary: '#1d00a5'
  primary-container: '#4f46e5'
  on-primary-container: '#dad7ff'
  inverse-primary: '#4d44e3'
  secondary: '#b7c8e1'
  on-secondary: '#213145'
  secondary-container: '#3a4a5f'
  on-secondary-container: '#a9bad3'
  tertiary: '#ffb695'
  on-tertiary: '#571f00'
  tertiary-container: '#a44100'
  on-tertiary-container: '#ffd2be'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#e2dfff'
  primary-fixed-dim: '#c3c0ff'
  on-primary-fixed: '#0f0069'
  on-primary-fixed-variant: '#3323cc'
  secondary-fixed: '#d3e4fe'
  secondary-fixed-dim: '#b7c8e1'
  on-secondary-fixed: '#0b1c30'
  on-secondary-fixed-variant: '#38485d'
  tertiary-fixed: '#ffdbcc'
  tertiary-fixed-dim: '#ffb695'
  on-tertiary-fixed: '#351000'
  on-tertiary-fixed-variant: '#7b2f00'
  background: '#13121b'
  on-background: '#e4e1ee'
  surface-variant: '#35343e'
typography:
  display-lg:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '700'
    lineHeight: 40px
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: '600'
    lineHeight: 28px
  body-sm:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  data-mono:
    fontFamily: JetBrains Mono
    fontSize: 13px
    fontWeight: '500'
    lineHeight: 16px
  label-caps:
    fontFamily: Inter
    fontSize: 11px
    fontWeight: '700'
    lineHeight: 12px
    letterSpacing: 0.05em
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  unit: 4px
  container-padding: 16px
  gutter: 12px
  sidebar-width: 240px
  compact-gap: 4px
  standard-gap: 8px
---

## Brand & Style
This design system is engineered for high-performance industrial environments where data density and clarity are paramount. The brand personality is technical, reliable, and strictly functional, evoking the feeling of a sophisticated control room or a high-end developer environment.

The aesthetic follows an **"Engineering Dark Mode"** approach, utilizing a palette of deep charcoals and slates to reduce eye strain during long monitoring sessions. It draws from **Modern Minimalism** and **Technical Utility**, favoring crisp lines, subtle borders, and high-contrast data visualization over decorative elements. Every pixel is intentional, designed to facilitate rapid decision-making in time-sensitive industrial workflows.

## Colors
The color palette is rooted in a deep, slate-based dark mode. The primary **Electric Indigo** serves as the focal point for active states, primary actions, and successful connections. 

- **Surfaces:** Use layered grays (`#0F172A` to `#1E293B`) to create depth without relying on heavy shadows.
- **Status Colors:** These are high-chroma to ensure immediate recognition. **Vibrant Green** indicates live polling and healthy data; **Amber** signifies warnings or non-critical exceptions; **Emergency Red** is reserved for critical failures and dangerous writes.
- **Contrast:** Text and icons use a range from pure white for high-priority labels to muted slate for secondary metadata.

## Typography
The typography system uses a dual-font approach to balance readability with technical precision. 

- **Inter** is the primary typeface for the user interface, chosen for its exceptional legibility and neutral, modern tone. It handles all navigation, labels, and descriptive text.
- **JetBrains Mono** is utilized for all raw data, including hex values, register addresses, and telemetry streams. This ensures that numeric characters are distinct and align vertically in dense tables.
- **Information Density:** Font sizes are kept compact. Captions and secondary labels should leverage the `label-caps` style to provide clear section headers without occupying significant vertical space.

## Layout & Spacing
The layout follows a **Fixed Sidebar + Fluid Content** model. The sidebar remains locked at 240px to provide constant access to machine hierarchies and global views.

- **Information Density:** This design system employs a compact 4px spacing grid. Dashboards should minimize whitespace to allow as much data as possible to be visible "above the fold."
- **Grid:** Use a 12-column fluid grid for the main workspace, but rely on horizontal tabbed interfaces to manage multiple active data streams or logic controllers.
- **Responsive Behavior:** On tablet devices, the sidebar collapses into a thin icon rail. On mobile, the layout reflows into a single-column stack with data tables becoming horizontally scrollable cards.

## Elevation & Depth
In this system, depth is conveyed through **Tonal Layering** and **Low-Contrast Outlines** rather than traditional shadows. 

- **Z-Index Tiers:** The background is the lowest tier (`#0F172A`). Cards and workspace panels sit on the first elevation (`#1E293B`). Modals and dropdowns sit on the highest tier, utilizing a slightly lighter gray and a sharp, 1px border.
- **Borders:** Every container should have a subtle 1px border (`#334155`) to provide definition in the dark environment. 
- **Backdrop Blurs:** Use a 12px backdrop blur on overlays to maintain context of the underlying data while ensuring the foreground is legible.

## Shapes
The shape language is "Soft" (0.25rem), providing just enough rounding to feel modern while maintaining a disciplined, architectural structure. 

- **Precision:** Buttons, input fields, and panel corners use the base 4px radius. 
- **Icons:** Icons should be stroke-based (1.5px weight) with squared-off terminals to match the technical aesthetic.
- **Status Pips:** Small circular indicators are the only exception, used for real-time status lights to mimic hardware LEDs.

## Components
- **Buttons:** Primary buttons are solid Electric Indigo with white text. Secondary buttons use a ghost/outlined style. All buttons have a high-contrast hover state that increases brightness by 10%.
- **Data Tables:** These are the core of the system. Use a "Zebra" striping pattern with very low opacity differences. Headers are sticky and use the `label-caps` typography. Cells containing registers use `data-mono`.
- **Status Indicators:** Use a "Glow" effect for active states. A live connection should have a small green pip with a 4px outer glow of the same color.
- **Input Fields:** Inset background with a 1px border that turns Electric Indigo on focus. Use monospaced font for numeric inputs.
- **Tabs:** Underline style for active workspaces. Active tabs use a 2px Electric Indigo bottom border to signal the current view.
- **Chips:** Used for filtering tags or state flags. These are compact, using the `label-caps` font and a background color that matches the status (e.g., a muted red background for an "Error" tag).