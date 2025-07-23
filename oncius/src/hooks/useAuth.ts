import { useQuery, useMutation, useQueryClient } from 'react-query'
import { api, getAuthToken, removeAuthToken } from '@/utils/api'
import { User, LoginRequest, RegisterRequest } from '@/types'
import { useRouter } from 'next/navigation'

export const useAuth = () => {
  const queryClient = useQueryClient()
  const router = useRouter()

  // Get current user
  const {
    data: user,
    isLoading: isLoadingUser,
    error: userError,
  } = useQuery<User>(
    'currentUser',
    async () => {
      const token = getAuthToken()
      if (!token) {
        throw new Error('No auth token')
      }
      const response = await api.getCurrentUser()
      if (!response.success) {
        throw new Error(response.message || 'Failed to get user')
      }
      return response.data
    },
    {
      retry: false,
      enabled: !!getAuthToken(),
    }
  )

  // Login mutation
  const loginMutation = useMutation(
    async (credentials: LoginRequest) => {
      const response = await api.login(credentials.email, credentials.password)
      if (!response.success) {
        throw new Error(response.message || 'Login failed')
      }
      return response.data
    },
    {
      onSuccess: async (data) => {
        // Fetch user data after successful login
        try {
          const userResponse = await api.getCurrentUser()
          if (userResponse.success && userResponse.data) {
            queryClient.setQueryData('currentUser', userResponse.data)
          }
        } catch (error) {
          console.error('Failed to fetch user data:', error)
        }
        router.push('/dashboard')
      },
      onError: (error) => {
        console.error('Login error:', error)
      },
    }
  )

  // Register mutation
  const registerMutation = useMutation(
    async (userData: RegisterRequest) => {
      const response = await api.register(userData.name, userData.email, userData.password)
      if (!response.success) {
        throw new Error(response.message || 'Registration failed')
      }
      return response.data
    },
    {
      onSuccess: () => {
        router.push('/login?message=Registration successful. Please login.')
      },
      onError: (error) => {
        console.error('Registration error:', error)
      },
    }
  )

  // Logout function
  const logout = () => {
    api.logout()
    queryClient.clear()
    router.push('/login')
  }

  // Check if user is authenticated
  const isAuthenticated = !!user && !userError

  return {
    user,
    isLoadingUser,
    userError,
    isAuthenticated,
    login: loginMutation.mutate,
    loginLoading: loginMutation.isLoading,
    loginError: loginMutation.error as Error | null,
    register: registerMutation.mutate,
    registerLoading: registerMutation.isLoading,
    registerError: registerMutation.error as Error | null,
    logout,
  }
}
