import { Button } from '../atoms/Button'
import { ProgressBar } from '../atoms/ProgressBar'
import { Spinner } from '../atoms/Spinner'
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

// 変換設定(形式・品質)と実行・キャンセル・進捗をまとめたパネル(FR-2 / FR-3)。
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
      className="flex flex-col gap-4 rounded-2xl border border-slate-200 bg-white p-4"
    >
      <h2 className="text-sm font-semibold text-slate-900">Output formats</h2>
      {formatsError ? (
        <p className="text-sm text-rose-600">{formatsError}</p>
      ) : (
        <FormatSelector
          formats={formats}
          selected={selected}
          disabled={converting}
          onToggle={onToggleFormat}
        />
      )}
      <QualitySlider value={quality} enabled={qualityEnabled && !converting} onChange={onQualityChange} />
      {converting ? (
        <div className="flex flex-col gap-2">
          <ProgressBar value={progress.done} max={progress.total} />
          <div className="flex items-center justify-between">
            <span className="flex items-center gap-2 text-sm text-slate-500">
              <Spinner />
              {progress.done} / {progress.total}
            </span>
            <Button variant="secondary" onClick={onCancel}>
              Cancel
            </Button>
          </div>
        </div>
      ) : (
        <Button disabled={!canConvert} onClick={onConvert}>
          Convert
        </Button>
      )}
    </section>
  )
}
