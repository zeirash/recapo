import * as React from "react"
// import { Button } from 'theme-ui'

export interface TextFieldProps {
  placeholder: string
}

export const TextField: React.FC<TextFieldProps> = ({ placeholder }) => (
  <div>
    {placeholder}
  </div>
)

export default TextField
