import type { CSSProperties } from 'react'

interface Props {
  id: string
  min: number
  max: number
  value: number
  disabled?: boolean
  onChange: (value: number) => void
}

// 計器の目盛りを模したスライダー。塗りの位置はCSS変数(--gauge-fill)で描画する。
export function Slider({ id, min, max, value, disabled, onChange }: Props) {
  const fill = ((value - min) / (max - min)) * 100
  return (
    <input
      id={id}
      type="range"
      className="gauge"
      style={{ '--gauge-fill': `${fill}%` } as CSSProperties}
      min={min}
      max={max}
      value={value}
      disabled={disabled}
      onChange={(event) => onChange(Number(event.target.value))}
    />
  )
}
