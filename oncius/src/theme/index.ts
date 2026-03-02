import { createTheme } from '@mui/material/styles'

const theme = createTheme({
  palette: {
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
  },
  typography: {
    fontFamily: 'Inter, system-ui, sans-serif',
  },
})

export default theme
