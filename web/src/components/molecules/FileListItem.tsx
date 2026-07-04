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

// 1ファイル分の状態行。ファイル名・サイズはモノスペースで桁を揃え、
// 変換結果はダウンロードチップとして行内に展開する(FR-3 / FR-4)。
export function FileListItem({ entry, disabled, canShare, onRemove, onShare }: Props) {
  return (
    <li className="group border-b border-line px-3 py-2.5 transition-colors last:border-b-0 hover:bg-panel-raised/50">
      <div className="flex items-center gap-3">
        <Badge status={entry.status} />
        <span
          className="min-w-0 flex-1 truncate font-mono text-xs text-text"
          title={entry.file.name}
        >
          {entry.file.name}
        </span>
        <span className="shrink-0 font-mono text-[10px] tabular-nums text-faint">
          {formatBytes(entry.file.size)}
        </span>
        <Button
          variant="ghost"
          className="min-h-8 px-2 text-xs"
          aria-label={`Remove ${entry.file.name}`}
          disabled={disabled}
          onClick={() => onRemove(entry.id)}
        >
          ✕
        </Button>
      </div>
      {entry.error && (
        <p className="mt-1 pl-[6.75rem] font-mono text-[11px] text-err">{entry.error}</p>
      )}
      {entry.results.length > 0 && (
        <div className="mt-2 flex flex-wrap items-center gap-1.5 pl-[6.75rem]">
          {entry.results.map((result) => (
            <a
              key={result.format}
              href={result.url}
              download={result.filename}
              className="inline-flex min-h-8 items-center gap-1.5 rounded-sm border border-line bg-well px-2.5 font-mono text-[11px] text-dim transition-colors hover:border-amber/60 hover:text-amber-bright"
            >
              <span aria-hidden="true" className="text-amber">
                ↓
              </span>
              {result.filename}
            </a>
          ))}
          {canShare && (
            <Button
              variant="secondary"
              className="min-h-8 px-2.5 text-[11px]"
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
