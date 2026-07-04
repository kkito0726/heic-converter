import { useState, type DragEvent } from 'react'

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

  const border = dragging
    ? 'border-indigo-500 bg-indigo-50'
    : 'border-slate-300 bg-white hover:border-indigo-400'

  return (
    <label
      className={`block w-full cursor-pointer rounded-2xl border-2 border-dashed p-8 text-center transition-colors ${border} ${disabled ? 'pointer-events-none opacity-50' : ''}`}
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
      <p className="text-lg font-medium text-slate-700">Tap to choose HEIC files</p>
      <p className="mt-1 text-sm text-slate-400">or drag &amp; drop here — multiple files OK</p>
    </label>
  )
}
