'use client'

import { useEffect, useRef, useState } from 'react'
import { useQuery } from 'react-query'
import { Box, OutlinedInput } from '@mui/material'
import { Search, X } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { api } from '@/utils/api'

type Product = {
  id: number
  name: string
  price: number
}

type Props = {
  value: number | null
  onChange: (id: number | null, price?: number) => void
  placeholder: string
  searchPlaceholder: string
  noResultsText?: string
}

function useDebounce(value: string, delay: number) {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])
  return debounced
}

export default function ProductSearchSelect({
  value,
  onChange,
  placeholder,
  searchPlaceholder,
  noResultsText = 'No products found',
}: Props) {
  const t = useTranslations('common')
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedName, setSelectedName] = useState('')
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const debouncedSearch = useDebounce(searchTerm, 300)

  const { data: products, isLoading } = useQuery(
    ['products', debouncedSearch],
    () =>
      api
        .getProducts(debouncedSearch || undefined)
        .then((res) => res.data as Product[]),
    { enabled: isOpen, keepPreviousData: true }
  )

  useEffect(() => {
    if (value === null) {
      setSelectedName('')
      setSearchTerm('')
    }
  }, [value])

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
      onChange(null, undefined)
      setSelectedName('')
    }
  }

  function handleFocus() {
    setIsOpen(true)
  }

  function handleSelect(product: Product) {
    setSelectedName(product.name)
    setSearchTerm('')
    onChange(product.id, product.price)
    setIsOpen(false)
  }

  function handleClear() {
    onChange(null, undefined)
    setSelectedName('')
    setSearchTerm('')
    setIsOpen(false)
  }

  const inputValue = value ? selectedName : searchTerm
  const inputPlaceholder = value ? selectedName : (isOpen ? searchPlaceholder : placeholder)

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
          ) : !products || products.length === 0 ? (
            <Box sx={{ display: 'block', px: '16px', py: '8px', fontSize: '14px', color: 'text.secondary' }}>
              {noResultsText}
            </Box>
          ) : (
            products.map((p) => (
              <Box
                key={p.id}
                onClick={() => handleSelect(p)}
                sx={{
                  px: '16px',
                  py: '8px',
                  cursor: 'pointer',
                  '&:hover': { bgcolor: 'action.hover' },
                  borderBottom: '1px solid',
                  borderColor: 'grey.200',
                  '&:last-child': { borderBottom: 'none' },
                }}
              >
                <Box sx={{ display: 'block', fontWeight: 500, fontSize: '14px' }}>{p.name}</Box>
              </Box>
            ))
          )}
        </Box>
      )}
    </Box>
  )
}
