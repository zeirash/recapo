import { ImageResponse } from 'next/og'
import RecapoLogo from '@/components/ui/RecapoLogo'

export const size = { width: 32, height: 32 }
export const contentType = 'image/png'

export default function Icon() {
  return new ImageResponse(
    <RecapoLogo width={32} height={32} />,
    { ...size }
  )
}
