export default function RecapoLogoText({ width = 210, height = 60 }: { width?: number; height?: number }) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 280 80" width={width} height={height} style={{ display: 'block' }}>
      <g transform="translate(0, 8) scale(0.75)">
        <rect x="12" y="26" width="56" height="44" rx="6" fill="#eff6ff" stroke="#3b82f6" strokeWidth="3"/>
        <rect x="9" y="16" width="62" height="14" rx="5" fill="#3b82f6"/>
        <rect x="34" y="16" width="12" height="14" rx="3" fill="#1d4ed8"/>
        <polyline points="26,48 34,57 54,38" stroke="#3b82f6" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" fill="none"/>
      </g>
      <text x="68" y="48"
        fontFamily="'Inter', 'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        fontSize="30"
        fontWeight="800"
        letterSpacing="-1.2"
        fill="#0f172a">recap<tspan fill="#3b82f6">o</tspan></text>
    </svg>
  )
}
