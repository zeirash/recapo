import { Box } from '@mui/material'

const shimmerSx = {
  background: 'linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%)',
  backgroundSize: '400px 100%',
  animation: 'shimmer 1.4s ease-in-out infinite',
  borderRadius: '6px',
  '@keyframes shimmer': {
    '0%': { backgroundPosition: '-400px 0' },
    '100%': { backgroundPosition: '400px 0' },
  },
}

const ShimmerBlock = ({ width, height, sx = {} }: { width: number | string; height: number; sx?: object }) => (
  <Box sx={{ width, height, flexShrink: 0, ...shimmerSx, ...sx }} />
)

const PageLoadingSkeleton = () => (
  <Box sx={{ p: '24px', width: '100%' }}>
    <ShimmerBlock width="30%" height={24} sx={{ mb: '24px' }} />
    {[...Array(5)].map((_, i) => (
      <Box key={i} sx={{ display: 'flex', gap: '16px', alignItems: 'center', py: '12px', borderBottom: '1px solid', borderColor: 'grey.100' }}>
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
