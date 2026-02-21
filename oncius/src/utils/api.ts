import { ApiResponse } from '@/types'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:4000'

interface LoginResponse {
  access_token: string
  refresh_token: string
}

class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: any
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

// Get auth token from localStorage
const getAuthToken = (): string | null => {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('authToken')
  }
  return null
}

// Set auth token to localStorage
const setAuthToken = (token: string): void => {
  if (typeof window !== 'undefined') {
    localStorage.setItem('authToken', token)
  }
}

// Remove auth token from localStorage
const removeAuthToken = (): void => {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('authToken')
  }
}

// Get refresh token from localStorage
const getRefreshToken = (): string | null => {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('refreshToken')
  }
  return null
}

// Set refresh token to localStorage
const setRefreshToken = (token: string): void => {
  if (typeof window !== 'undefined') {
    localStorage.setItem('refreshToken', token)
  }
}

// Remove refresh token from localStorage
const removeRefreshToken = (): void => {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('refreshToken')
  }
}

// Remove all tokens and redirect to login
const clearTokensAndRedirect = (): void => {
  removeAuthToken()
  removeRefreshToken()
  if (typeof window !== 'undefined') {
    window.location.href = '/login'
  }
}

// Flag to prevent multiple refresh attempts
let isRefreshing = false
let refreshPromise: Promise<boolean> | null = null

