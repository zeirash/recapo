"use client"

import { Button } from '@mui/material'
import { Plus } from 'lucide-react'

type AddButtonProps = {
  onClick: () => void
  title: string
}

export default function AddButton({ onClick, title }: AddButtonProps) {
  return (
    <Button
      onClick={onClick}
      title={title}
      variant="contained"
      disableElevation
      sx={{
        width: 36,
        minWidth: 36,
        height: 36,
        minHeight: 36,
        p: 0,
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        borderRadius: '8px',
        fontSize: '16px',
        fontWeight: 700,
      }}
    >
      <Plus size={18} />
    </Button>
  )
}
