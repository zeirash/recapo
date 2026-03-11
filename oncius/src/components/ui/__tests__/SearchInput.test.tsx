import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import SearchInput from '../SearchInput'

describe('SearchInput', () => {
  it('renders with placeholder text', () => {
    render(<SearchInput value="" onChange={jest.fn()} placeholder="Search orders..." />)
    expect(screen.getByPlaceholderText('Search orders...')).toBeInTheDocument()
  })

  it('calls onChange when typing', async () => {
    const handleChange = jest.fn()
    render(<SearchInput value="" onChange={handleChange} placeholder="Search" />)
    await userEvent.type(screen.getByPlaceholderText('Search'), 'hello')
    expect(handleChange).toHaveBeenCalled()
  })

  it('renders the Search icon', () => {
    const { container } = render(<SearchInput value="" onChange={jest.fn()} placeholder="Search" />)
    // lucide-react renders SVG icons
    expect(container.querySelector('svg')).toBeInTheDocument()
  })

  it('displays the current value', () => {
    render(<SearchInput value="test query" onChange={jest.fn()} placeholder="Search" />)
    expect(screen.getByDisplayValue('test query')).toBeInTheDocument()
  })
})
