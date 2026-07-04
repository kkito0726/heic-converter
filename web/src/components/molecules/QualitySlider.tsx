import { Slider } from '../atoms/Slider'

interface Props {
  value: number
  enabled: boolean
  onChange: (value: number) => void
}

// 非可逆形式(jpg / webp)選択時のみ有効な品質スライダー(FR-2)。
export function QualitySlider({ value, enabled, onChange }: Props) {
  return (
    <div className={enabled ? '' : 'opacity-40'}>
      <div className="mb-2.5 flex items-baseline justify-between">
        <label
          htmlFor="quality"
          className="font-mono text-[10px] tracking-[0.14em] text-dim uppercase"
        >
          Quality (jpg / webp)
        </label>
        <span className="font-mono text-sm tabular-nums text-amber-bright">
          {String(value).padStart(3, '0')}
        </span>
      </div>
      <Slider id="quality" min={1} max={100} value={value} disabled={!enabled} onChange={onChange} />
    </div>
  )
}
