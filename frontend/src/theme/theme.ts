import { createTheme, alpha } from '@mui/material/styles';

// Warm amber/orange professional palette
const BRAND = {
  primary: '#e67e22',      // warm amber-orange
  primaryDark: '#d35400',  // deeper orange
  primaryLight: '#f0a04b', // lighter amber
  secondary: '#2c3e50',    // slate for contrast
};

export const getTheme = (mode: 'light' | 'dark') =>
  createTheme({
    palette: {
      mode,
      ...(mode === 'light'
        ? {
            primary: { main: BRAND.primary, dark: BRAND.primaryDark, light: BRAND.primaryLight },
            secondary: { main: BRAND.secondary },
            background: { default: '#faf7f4', paper: '#ffffff' },
            divider: alpha(BRAND.primary, 0.12),
          }
        : {
            primary: { main: '#f0a04b', dark: BRAND.primary, light: '#f5c079' },
            secondary: { main: '#8eaccd' },
            background: { default: '#1a1714', paper: '#252118' },
            divider: alpha('#f0a04b', 0.15),
          }),
    },
    typography: {
      fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
      fontSize: 13,
    },
    shape: { borderRadius: 8 },
    components: {
      MuiButton: {
        defaultProps: { size: 'small' },
        styleOverrides: { root: { textTransform: 'none', fontWeight: 600 } },
      },
      MuiTableCell: {
        styleOverrides: { root: { padding: '8px 12px', fontSize: '0.8125rem' } },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            boxShadow:
              mode === 'light'
                ? '0 1px 4px rgba(230,126,34,0.08), 0 0 1px rgba(0,0,0,0.06)'
                : '0 1px 4px rgba(0,0,0,0.3), 0 0 1px rgba(240,160,75,0.1)',
            borderWidth: 1,
            borderStyle: 'solid',
            borderColor:
              mode === 'light'
                ? alpha(BRAND.primary, 0.08)
                : alpha('#f0a04b', 0.1),
          },
        },
      },
      MuiChip: {
        defaultProps: { size: 'small' },
      },
      MuiAccordion: {
        styleOverrides: {
          root: {
            borderWidth: 1,
            borderStyle: 'solid',
            borderColor:
              mode === 'light'
                ? alpha(BRAND.primary, 0.1)
                : alpha('#f0a04b', 0.1),
            '&:before': { display: 'none' },
          },
        },
      },
    },
  });
