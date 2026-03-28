"use client"

import { Box } from '@mui/material'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { useTheme } from '@mui/material/styles'
import { alpha } from '@mui/material/styles'
import { LayoutDashboard, Package, ClipboardList, ShoppingCart, ShoppingBag, Users, CreditCard, type LucideIcon } from 'lucide-react'

interface BottomNavProps {
  selectedMenu: string
  onMenuSelect: (menu: string) => void
}

const menuItems: { id: string; labelKey: string; icon: LucideIcon; path: string }[] = [
  { id: 'dashboard',    labelKey: 'dashboard',   icon: LayoutDashboard, path: '/dashboard' },
  { id: 'products',     labelKey: 'products',    icon: Package,         path: '/products' },
  { id: 'orders',       labelKey: 'orders',      icon: ClipboardList,   path: '/orders' },
  { id: 'temp_orders',  labelKey: 'tempOrders',  icon: ShoppingCart,    path: '/temp-orders' },
  { id: 'purchase',     labelKey: 'purchase',    icon: ShoppingBag,     path: '/purchase' },
  { id: 'customers',    labelKey: 'customers',   icon: Users,           path: '/customers' },
  { id: 'subscription', labelKey: 'subscription',icon: CreditCard,      path: '/subscription' },
]

export default function BottomNav({ selectedMenu, onMenuSelect }: BottomNavProps) {
  const t = useTranslations('nav')
  const router = useRouter()
  const theme = useTheme()

  return (
    <Box
      sx={{
        display: { xs: 'flex', sm: 'none' },
        position: 'fixed',
        bottom: 0,
        left: 0,
        right: 0,
        zIndex: 100,
        bgcolor: 'background.paper',
        borderTop: '1px solid',
        borderColor: 'divider',
        overflowX: 'auto',
        scrollbarWidth: 'none',
        '&::-webkit-scrollbar': { display: 'none' },
      }}
    >
      {menuItems.map((item) => {
        const isActive = selectedMenu === item.id
        return (
          <Box
            key={item.id}
            onClick={() => { onMenuSelect(item.id); router.push(item.path) }}
            sx={{
              flex: '0 0 auto',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '3px',
              px: '14px',
              py: '10px',
              cursor: 'pointer',
              color: isActive ? 'primary.main' : 'grey.500',
              bgcolor: isActive ? alpha(theme.palette.primary.main, 0.07) : 'transparent',
              borderTop: '2px solid',
              borderColor: isActive ? 'primary.main' : 'transparent',
              transition: 'color 0.15s, background-color 0.15s',
              minWidth: '64px',
            }}
          >
            <item.icon size={20} />
            <Box sx={{ fontSize: '10px', fontWeight: isActive ? 600 : 400, whiteSpace: 'nowrap', lineHeight: 1 }}>
              {t(item.labelKey)}
            </Box>
          </Box>
        )
      })}
    </Box>
  )
}
