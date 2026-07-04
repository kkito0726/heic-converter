import { Button } from '../atoms/Button'
import { ProgressBar } from '../atoms/ProgressBar'
import { FormatSelector } from '../molecules/FormatSelector'
import { QualitySlider } from '../molecules/QualitySlider'

// 品質スライダーが意味を持つ非可逆形式。
const LOSSY_FORMATS = ['jpg', 'webp']

interface Props {
  formats: string[]
  formatsError: string | null
  selected: string[]
  quality: number
  converting: boolean
  canConvert: boolean
  progress: { done: number; total: number }
  onToggleFormat: (format: string) => void
  onQualityChange: (value: number) => void
  onConvert: () => void
  onCancel: () => void
}

// 変換設定(形式・品質)と実行・キャンセル・進捗をまとめた操作パネル(FR-2 / FR-3)。
export function ConvertPanel({
  formats,
  formatsError,
  selected,
  quality,
  converting,
  canConvert,
  progress,
  onToggleFormat,
  onQualityChange,
  onConvert,
  onCancel,
}: Props) {
  const qualityEnabled = selected.some((f) => LOSSY_FORMATS.includes(f))
  return (
    <section
      aria-label="Conversion settings"
      className="flex flex-col gap-5 rounded-sm border border-line bg-panel p-4"
    >
      <div>
        <h2 className="mb-2.5 font-mono text-[10px] tracking-[0.16em] text-dim uppercase">
          Output formats
        </h2>
        {formatsError ? (
          <p className="font-mono text-xs text-err">{formatsError}</p>
        ) : (
          <FormatSelector
            formats={formats}
            selected={selected}
            disabled={converting}
            onToggle={onToggleFormat}
          />
        )}
      </div>
      <QualitySlider
        value={quality}
        enabled={qualityEnabled && !converting}
        onChange={onQualityChange}
      />
      {converting ? (
        <div className="flex flex-col gap-3">
          <ProgressBar value={progress.done} max={progress.total} />
          <div className="flex items-center justify-between">
            <span className="font-mono text-xs tabular-nums text-amber-bright">
              {progress.done}&thinsp;/&thinsp;{progress.total}
              <span className="ml-2 text-faint">processing…</span>
            </span>
            <Button variant="secondary" className="min-h-9 px-3" onClick={onCancel}>
              Cancel
            </Button>
          </div>
        </div>
      ) : (
        <Button className="w-full" disabled={!canConvert} onClick={onConvert}>
          Convert
        </Button>
      )}
    </section>
  )
}
