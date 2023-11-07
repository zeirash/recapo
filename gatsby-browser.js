import React from "react"
import { QueryClient, QueryClientProvider } from 'react-query';

const queryClient = new QueryClient();
const Root = function (props) {
  return (
    <QueryClientProvider client={queryClient}>{props.children}</QueryClientProvider>
  )
}

/** * Gatsby SSR config: https://www.gatsbyjs.org/docs/ssr-apis/ */
export const wrapRootElement = ({ element }) => {
  return <Root>{element}</Root>;
};
