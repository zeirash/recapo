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

const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}))

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const Wrapper = ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children)
  return Wrapper
}

describe('useAuth', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockPush.mockReset()
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

  it('isAuthenticated is false when getCurrentUser returns success false', async () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})
    ;(getAuthToken as jest.Mock).mockReturnValue('some-token')
    ;(api.getCurrentUser as jest.Mock).mockResolvedValue({ success: false, message: 'Forbidden' })

    const { result } = renderHook(() => useAuth(), { wrapper: makeWrapper() })

    await waitFor(() => {
      expect(result.current.isLoadingUser).toBe(false)
    })

    expect(result.current.isAuthenticated).toBe(false)
    consoleSpy.mockRestore()
  })

  it('login mutation calls router.push on success', async () => {
    ;(getAuthToken as jest.Mock).mockReturnValue(null)
    ;(api.login as jest.Mock).mockResolvedValue({ success: true, data: { token: 'tok' } })
    ;(api.getCurrentUser as jest.Mock).mockResolvedValue({ success: true, data: { id: 1, name: 'User' } })

    const { result } = renderHook(() => useAuth(), { wrapper: makeWrapper() })

    await act(async () => {
      result.current.login({ email: 'a@b.com', password: 'pass' })
    })

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/dashboard')
    })
  })

  it('login mutation sets loginError on failure', async () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})
    ;(getAuthToken as jest.Mock).mockReturnValue(null)
    ;(api.login as jest.Mock).mockResolvedValue({ success: false, message: 'Invalid credentials' })

    const { result } = renderHook(() => useAuth(), { wrapper: makeWrapper() })

    await act(async () => {
      result.current.login({ email: 'a@b.com', password: 'wrong' })
    })

    await waitFor(() => {
      expect(result.current.loginError?.message).toBe('Invalid credentials')
    })
    consoleSpy.mockRestore()
  })

  it('register mutation calls router.push on success', async () => {
    ;(getAuthToken as jest.Mock).mockReturnValue(null)
    ;(api.register as jest.Mock).mockResolvedValue({ success: true, data: {} })

    const { result } = renderHook(() => useAuth(), { wrapper: makeWrapper() })

    await act(async () => {
      result.current.register({ name: 'User', email: 'a@b.com', password: 'pass' })
    })

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/login?message=Registration successful. Please login.')
    })
  })

  it('register mutation sets registerError on failure', async () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})
    ;(getAuthToken as jest.Mock).mockReturnValue(null)
    ;(api.register as jest.Mock).mockResolvedValue({ success: false, message: 'Email taken' })

    const { result } = renderHook(() => useAuth(), { wrapper: makeWrapper() })

    await act(async () => {
      result.current.register({ name: 'User', email: 'taken@b.com', password: 'pass' })
    })

    await waitFor(() => {
      expect(result.current.registerError?.message).toBe('Email taken')
    })
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
