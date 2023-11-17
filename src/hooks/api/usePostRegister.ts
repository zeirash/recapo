import { useMutation } from "react-query"
import { navigate } from "gatsby"

interface PostRegisterRequestBody {
  name:     string
  email:    string
  password: string
}

const usePostRegister = () => {
  const postRegister = async (payload: PostRegisterRequestBody) => {
    const response = await fetch(`http://localhost:3000/register`, {
      headers: { "Content-Type": "application/json" },
      method: "POST",
      body: JSON.stringify(payload),
    })

    if (response.ok) {
      navigate('/')
    } else {
      const result = await response.json()
      throw new Error(result?.error ?? "")
    }

    return response
  }

  return useMutation(postRegister)
}

export default usePostRegister
