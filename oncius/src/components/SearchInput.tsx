"use client"

import { Box, Input } from 'theme-ui'
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
          left: 3,
          top: '50%',
          transform: 'translateY(-50%)',
          color: 'text.secondary',
          pointerEvents: 'none',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Search size={18} />
      </Box>
      <Input
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        aria-label={placeholder}
        sx={{
          pl: 5,
          width: '100%',
          borderRadius: 'medium',
          bg: 'background.secondary',
          border: '1px solid',
          borderColor: 'border',
        }}
      />
    </Box>
  )
}
