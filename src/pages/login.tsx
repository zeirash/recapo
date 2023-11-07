import React, { useState } from "react"
import { PageProps } from "gatsby"
import { Box, Button, Container, Input, Text } from 'theme-ui'
import usePostLogin from "../hooks/api/usePostLogin"

const LoginPage: React.FC<PageProps> = () => {
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const { mutateAsync, error } = usePostLogin()

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
            await mutateAsync({ email, password })
          } catch (e) {
            // TODO: handle error correctly
            console.log("ERROR", e)
          }
        }}
      >
        LOGIN
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
          {error ? "wrong email or password" : ""}
        </Text>
        <Button sx={{
          backgroundColor: "#959FD7",
          width: "100%",
          marginTop: "15px"
        }}>
          LOGIN
        </Button>
      </Box>
    </div>
  )
}

export default LoginPage
