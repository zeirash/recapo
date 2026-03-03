export default function RecapoLogo({ width = 50, height = 60 }: { width?: number; height?: number }) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 80 80" width={width} height={height} style={{ display: 'block' }}>
      <rect x="12" y="24" width="56" height="44" rx="6" fill="#eff6ff" stroke="#3b82f6" strokeWidth="3"/>
      <rect x="9" y="16" width="62" height="14" rx="5" fill="#3b82f6"/>
      <rect x="34" y="16" width="12" height="14" rx="3" fill="#1d4ed8"/>
      <polyline points="26,48 34,57 54,38" stroke="#3b82f6" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" fill="none"/>
    </svg>
  )
}
