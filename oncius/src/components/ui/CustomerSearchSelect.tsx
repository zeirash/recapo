'use client'

import { useEffect, useRef, useState } from 'react'
import { useQuery } from 'react-query'
import { Box, OutlinedInput } from '@mui/material'
import { Search, UserPlus, X } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { api } from '@/utils/api'

type Customer = {
  id: number
  name: string
  phone: string
  address: string
}

type Props = {
  value: number | null
  onChange: (id: number | null) => void
  placeholder: string
  searchPlaceholder: string
  noResultsText?: string
  onCreateCustomer?: (searchTerm: string) => void
  selectedLabel?: string
}

function useDebounce(value: string, delay: number) {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])
  return debounced
}

export default function CustomerSearchSelect({
  value,
  onChange,
  placeholder,
  searchPlaceholder,
  noResultsText = 'No customers found',
  onCreateCustomer,
  selectedLabel,
}: Props) {
  const t = useTranslations('common')
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedName, setSelectedName] = useState('')
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const debouncedSearch = useDebounce(searchTerm, 300)

  const { data: customers, isLoading } = useQuery(
    ['customers', debouncedSearch],
    () =>
      api
        .getCustomers(debouncedSearch || undefined)
        .then((res) => res.data as Customer[]),
    { enabled: isOpen, keepPreviousData: true }
  )

  // Clear display when value is reset externally
  useEffect(() => {
    if (value === null) {
      setSelectedName('')
      setSearchTerm('')
    }
  }, [value])

  // Click-outside to close
  useEffect(() => {
    function handleMouseDown(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleMouseDown)
    return () => document.removeEventListener('mousedown', handleMouseDown)
  }, [])

  function handleInputChange(e: React.ChangeEvent<HTMLInputElement>) {
    setSearchTerm(e.target.value)
    if (value !== null) {
      onChange(null)
      setSelectedName('')
    }
  }

  function handleFocus() {
    setIsOpen(true)
  }

  function handleSelect(customer: Customer) {
    setSelectedName(customer.name)
    setSearchTerm('')
    onChange(customer.id)
    setIsOpen(false)
  }

  function handleClear() {
    onChange(null)
    setSelectedName('')
    setSearchTerm('')
    setIsOpen(false)
  }

  const inputValue = value ? (selectedName || selectedLabel || '') : searchTerm
  const inputPlaceholder = value ? (selectedName || selectedLabel || '') : (isOpen ? searchPlaceholder : placeholder)

  return (
    <Box ref={containerRef} sx={{ position: 'relative' }}>
      <Box sx={{ position: 'relative' }}>
        <Box
          sx={{
            position: 'absolute',
            left: '8px',
            top: '50%',
            transform: 'translateY(-50%)',
            pointerEvents: 'none',
            color: 'text.secondary',
            display: 'flex',
            alignItems: 'center',
          }}
        >
          <Search size={14} />
        </Box>
        <OutlinedInput
          size="small"
          value={inputValue}
          onChange={handleInputChange}
          onFocus={handleFocus}
          placeholder={inputPlaceholder}
          sx={{
            width: '100%',
            pr: value ? '30px' : '8px',
            '& .MuiOutlinedInput-input': { paddingLeft: '30px' },
          }}
        />
        {value && (
          <button
            type="button"
            onClick={handleClear}
            style={{
              position: 'absolute',
              right: 8,
              top: '50%',
              transform: 'translateY(-50%)',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              padding: 0,
              display: 'flex',
              alignItems: 'center',
              color: 'inherit',
              opacity: 0.5,
            }}
          >
            <X size={14} />
          </button>
        )}
      </Box>

      {isOpen && (
        <Box
          sx={{
            position: 'absolute',
            top: '100%',
            left: 0,
            right: 0,
            zIndex: 100,
            mt: '4px',
            bgcolor: 'background.paper',
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: '8px',
            boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)',
            maxHeight: '200px',
            overflowY: 'auto',
          }}
        >
          {isLoading ? (
            <Box sx={{ display: 'block', px: '16px', py: '8px', fontSize: '14px', color: 'text.secondary' }}>
              {t('loading')}
            </Box>
          ) : !customers || customers.length === 0 ? (
            <Box>
              <Box sx={{ display: 'block', px: '16px', py: '8px', fontSize: '14px', color: 'text.secondary' }}>
                {noResultsText}
              </Box>
              {onCreateCustomer && debouncedSearch && (
                <Box
                  onClick={() => {
                    onCreateCustomer(debouncedSearch)
                    setIsOpen(false)
                  }}
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                    px: '16px',
                    py: '8px',
                    cursor: 'pointer',
                    color: 'primary.main',
                    fontSize: '14px',
                    fontWeight: 500,
                    borderTop: '1px solid',
                    borderColor: 'divider',
                    '&:hover': { bgcolor: 'action.hover' },
                  }}
                >
                  <UserPlus size={14} />
                  {t('createItem', { name: debouncedSearch })}
                </Box>
              )}
            </Box>
          ) : (
            customers.map((c) => (
              <Box
                key={c.id}
                onClick={() => handleSelect(c)}
                sx={{
                  px: '16px',
                  py: '8px',
                  cursor: 'pointer',
                  '&:hover': { bgcolor: 'action.hover' },
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                  '&:last-child': { borderBottom: 'none' },
                }}
              >
                <Box sx={{ display: 'block', fontWeight: 500, fontSize: '14px' }}>{c.name}</Box>
                <Box sx={{ display: 'block', fontSize: '12px', color: 'text.secondary' }}>{c.phone}</Box>
              </Box>
            ))
          )}
        </Box>
      )}
    </Box>
  )
}
