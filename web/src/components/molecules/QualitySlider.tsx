import { Slider } from '../atoms/Slider'

interface Props {
  value: number
  enabled: boolean
  onChange: (value: number) => void
}

// 非可逆形式(jpg / webp)選択時のみ有効な品質スライダー(FR-2)。
export function QualitySlider({ value, enabled, onChange }: Props) {
  const labelColor = enabled ? 'text-slate-700' : 'text-slate-400'
  return (
    <div>
      <div className="mb-1 flex items-center justify-between text-sm">
        <label htmlFor="quality" className={labelColor}>
          Quality (jpg / webp)
        </label>
        <span className={enabled ? 'font-medium text-slate-900' : 'text-slate-400'}>{value}</span>
      </div>
      <Slider id="quality" min={1} max={100} value={value} disabled={!enabled} onChange={onChange} />
    </div>
  )
}
