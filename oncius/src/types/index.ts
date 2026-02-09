import React from 'react'
import { UserRole } from '@/constants/roles'

// API Response Types
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  code?: string
  message?: string
}

// Authentication Types
export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
}

export interface AuthResponse {
  accessToken: string
  refreshToken: string
  user: User
}

export interface User {
  id: string
  name: string
  email: string
  role: UserRole
  createdAt: string
  updatedAt: string
}

// Product Types
export interface Product {
  id: string
  name: string
  description?: string
  price: number
  currency: string
  stock: number
  category?: string
  imageUrl?: string
  userId: string
  createdAt: string
  updatedAt: string
}

export interface CreateProductRequest {
  name: string
  description?: string
  price: number
  currency: string
  stock: number
  category?: string
  imageUrl?: string
}

export interface UpdateProductRequest extends Partial<CreateProductRequest> {
  id: string
}

// Customer Types
export interface Customer {
  id: string
  name: string
  email?: string
  phone?: string
  address?: string
  notes?: string
  userId: string
  createdAt: string
  updatedAt: string
}

export interface CreateCustomerRequest {
  name: string
  email?: string
  phone?: string
  address?: string
  notes?: string
}

export interface UpdateCustomerRequest extends Partial<CreateCustomerRequest> {
  id: string
}

// Order Types
export interface Order {
  id: string
  orderNumber: string
  customerId: string
  customer: Customer
  items: OrderItem[]
  totalAmount: number
  currency: string
  status: OrderStatus
  notes?: string
  userId: string
  createdAt: string
  updatedAt: string
}

export interface OrderItem {
  id: string
  productId: string
  product: Product
  quantity: number
  price: number
  currency: string
}

export type OrderStatus = 'created' | 'in_progress' | 'in_delivery' | 'done' | 'cancelled'

export interface CreateOrderRequest {
  customerId: string
  items: {
    productId: string
    quantity: number
  }[]
  notes?: string
}

export interface UpdateOrderRequest {
  id: string
  status?: OrderStatus
  notes?: string
}

// Form Types
export interface FormField {
  name: string
  label: string
  type: 'text' | 'email' | 'password' | 'number' | 'textarea' | 'select'
  required?: boolean
  placeholder?: string
  options?: { value: string; label: string }[]
  validation?: {
    min?: number
    max?: number
    pattern?: string
    message?: string
  }
}

// UI Types
export interface ModalProps {
  isOpen: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
}

export interface TableColumn<T> {
  key: keyof T
  label: string
  render?: (value: any, item: T) => React.ReactNode
  sortable?: boolean
  width?: string
}

// Currency Types
export interface Currency {
  code: string
  name: string
  symbol: string
  rate: number // Exchange rate to base currency
}

// Common Types
export interface PaginationParams {
  page: number
  limit: number
}

export interface PaginatedResponse<T> {
  data: T[]
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
}

export interface SearchParams {
  query?: string
  category?: string
  status?: string
  dateFrom?: string
  dateTo?: string
}
