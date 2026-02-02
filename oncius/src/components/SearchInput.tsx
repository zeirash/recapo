"use client"

import { Box, Input } from 'theme-ui'

type SearchInputProps = {
  value: string
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  placeholder: string
  sx?: object
}

const SearchIcon = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="11" cy="11" r="8" />
    <path d="m21 21-4.35-4.35" />
  </svg>
)

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
        <SearchIcon />
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
