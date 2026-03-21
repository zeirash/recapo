import { Box } from '@mui/material'
import { useTranslations } from 'next-intl'

type Props = {
  dateFrom: string
  dateTo: string
  onDateFromChange: (value: string) => void
  onDateToChange: (value: string) => void
}

export default function DateRangeFilter({ dateFrom, dateTo, onDateFromChange, onDateToChange }: Props) {
  const to = useTranslations('orders')

  return (
    <Box sx={{ mt: '8px', display: 'flex', gap: '8px', alignItems: 'center' }}>
      <input
        type="date"
        value={dateFrom}
        onChange={(e) => onDateFromChange(e.target.value)}
        title={to('dateFrom')}
        style={{
          flex: 1,
          padding: '6px',
          fontSize: 12,
          borderRadius: 6,
          border: '1px solid var(--theme-ui-colors-border, #e0e0e0)',
          backgroundColor: 'white',
          color: dateFrom ? 'var(--theme-ui-colors-text, #333)' : '#aaa',
        }}
      />
      <input
        type="date"
        value={dateTo}
        onChange={(e) => onDateToChange(e.target.value)}
        title={to('dateTo')}
        style={{
          flex: 1,
          padding: '6px',
          fontSize: 12,
          borderRadius: 6,
          border: '1px solid var(--theme-ui-colors-border, #e0e0e0)',
          backgroundColor: 'white',
          color: dateTo ? 'var(--theme-ui-colors-text, #333)' : '#aaa',
        }}
      />
    </Box>
  )
}
