import { useState, type DragEvent } from 'react'

// 四隅のコーナーマーク(ファインダーのフレーミングを模す)。
const CORNERS = [
  'top-2 left-2 border-t border-l',
  'top-2 right-2 border-t border-r',
  'bottom-2 left-2 border-b border-l',
  'bottom-2 right-2 border-b border-r',
] as const

interface Props {
  disabled?: boolean
  onFiles: (files: File[]) => void
}

// ファイルピッカーとドラッグ&ドロップの両方でファイルを受け付ける(FR-1)。
// スマホではラベルのタップでOSのファイル/写真選択が開く。
export function DropZone({ disabled, onFiles }: Props) {
  const [dragging, setDragging] = useState(false)

  const handleDrop = (event: DragEvent) => {
    event.preventDefault()
    setDragging(false)
    if (disabled) return
    onFiles(Array.from(event.dataTransfer.files))
  }

  const tone = dragging
    ? 'border-amber bg-amber/5'
    : 'border-line-strong bg-panel hover:border-amber/60 hover:bg-panel-raised'

  return (
    <label
      className={`group relative block w-full cursor-pointer rounded-sm border border-dashed px-8 py-12 text-center transition-all duration-200 ${tone} ${disabled ? 'pointer-events-none opacity-50' : ''}`}
      onDragOver={(event) => {
        event.preventDefault()
        setDragging(true)
      }}
      onDragLeave={() => setDragging(false)}
      onDrop={handleDrop}
    >
      <input
        type="file"
        aria-label="Select HEIC files"
        accept=".heic,.heif"
        multiple
        disabled={disabled}
        className="sr-only"
        onChange={(event) => {
          onFiles(Array.from(event.target.files ?? []))
          // 同じファイルをもう一度選び直せるように選択状態をリセットする
          event.target.value = ''
        }}
      />
      {CORNERS.map((pos) => (
        <span
          key={pos}
          aria-hidden="true"
          className={`absolute size-3 transition-colors duration-200 ${pos} ${dragging ? 'border-amber' : 'border-line-strong group-hover:border-amber/60'}`}
        />
      ))}
      <span
        aria-hidden="true"
        className={`mb-4 inline-flex size-11 items-center justify-center rounded-full border text-lg transition-all duration-200 ${dragging ? 'border-amber text-amber' : 'border-line-strong text-dim group-hover:border-amber/60 group-hover:text-amber'}`}
      >
        ↓
      </span>
      <p className="text-lg font-semibold tracking-tight text-text">Tap to choose HEIC files</p>
      <p className="mt-2 font-mono text-[11px] tracking-[0.08em] text-faint uppercase">
        or drag &amp; drop here — multiple files OK
      </p>
    </label>
  )
}