// Attempt to refresh tokens
const refreshTokens = async (): Promise<boolean> => {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    return false
  }

  try {
    const response = await fetch(`${API_BASE_URL}/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!response.ok) {
      return false
    }

    const data = await response.json() as ApiResponse<LoginResponse>
    if (data.success && data.data?.access_token && data.data?.refresh_token) {
      setAuthToken(data.data.access_token)
      setRefreshToken(data.data.refresh_token)
      return true
    }

    return false
  } catch {
    return false
  }
}

// Get locale for Accept-Language header (matches frontend i18n)
const getApiLocale = (): string => {
  if (typeof window !== 'undefined') {
    const locale = localStorage.getItem('locale')
    if (locale === 'id' || locale === 'en') return locale
  }
  return 'en'
}

// Base API request function
const apiRequest = async <T>(
  endpoint: string,
  options: RequestInit = {},
  skipAuth: boolean = false
): Promise<T> => {
  const token = skipAuth ? null : getAuthToken()

  const config: RequestInit = {
    headers: {
      'Content-Type': 'application/json',
      'Accept-Language': getApiLocale(),
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    },
    ...options,
  }

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, config)

    if (!response.ok) {
      // Handle 401 Unauthorized - attempt token refresh
      if (response.status === 401 && !skipAuth) {
        // If already refreshing, wait for that to complete
        if (isRefreshing && refreshPromise) {
          const refreshed = await refreshPromise
          if (refreshed) {
            // Retry the original request with new token
            return apiRequest<T>(endpoint, options, skipAuth)
          }
        } else {
          // Start refreshing
          isRefreshing = true
          refreshPromise = refreshTokens()
          const refreshed = await refreshPromise
          isRefreshing = false
          refreshPromise = null

          if (refreshed) {
            // Retry the original request with new token
            return apiRequest<T>(endpoint, options, skipAuth)
          }
        }

        // Refresh failed, clear tokens and redirect
        clearTokensAndRedirect()
        throw new ApiError('Unauthorized', 401)
      }

      const errorData = await response.json().catch(() => ({}))
      throw new ApiError(
        errorData.message || `HTTP error! status: ${response.status}`,
        response.status,
        errorData
      )
    }

    const data = await response.json()
    return data
  } catch (error) {
    if (error instanceof ApiError) {
      throw error
    }
    throw new ApiError(
      error instanceof Error ? error.message : 'Network error',
      0
    )
  }
}

// API methods
export const api = {
  // Authentication
  login: async (email: string, password: string) => {
    const response = await apiRequest<ApiResponse<LoginResponse>>('/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }, true)

    if (response.success && response.data?.access_token) {
      setAuthToken(response.data.access_token)
      if (response.data.refresh_token) {
        setRefreshToken(response.data.refresh_token)
      }
    }

    return response
  },

  register: async (name: string, email: string, password: string) => {
    return apiRequest<ApiResponse>('/register', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    }, true)
  },

  logout: () => {
    removeAuthToken()
    removeRefreshToken()
  },

  // User
  getCurrentUser: async () => {
    return apiRequest<ApiResponse<any>>('/user')
  },

  // Shop
  getShopShareToken: async () => {
    return apiRequest<ApiResponse<{ share_token: string }>>('/shop/share_token')
  },

  updateUser: async (data: any) => {
    return apiRequest<ApiResponse>('/user', {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  },

  // Products
  getProducts: (search?: string) => {
    const params = search ? `?search=${encodeURIComponent(search)}` : ''
    return apiRequest<ApiResponse<any[]>>(`/products${params}`)
  },

  getProduct: (id: number | string) => {
    return apiRequest<ApiResponse<any>>(`/products/${id}`)
  },

  createProduct: (data: { name: string; description?: string; price: number; original_price?: number }) => {
    return apiRequest<ApiResponse>('/product', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  updateProduct: (id: number | string,
    data: Partial<{ name: string; description: string; price: number; original_price: number }>
  ) => {
    return apiRequest<ApiResponse>(`/products/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  },

  deleteProduct: (id: number | string) => {
    return apiRequest<ApiResponse>(`/products/${id}`, {
      method: 'DELETE',
    })
  },

  // Customers
  getCustomers: (search?: string) => {
    const params = search ? `?search=${encodeURIComponent(search)}` : ''
    return apiRequest<ApiResponse<any[]>>(`/customers${params}`)
  },

  getCustomer: (id: number | string) => {
    return apiRequest<ApiResponse<any>>(`/customers/${id}`)
  },

  createCustomer: (data: { name: string; phone: string; address: string }) => {
    return apiRequest<ApiResponse>('/customer', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  updateCustomer: (
    id: number | string,
    data: Partial<{ name: string; phone: string; address: string }>
  ) => {
    return apiRequest<ApiResponse>(`/customers/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  },

  deleteCustomer: (id: number | string) => {
    return apiRequest<ApiResponse>(`/customers/${id}`, {
      method: 'DELETE',
    })
  },

  checkActiveOrder: (data: { phone: string; name: string }) => {
    return apiRequest<ApiResponse<{ customer_id: number; active_order_id: number }>>(
      '/customers/check_active_order',
      {
        method: 'POST',
        body: JSON.stringify(data),
      }
    )
  },

  // Orders
  getOrders: (opts?: { search?: string; date_from?: string; date_to?: string }) => {
    const params = new URLSearchParams()
    if (opts?.search) params.set('search', opts.search)
    if (opts?.date_from) params.set('date_from', opts.date_from)
    if (opts?.date_to) params.set('date_to', opts.date_to)
    const qs = params.toString()
    return apiRequest<ApiResponse<any[]>>(`/orders${qs ? `?${qs}` : ''}`)
  },

  getOrder: (id: number | string) => {
    return apiRequest<ApiResponse<any>>(`/orders/${id}`)
  },

  createOrder: (data: { customer_id: number; notes?: string }) => {
    return apiRequest<ApiResponse>('/order', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  updateOrder: (
    id: number | string,
    data: Partial<{ customer_id: number; total_price: number; status: string; notes: string }>
  ) => {
    return apiRequest<ApiResponse>(`/orders/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  },

  deleteOrder: (id: number | string) => {
    return apiRequest<ApiResponse>(`/orders/${id}`, {
      method: 'DELETE',
    })
  },

  mergeTempOrder: (data: {
    temp_order_id: number
    customer_id: number
    active_order_id?: number
  }) => {
    return apiRequest<ApiResponse<any>>('/temp_orders/merge', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  // Order Items
  getOrderItems: (orderId: number | string) => {
    return apiRequest<ApiResponse<any[]>>(`/orders/${orderId}/items`)
  },

  createOrderItem: (orderId: number | string, data: { product_id: number; qty: number }) => {
    return apiRequest<ApiResponse>(`/orders/${orderId}/item`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  updateOrderItem: (
    orderId: number | string,
    itemId: number | string,
    data: Partial<{ product_id: number; qty: number }>
  ) => {
    return apiRequest<ApiResponse>(`/orders/${orderId}/items/${itemId}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  },

  deleteOrderItem: (orderId: number | string, itemId: number | string) => {
    return apiRequest<ApiResponse>(`/orders/${orderId}/items/${itemId}`, {
      method: 'DELETE',
    })
  },

  // Temp orders (from public share page)
  getTempOrders: (opts?: { search?: string; status?: string; date_from?: string; date_to?: string }) => {
    const params = new URLSearchParams()
    if (opts?.search) params.set('search', opts.search)
    if (opts?.status) params.set('status', opts.status)
    if (opts?.date_from) params.set('date_from', opts.date_from)
    if (opts?.date_to) params.set('date_to', opts.date_to)
    const qs = params.toString()
    return apiRequest<ApiResponse<any[]>>(`/temp_orders${qs ? `?${qs}` : ''}`)
  },

  getTempOrder: (id: number | string) => {
    return apiRequest<ApiResponse<any>>(`/temp_orders/${id}`)
  },

  rejectTempOrder: (tempOrderId: number | string) => {
    return apiRequest<ApiResponse<any>>(`/temp_orders/${tempOrderId}/reject`, {
      method: 'PATCH',
    })
  },

  // Purchase list
  getPurchaseListProducts: () => {
    return apiRequest<ApiResponse<any[]>>('/products/purchase_list')
  },

  // Health check
  health: () => {
    return apiRequest<ApiResponse>('/health')
  },

  // Public (no auth)
  getPublicProducts: (shareToken: string) => {
    return apiRequest<ApiResponse<any[]>>(`/public/shops/${encodeURIComponent(shareToken)}/products`, {}, true)
  },

  createPublicOrderTemp: (
    shareToken: string,
    data: { customer_name: string; customer_phone: string; order_items: Array<{ product_id: number; qty: number }> }
  ) => {
    return apiRequest<ApiResponse<any>>(
      `/public/shops/${encodeURIComponent(shareToken)}/order`,
      { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) },
      true
    )
  },
}

export { ApiError, getAuthToken, setAuthToken, removeAuthToken, getRefreshToken, setRefreshToken, removeRefreshToken }
