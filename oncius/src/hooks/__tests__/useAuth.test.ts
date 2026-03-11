import { renderHook, act, waitFor } from '@testing-library/react'
import React from 'react'
import { QueryClient, QueryClientProvider } from 'react-query'
import { useAuth } from '../useAuth'
import { api, getAuthToken } from '@/utils/api'

jest.mock('@/utils/api', () => ({
  api: {
    getCurrentUser: jest.fn(),
    login: jest.fn(),
    logout: jest.fn(),
    register: jest.fn(),
  },
  getAuthToken: jest.fn(),
  removeAuthToken: jest.fn(),
}))

jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: jest.fn() }),
}))

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('useAuth', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('isAuthenticated is false when no token', async () => {
    ;(getAuthToken as jest.Mock).mockReturnValue(null)

    const { result } = renderHook(() => useAuth(), { wrapper: makeWrapper() })

    await waitFor(() => {
      expect(result.current.isLoadingUser).toBe(false)
    })

    expect(result.current.isAuthenticated).toBe(false)
  })

  it('isAuthenticated is true when getCurrentUser succeeds', async () => {
    const mockUser = { id: 1, name: 'Test User', email: 'test@example.com' }
    ;(getAuthToken as jest.Mock).mockReturnValue('mock-token')
    ;(api.getCurrentUser as jest.Mock).mockResolvedValue({ success: true, data: mockUser })

    const { result } = renderHook(() => useAuth(), { wrapper: makeWrapper() })

    await waitFor(() => {
      expect(result.current.isAuthenticated).toBe(true)
    })

    expect(result.current.user).toEqual(mockUser)
  })

  it('isAuthenticated is false when getCurrentUser fails', async () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})
    ;(getAuthToken as jest.Mock).mockReturnValue('bad-token')
    ;(api.getCurrentUser as jest.Mock).mockRejectedValue(new Error('Unauthorized'))

    const { result } = renderHook(() => useAuth(), { wrapper: makeWrapper() })

    await waitFor(() => {
      expect(result.current.isLoadingUser).toBe(false)
    })

    expect(result.current.isAuthenticated).toBe(false)
    consoleSpy.mockRestore()
  })

  it('logout calls api.logout', () => {
    ;(getAuthToken as jest.Mock).mockReturnValue(null)

    const { result } = renderHook(() => useAuth(), { wrapper: makeWrapper() })

    act(() => {
      result.current.logout()
    })

    expect(api.logout).toHaveBeenCalledTimes(1)
  })
})
