import * as React from "react"
import { Button } from 'theme-ui'

export interface PrimaryButtonProps {
  title: string
}

export const PrimaryButton: React.FC<PrimaryButtonProps> = ({ title }) => (
  <Button>
    {title}
  </Button>
)

export default PrimaryButton
