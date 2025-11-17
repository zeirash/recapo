import { Theme } from 'theme-ui'

const theme: Theme = {
  colors: {
    primary: '#2563eb',
    secondary: '#64748b',
    accent: '#f59e0b',
    success: '#10b981',
    error: '#ef4444',
    warning: '#f59e0b',
    text: '#1f2937',
    'text.secondary': '#6b7280',
    background: '#ffffff',
    'background.secondary': '#f9fafb',
    backgroundLight: '#f3f4f6',
    border: '#e5e7eb',
  },
  fonts: {
    body: 'Inter, system-ui, sans-serif',
    heading: 'Inter, system-ui, sans-serif',
  },
  fontSizes: [12, 14, 16, 18, 20, 24, 30, 36, 48, 60, 72],
  fontWeights: {
    body: 400,
    heading: 600,
    bold: 700,
  },
  lineHeights: {
    body: 1.5,
    heading: 1.2,
  },
  space: [0, 4, 8, 16, 24, 32, 48, 64, 96, 128],
  sizes: {
    container: 1600,
  },
  radii: {
    small: 4,
    medium: 8,
    large: 12,
    round: '50%',
  },
  shadows: {
    small: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    medium: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
    large: '0 10px 15px -3px rgba(0, 0, 0, 0.1)',
  },
  buttons: {
    primary: {
      bg: 'primary',
      color: 'white',
      border: 'none',
      borderRadius: 'medium',
      px: 4,
      py: 2,
      fontSize: 1,
      fontWeight: 'heading',
      cursor: 'pointer',
      '&:hover': {
        bg: '#1d4ed8',
      },
      '&:disabled': {
        bg: 'secondary',
        cursor: 'not-allowed',
      },
    },
    secondary: {
      bg: 'transparent',
      color: 'primary',
      border: '2px solid',
      borderColor: 'primary',
      borderRadius: 'medium',
      px: 4,
      py: 2,
      fontSize: 1,
      fontWeight: 'heading',
      cursor: 'pointer',
      '&:hover': {
        bg: 'primary',
        color: 'white',
      },
    },
    small: {
      px: 3,
      py: 1,
      fontSize: 0,
    },
    large: {
      px: 6,
      py: 3,
      fontSize: 2,
    },
  },
  forms: {
    input: {
      border: '1px solid',
      borderColor: 'border',
      borderRadius: 'medium',
      px: 3,
      py: 2,
      fontSize: 1,
      '&:focus': {
        outline: 'none',
        borderColor: 'primary',
        boxShadow: '0 0 0 3px rgba(37, 99, 235, 0.1)',
      },
    },
    textarea: {
      border: '1px solid',
      borderColor: 'border',
      borderRadius: 'medium',
      px: 3,
      py: 2,
      fontSize: 1,
      '&:focus': {
        outline: 'none',
        borderColor: 'primary',
        boxShadow: '0 0 0 3px rgba(37, 99, 235, 0.1)',
      },
    },
  },
  cards: {
    primary: {
      bg: 'background',
      border: '1px solid',
      borderColor: 'border',
      borderRadius: 'large',
      p: 4,
      boxShadow: 'small',
    },
  },
  breakpoints: ['40em', '52em', '64em'],
}

export default theme
