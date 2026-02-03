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
        width: 26,
        minWidth: 40,
        height: 40,
        p: 0,
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        borderRadius: 'medium',
        fontSize: 3,
        fontWeight: 'bold',
      }}
    >
      +
    </Button>
  )
}
