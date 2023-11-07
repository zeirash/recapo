import { useMutation } from "react-query"
import { navigate } from "gatsby"

interface PostLoginRequestBody {
  email: string
  password: string
}

const usePostLogin = () => {
  const postLogin = async (payload: PostLoginRequestBody) => {
    const result = await fetch(`http://localhost:3000/login`, {
      headers: { "Content-Type": "application/json" },
      method: "POST",
      body: JSON.stringify(payload),
    })

    if (result.ok) {
      navigate('/')
    } else {
      const body = await result.json()
      throw new Error(body?.error ?? "")
    }

    return result
  }

  return useMutation(postLogin)
}

export default usePostLogin
