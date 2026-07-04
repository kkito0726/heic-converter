interface Props {
  id: string
  min: number
  max: number
  value: number
  disabled?: boolean
  onChange: (value: number) => void
}

export function Slider({ id, min, max, value, disabled, onChange }: Props) {
  return (
    <input
      id={id}
      type="range"
      className="w-full accent-indigo-600 disabled:opacity-40"
      min={min}
      max={max}
      value={value}
      disabled={disabled}
      onChange={(event) => onChange(Number(event.target.value))}
    />
  )
}
