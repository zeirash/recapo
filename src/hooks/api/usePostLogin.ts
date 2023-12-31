import { useMutation } from "react-query"
import { navigate } from "gatsby"

interface PostLoginRequestBody {
  email:    string
  password: string
}

const usePostLogin = () => {
  const postLogin = async (payload: PostLoginRequestBody) => {
    const response = await fetch(`http://localhost:3000/login`, {
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

  return useMutation(postLogin)
}

export default usePostLogin
