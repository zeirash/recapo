import { Box } from '@mui/material'
import { useTheme } from '@mui/material/styles'

const ShimmerBlock = ({ width, height, sx = {} }: { width: number | string; height: number; sx?: object }) => {
  const theme = useTheme()
  const base = theme.palette.mode === 'dark' ? '#2d3748' : '#f0f0f0'
  const highlight = theme.palette.mode === 'dark' ? '#3d4a5c' : '#e8e8e8'
  const shimmerSx = {
    background: `linear-gradient(90deg, ${base} 25%, ${highlight} 50%, ${base} 75%)`,
    backgroundSize: '400px 100%',
    animation: 'shimmer 1.4s ease-in-out infinite',
    borderRadius: '6px',
    '@keyframes shimmer': {
      '0%': { backgroundPosition: '-400px 0' },
      '100%': { backgroundPosition: '400px 0' },
    },
  }
  return <Box sx={{ width, height, flexShrink: 0, ...shimmerSx, ...sx }} />
}

const PageLoadingSkeleton = () => (
  <Box sx={{ p: '24px', width: '100%' }}>
    <ShimmerBlock width="30%" height={24} sx={{ mb: '24px' }} />
    {[...Array(5)].map((_, i) => (
      <Box key={i} sx={{ display: 'flex', gap: '16px', alignItems: 'center', py: '12px', borderBottom: '1px solid', borderColor: 'divider' }}>
        <ShimmerBlock width={40} height={40} />
        <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '8px' }}>
          <ShimmerBlock width="60%" height={14} />
          <ShimmerBlock width="40%" height={12} />
        </Box>
        <ShimmerBlock width={80} height={24} />
      </Box>
    ))}
  </Box>
)

export default PageLoadingSkeleton
