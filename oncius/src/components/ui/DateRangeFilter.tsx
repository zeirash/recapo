import { Box, OutlinedInput } from '@mui/material'
import { useTheme } from '@mui/material/styles'
import { useTranslations } from 'next-intl'

type Props = {
  dateFrom: string
  dateTo: string
  onDateFromChange: (value: string) => void
  onDateToChange: (value: string) => void
}

export default function DateRangeFilter({ dateFrom, dateTo, onDateFromChange, onDateToChange }: Props) {
  const to = useTranslations('orders')
  const theme = useTheme()

  const inputSx = {
    height: 36,
    fontSize: '13px',
    borderRadius: '6px',
    width: 130,
    flexShrink: 0,
    '& .MuiOutlinedInput-notchedOutline': { borderColor: 'grey.400' },
    '& .MuiOutlinedInput-input': { padding: '6px 8px' },
    '& input::-webkit-calendar-picker-indicator': {
      filter: theme.palette.mode === 'dark' ? 'invert(1)' : 'none',
      cursor: 'pointer',
    },
  }

  const isDark = theme.palette.mode === 'dark'
  const dateInputProps = {
    style: {
      colorScheme: theme.palette.mode as 'light' | 'dark',
      color: isDark ? 'rgba(255,255,255,0.87)' : 'rgba(0,0,0,0.87)',
      WebkitTextFillColor: isDark ? 'rgba(255,255,255,0.87)' : 'rgba(0,0,0,0.87)',
    },
  }

  return (
    <>
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
        <Box sx={{ fontSize: '11px', color: 'text.secondary', px: '2px' }}>{to('dateFrom')}</Box>
        <OutlinedInput
          type="date"
          size="small"
          value={dateFrom}
          onChange={(e) => onDateFromChange(e.target.value)}
          title={to('dateFrom')}
          sx={inputSx}
          inputProps={dateInputProps}
        />
      </Box>
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
        <Box sx={{ fontSize: '11px', color: 'text.secondary', px: '2px' }}>{to('dateTo')}</Box>
        <OutlinedInput
          type="date"
          size="small"
          value={dateTo}
          onChange={(e) => onDateToChange(e.target.value)}
          title={to('dateTo')}
          sx={inputSx}
          inputProps={dateInputProps}
        />
      </Box>
    </>
  )
}
