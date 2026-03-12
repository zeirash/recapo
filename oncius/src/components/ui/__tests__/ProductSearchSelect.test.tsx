import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from 'react-query'
import ProductSearchSelect from '../ProductSearchSelect'
import { api } from '@/utils/api'

jest.mock('@/utils/api', () => ({
  api: {
    getProducts: jest.fn(),
  },
}))

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const Wrapper = ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children)
  return Wrapper
}

const defaultProps = {
  value: null,
  onChange: jest.fn(),
  placeholder: 'Select product',
  searchPlaceholder: 'Search products...',
}

describe('ProductSearchSelect', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders with placeholder text', () => {
    render(<ProductSearchSelect {...defaultProps} />, { wrapper: makeWrapper() })
    expect(screen.getByPlaceholderText('Select product')).toBeInTheDocument()
  })

  it('opens dropdown on focus and shows loading then results', async () => {
    const mockProducts = [
      { id: 1, name: 'Product A', price: 10000 },
      { id: 2, name: 'Product B', price: 20000 },
    ]
    ;(api.getProducts as jest.Mock).mockResolvedValue({ success: true, data: mockProducts })

    render(<ProductSearchSelect {...defaultProps} />, { wrapper: makeWrapper() })

    await userEvent.click(screen.getByPlaceholderText('Select product'))

    await waitFor(() => {
      expect(screen.getByText('Product A')).toBeInTheDocument()
      expect(screen.getByText('Product B')).toBeInTheDocument()
    })
  })

  it('calls onChange with product id and price when item selected', async () => {
    const handleChange = jest.fn()
    const mockProducts = [{ id: 5, name: 'Widget', price: 5000 }]
    ;(api.getProducts as jest.Mock).mockResolvedValue({ success: true, data: mockProducts })

    render(
      <ProductSearchSelect {...defaultProps} onChange={handleChange} />,
      { wrapper: makeWrapper() }
    )

    await userEvent.click(screen.getByPlaceholderText('Select product'))

    await waitFor(() => {
      expect(screen.getByText('Widget')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('Widget'))

    expect(handleChange).toHaveBeenCalledWith(5, 5000)
  })

  it('shows no results text when products list is empty', async () => {
    ;(api.getProducts as jest.Mock).mockResolvedValue({ success: true, data: [] })

    render(
      <ProductSearchSelect {...defaultProps} noResultsText="Nothing found" />,
      { wrapper: makeWrapper() }
    )

    await userEvent.click(screen.getByPlaceholderText('Select product'))

    await waitFor(() => {
      expect(screen.getByText('Nothing found')).toBeInTheDocument()
    })
  })

  it('clears selection when clear button is clicked', async () => {
    const handleChange = jest.fn()
    ;(api.getProducts as jest.Mock).mockResolvedValue({ success: true, data: [] })

    render(
      <ProductSearchSelect
        {...defaultProps}
        value={5}
        onChange={handleChange}
      />,
      { wrapper: makeWrapper() }
    )

    const clearBtn = screen.getByRole('button')
    await userEvent.click(clearBtn)

    expect(handleChange).toHaveBeenCalledWith(null, undefined)
  })

  it('calls onChange(null) when typing while a value is selected', async () => {
    const handleChange = jest.fn()
    ;(api.getProducts as jest.Mock).mockResolvedValue({ success: true, data: [] })

    render(
      <ProductSearchSelect
        {...defaultProps}
        value={5}
        onChange={handleChange}
      />,
      { wrapper: makeWrapper() }
    )

    const input = screen.getByRole('textbox')
    await userEvent.type(input, 'a')

    expect(handleChange).toHaveBeenCalledWith(null, undefined)
  })

  it('closes dropdown when clicking outside', async () => {
    ;(api.getProducts as jest.Mock).mockResolvedValue({ success: true, data: [] })

    render(
      <div>
        <ProductSearchSelect {...defaultProps} noResultsText="Nothing found" />
        <div data-testid="outside">outside</div>
      </div>,
      { wrapper: makeWrapper() }
    )

    await userEvent.click(screen.getByPlaceholderText('Select product'))
    await waitFor(() => expect(screen.getByText('Nothing found')).toBeInTheDocument())

    await userEvent.click(screen.getByTestId('outside'))
    expect(screen.queryByText('Nothing found')).not.toBeInTheDocument()
  })
})
