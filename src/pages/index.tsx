import React from "react"
import { PageProps } from "gatsby"
import { PrimaryButton } from "../components/PrimaryButton/PrimaryButton"

const IndexPage: React.FC<PageProps> = () => {
  return (
    <div>
      <PrimaryButton title="hello"/>
      test index
    </div>
  )
}

export default IndexPage
