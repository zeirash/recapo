import { createTheme } from '@mui/material/styles'

export const createAppTheme = (mode: 'light' | 'dark') =>
  createTheme({
    palette: {
      mode,
      primary: {
        main: '#3b82f6',
      },
      error: {
        main: '#ef4444',
        dark: '#b91c1c',
        light: '#fef2f2',
      },
      success: {
        main: '#10b981',
      },
      grey: {
        50: '#f9fafb',
        100: '#f3f4f6',
        200: '#e5e7eb',
        500: '#6b7280',
        800: '#1f2937',
      },
      background: mode === 'dark'
        ? { default: '#0f172a', paper: '#1e293b' }
        : { default: '#f9fafb', paper: '#ffffff' },
      ...(mode === 'light' && {
        action: {
          hover: '#f9fafb',
          selected: '#f3f4f6',
        },
      }),
      ...(mode === 'dark' && {
        text: {
          primary: '#f1f5f9',
          secondary: '#94a3b8',
        },
      }),
    },
    typography: {
      fontFamily: 'Inter, system-ui, sans-serif',
    },
  })

export default createAppTheme('light')
