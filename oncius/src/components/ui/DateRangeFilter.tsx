import { OutlinedInput } from '@mui/material'
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
    '& .MuiOutlinedInput-input': {
      padding: '6px 8px',
      colorScheme: theme.palette.mode,
      color: theme.palette.text.primary,
      '&::-webkit-date-and-time-value': { color: theme.palette.text.primary },
    },
    '& input::-webkit-calendar-picker-indicator': {
      filter: theme.palette.mode === 'dark' ? 'invert(1)' : 'none',
      cursor: 'pointer',
    },
  }

  const dateInputProps = {
    style: {
      color: theme.palette.text.primary,
      WebkitTextFillColor: theme.palette.text.primary,
    },
  }

  return (
    <>
      <OutlinedInput
        type="date"
        size="small"
        value={dateFrom}
        onChange={(e) => onDateFromChange(e.target.value)}
        title={to('dateFrom')}
        sx={inputSx}
        inputProps={dateInputProps}
      />
      <OutlinedInput
        type="date"
        size="small"
        value={dateTo}
        onChange={(e) => onDateToChange(e.target.value)}
        title={to('dateTo')}
        sx={inputSx}
        inputProps={dateInputProps}
      />
    </>
  )
}
