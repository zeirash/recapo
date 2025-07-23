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
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    },
    ...options,
  }

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, config)

    if (!response.ok) {
      // Handle 401 Unauthorized
      if (response.status === 401) {
        removeAuthToken()
        window.location.href = '/login'
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
  },

  // User
  getCurrentUser: async () => {
    return apiRequest<ApiResponse<any>>('/user')
  },

  updateUser: async (data: any) => {
    return apiRequest<ApiResponse>('/user', {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  },

  // Products
  // getProducts: (params?: { page?: number; limit?: number; search?: string }) => {
  //   const searchParams = new URLSearchParams()
  //   if (params?.page) searchParams.append('page', params.page.toString())
  //   if (params?.limit) searchParams.append('limit', params.limit.toString())
  //   if (params?.search) searchParams.append('search', params.search)

  //   const query = searchParams.toString() ? `?${searchParams.toString()}` : ''
  //   return apiRequest<ApiResponse<any[]>>(`/products${query}`)
  // },

  // getProduct: (id: string) => {
  //   return apiRequest<ApiResponse<any>>(`/products/${id}`)
  // },

  // createProduct: (data: any) => {
  //   return apiRequest<ApiResponse>('/products', {
  //     method: 'POST',
  //     body: JSON.stringify(data),
  //   })
  // },

  // updateProduct: (id: string, data: any) => {
  //   return apiRequest<ApiResponse>(`/products/${id}`, {
  //     method: 'PUT',
  //     body: JSON.stringify(data),
  //   })
  // },

  // deleteProduct: (id: string) => {
  //   return apiRequest<ApiResponse>(`/products/${id}`, {
  //     method: 'DELETE',
  //   })
  // },

  // Customers
  // getCustomers: (params?: { page?: number; limit?: number; search?: string }) => {
  //   const searchParams = new URLSearchParams()
  //   if (params?.page) searchParams.append('page', params.page.toString())
  //   if (params?.limit) searchParams.append('limit', params.limit.toString())
  //   if (params?.search) searchParams.append('search', params.search)

  //   const query = searchParams.toString() ? `?${searchParams.toString()}` : ''
  //   return apiRequest<ApiResponse<any[]>>(`/customers${query}`)
  // },

  // getCustomer: (id: string) => {
  //   return apiRequest<ApiResponse<any>>(`/customers/${id}`)
  // },

  // createCustomer: (data: any) => {
  //   return apiRequest<ApiResponse>('/customers', {
  //     method: 'POST',
  //     body: JSON.stringify(data),
  //   })
  // },

  // updateCustomer: (id: string, data: any) => {
  //   return apiRequest<ApiResponse>(`/customers/${id}`, {
  //     method: 'PUT',
  //     body: JSON.stringify(data),
  //   })
  // },

  // deleteCustomer: (id: string) => {
  //   return apiRequest<ApiResponse>(`/customers/${id}`, {
  //     method: 'DELETE',
  //   })
  // },

  // Orders
  // getOrders: (params?: { page?: number; limit?: number; status?: string; dateFrom?: string; dateTo?: string }) => {
  //   const searchParams = new URLSearchParams()
  //   if (params?.page) searchParams.append('page', params.page.toString())
  //   if (params?.limit) searchParams.append('limit', params.limit.toString())
  //   if (params?.status) searchParams.append('status', params.status)
  //   if (params?.dateFrom) searchParams.append('dateFrom', params.dateFrom)
  //   if (params?.dateTo) searchParams.append('dateTo', params.dateTo)

  //   const query = searchParams.toString() ? `?${searchParams.toString()}` : ''
  //   return apiRequest<ApiResponse<any[]>>(`/orders${query}`)
  // },

  // getOrder: (id: string) => {
  //   return apiRequest<ApiResponse<any>>(`/orders/${id}`)
  // },

  // createOrder: (data: any) => {
  //   return apiRequest<ApiResponse>('/orders', {
  //     method: 'POST',
  //     body: JSON.stringify(data),
  //   })
  // },

  // updateOrder: (id: string, data: any) => {
  //   return apiRequest<ApiResponse>(`/orders/${id}`, {
  //     method: 'PUT',
  //     body: JSON.stringify(data),
  //   })
  // },

  // deleteOrder: (id: string) => {
  //   return apiRequest<ApiResponse>(`/orders/${id}`, {
  //     method: 'DELETE',
  //   })
  // },

  // Health check
  health: () => {
    return apiRequest<ApiResponse>('/health')
  },
}

export { ApiError, getAuthToken, setAuthToken, removeAuthToken }
