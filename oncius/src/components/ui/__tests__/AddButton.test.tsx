import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AddButton from '../AddButton'

describe('AddButton', () => {
  it('renders with the correct title attribute', () => {
    render(<AddButton onClick={jest.fn()} title="Add Order" />)
    expect(screen.getByTitle('Add Order')).toBeInTheDocument()
  })

  it('calls onClick when clicked', async () => {
    const handleClick = jest.fn()
    render(<AddButton onClick={handleClick} title="Add" />)
    await userEvent.click(screen.getByTitle('Add'))
    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('renders a Plus icon (SVG)', () => {
    const { container } = render(<AddButton onClick={jest.fn()} title="Add" />)
    expect(container.querySelector('svg')).toBeInTheDocument()
  })
})
