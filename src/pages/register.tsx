import React, { useState } from "react"
import { PageProps } from "gatsby"
import usePostRegister from "../hooks/api/usePostRegister"
import { Box, Button, Input, Text } from "theme-ui"

const RegisterPage: React.FC<PageProps> = () => {
  const [name, setName] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const { mutateAsync, error } = usePostRegister()

  return (
    <div>
      <Box
        as="form"
        p={4}
        bg="muted"
        sx={{
          boxShadow:"0 0 8px rgba(0, 0, 0, 0.125)",
          minWidth: "400px",
          display: "block",
          marginTop: "-100px",
          marginLeft: "-200px",
          position: "absolute",
          left: "50%",
          top: "50%",
        }}
        onSubmit={async (event) => {
          try {
            event.preventDefault()
            await mutateAsync({ name, email, password })
          } catch (e) {
            // TODO: handle error correctly
            console.log("ERROR", e)
          }
        }}
      >
        REGISTER
        <Input
          placeholder="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          sx={{
            margin: "10px 0"
          }}
        />
        <Input
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          sx={{
            margin: "10px 0"
          }}
        />
        <Input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          sx={{
            margin: "10px 0"
          }}
        />
        <Text
          sx={{
            color: "red"
          }}
        >
          {error ? (error as Error).message : ""}
        </Text>
        <Button
          sx={{
            backgroundColor: "#959FD7",
            width: "100%",
            marginTop: "15px"
          }}
        >
          REGISTER
        </Button>
      </Box>
    </div>
  )
}

export default RegisterPage
