// User role constants
export const USER_ROLES = {
  SYSTEM: 'system',
  OWNER: 'owner',
  ADMIN: 'admin',
} as const

export type UserRole = typeof USER_ROLES[keyof typeof USER_ROLES]
