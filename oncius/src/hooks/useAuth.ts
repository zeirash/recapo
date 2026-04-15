import { useQuery, useMutation, useQueryClient } from 'react-query'
import { api, getAuthToken, removeAuthToken } from '@/utils/api'
import { User, LoginRequest, RegisterRequest } from '@/types'
import { useRouter } from 'next/navigation'

export const REGISTER_DATA_KEY = 'registerPendingData'

// Module-level flag: once we receive subscription_required, disable the query for the
// lifetime of this tab session. react-query v3 re-fires queryFn on every new observer
// subscription (component mount) when the query is in error state with no cached data —
// refetchOnMount: false only prevents re-fetching stale *successful* data. This flag
// makes enabled: false immediately, so no new observer can trigger a fetch.
let _subscriptionRequired = false

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
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      enabled: !!getAuthToken() && !_subscriptionRequired,
      onSuccess: (data: User) => {
        if (data && !data.subscription_active && data.role !== 'system') {
          _subscriptionRequired = true
          if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/subscription')) {
            router.replace('/subscription')
          }
        }
      },
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
      onSuccess: async () => {
        _subscriptionRequired = false
        // Fetch user data after successful login
        try {
          const userResponse = await api.getCurrentUser()
          if (userResponse.success && userResponse.data) {
            queryClient.setQueryData('currentUser', userResponse.data)
            if (!userResponse.data.subscription_active && userResponse.data.role !== 'system') {
              _subscriptionRequired = true
              router.push('/subscription')
              return
            }
          }
        } catch (error: any) {
          console.error('Failed to fetch user data:', error)
        }
        router.push('/dashboard')
      },
      onError: (error) => {
        console.error('Login error:', error)
      },
    }
  )

  // Send OTP mutation (step 1 of registration)
  const sendOtpMutation = useMutation(
    async (userData: RegisterRequest) => {
      const response = await api.sendOtp(userData.email)
      if (!response.success) {
        throw new Error(response.message || 'Failed to send verification code')
      }
      return userData
    },
    {
      onSuccess: (userData) => {
        if (typeof window !== 'undefined') {
          sessionStorage.setItem(REGISTER_DATA_KEY, JSON.stringify({
            name: userData.name,
            email: userData.email,
            password: userData.password,
          }))
        }
        router.push(`/confirm-email?email=${encodeURIComponent(userData.email)}`)
      },
      onError: (error) => {
        console.error('Send OTP error:', error)
      },
    }
  )

  // Forgot password mutation (step 1 of reset flow)
  const forgotPasswordMutation = useMutation(
    async (email: string) => {
      const response = await api.forgotPassword(email)
      if (!response.success) {
        throw new Error(response.message || 'Failed to send reset code')
      }
    },
    {
      onSuccess: (_, email) => {
        router.push(`/reset-password?email=${encodeURIComponent(email)}`)
      },
      onError: (error) => {
        console.error('Forgot password error:', error)
      },
    }
  )

  // Reset password mutation (step 2 — called from reset-password page with OTP + new password)
  const resetPasswordMutation = useMutation(
    async ({ email, otp, password }: { email: string; otp: string; password: string }) => {
      const response = await api.resetPassword(email, otp, password)
      if (!response.success) {
        throw new Error(response.message || 'Failed to reset password')
      }
    },
    {
      onSuccess: () => {
        router.push('/login?reset=1')
      },
      onError: (error) => {
        console.error('Reset password error:', error)
      },
    }
  )

  // Register mutation (step 2 — called from confirm-email page with OTP)
  const registerMutation = useMutation(
    async ({ otp }: { otp: string }) => {
      const stored = typeof window !== 'undefined'
        ? sessionStorage.getItem(REGISTER_DATA_KEY)
        : null

      if (!stored) {
        throw new Error('Registration data not found. Please start over.')
      }

      const { name, email, password } = JSON.parse(stored) as RegisterRequest

      const response = await api.register(name, email, password, otp)
      if (!response.success) {
        throw new Error(response.message || 'Registration failed')
      }
      return response.data
    },
    {
      onSuccess: () => {
        if (typeof window !== 'undefined') {
          sessionStorage.removeItem(REGISTER_DATA_KEY)
        }
        router.push('/login?registered=1')
      },
      onError: (error) => {
        console.error('Registration error:', error)
      },
    }
  )

  // Logout function
  const logout = async () => {
    _subscriptionRequired = false
    await api.logout()
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
    isSubscriptionRequired: _subscriptionRequired,
    login: loginMutation.mutate,
    loginLoading: loginMutation.isLoading,
    loginError: loginMutation.error as Error | null,
    sendOtp: sendOtpMutation.mutate,
    sendOtpLoading: sendOtpMutation.isLoading,
    sendOtpError: sendOtpMutation.error as Error | null,
    register: registerMutation.mutate,
    registerLoading: registerMutation.isLoading,
    registerError: registerMutation.error as Error | null,
    forgotPassword: forgotPasswordMutation.mutate,
    forgotPasswordLoading: forgotPasswordMutation.isLoading,
    forgotPasswordError: forgotPasswordMutation.error as Error | null,
    resetPassword: resetPasswordMutation.mutate,
    resetPasswordLoading: resetPasswordMutation.isLoading,
    resetPasswordError: resetPasswordMutation.error as Error | null,
    logout,
  }
}
