import { formatBytes } from '../../lib/file'
import type { FileEntry } from '../../types'
import { Badge } from '../atoms/Badge'
import { Button } from '../atoms/Button'

interface Props {
  entry: FileEntry
  disabled?: boolean
  canShare: boolean
  onRemove: (id: string) => void
  onShare: (entry: FileEntry) => void
}

// 1ファイル分の状態表示と、変換結果ごとのダウンロード・共有導線(FR-3 / FR-4)。
export function FileListItem({ entry, disabled, canShare, onRemove, onShare }: Props) {
  return (
    <li className="rounded-xl border border-slate-200 bg-white p-3">
      <div className="flex items-center gap-2">
        <Badge status={entry.status} />
        <span className="min-w-0 flex-1 truncate text-sm text-slate-800" title={entry.file.name}>
          {entry.file.name}
        </span>
        <span className="shrink-0 text-xs text-slate-400">{formatBytes(entry.file.size)}</span>
        <Button
          variant="ghost"
          className="min-h-9 px-2"
          aria-label={`Remove ${entry.file.name}`}
          disabled={disabled}
          onClick={() => onRemove(entry.id)}
        >
          ✕
        </Button>
      </div>
      {entry.error && <p className="mt-1 text-xs text-rose-600">{entry.error}</p>}
      {entry.results.length > 0 && (
        <div className="mt-2 flex flex-wrap items-center gap-2">
          {entry.results.map((result) => (
            <a
              key={result.format}
              href={result.url}
              download={result.filename}
              className="inline-flex min-h-9 items-center rounded-lg bg-slate-100 px-3 text-xs font-medium text-slate-700 hover:bg-slate-200"
            >
              ↓ {result.filename}
            </a>
          ))}
          {canShare && (
            <Button
              variant="secondary"
              className="min-h-9 px-3 text-xs"
              onClick={() => onShare(entry)}
            >
              Share
            </Button>
          )}
        </div>
      )}
    </li>
  )
}
