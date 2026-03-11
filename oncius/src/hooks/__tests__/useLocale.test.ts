import { renderHook, act } from '@testing-library/react'
import { useChangeLocale } from '../useLocale'

// Note: window.location.reload cannot be spied on in jsdom — we verify the
// localStorage side-effect and that the call does not throw.
describe('useChangeLocale', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('writes "id" locale to localStorage', () => {
    const { result } = renderHook(() => useChangeLocale())
    // location.reload will trigger jsdom's "not implemented" — suppress console.error
    const spy = jest.spyOn(console, 'error').mockImplementation(() => {})
    act(() => {
      result.current('id')
    })
    spy.mockRestore()
    expect(localStorage.getItem('locale')).toBe('id')
  })

  it('writes "en" locale to localStorage', () => {
    const { result } = renderHook(() => useChangeLocale())
    const spy = jest.spyOn(console, 'error').mockImplementation(() => {})
    act(() => {
      result.current('en')
    })
    spy.mockRestore()
    expect(localStorage.getItem('locale')).toBe('en')
  })
})
