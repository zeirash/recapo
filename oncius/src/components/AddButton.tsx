"use client"

import { Button } from 'theme-ui'

type AddButtonProps = {
  onClick: () => void
  title: string
}

export default function AddButton({ onClick, title }: AddButtonProps) {
  return (
    <Button
      onClick={onClick}
      title={title}
      sx={{
        width: 36,
        minWidth: 36,
        height: 36,
        p: 0,
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        borderRadius: 'medium',
        fontSize: 2,
        fontWeight: 'bold',
      }}
    >
      +
    </Button>
  )
}
