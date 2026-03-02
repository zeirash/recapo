"use client"

import { Box, OutlinedInput } from '@mui/material'
import { Search } from 'lucide-react'

type SearchInputProps = {
  value: string
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  placeholder: string
  sx?: object
}

export default function SearchInput({ value, onChange, placeholder, sx = {} }: SearchInputProps) {
  return (
    <Box sx={{ position: 'relative', flex: 1, ...sx }}>
      <Box
        sx={{
          position: 'absolute',
          left: '16px',
          top: '50%',
          transform: 'translateY(-50%)',
          color: '#6b7280',
          pointerEvents: 'none',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1,
        }}
      >
        <Search size={18} />
      </Box>
      <OutlinedInput
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        inputProps={{ 'aria-label': placeholder }}
        size="small"
        sx={{
          width: '100%',
          borderRadius: '8px',
          bgcolor: '#f9fafb',
          '& .MuiOutlinedInput-notchedOutline': {
            borderColor: '#e5e7eb',
          },
          '& .MuiOutlinedInput-input': {
            paddingLeft: '40px',
          },
        }}
      />
    </Box>
  )
}
