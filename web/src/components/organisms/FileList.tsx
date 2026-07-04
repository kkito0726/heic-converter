import type { FileEntry } from '../../types'
import { FileListItem } from '../molecules/FileListItem'

interface Props {
  entries: FileEntry[]
  converting: boolean
  canShare: boolean
  onRemove: (id: string) => void
  onShare: (entry: FileEntry) => void
}

// 選択ファイルの台帳。ヘアライン罫線で行を区切るテーブル的な見せ方にする。
export function FileList({ entries, converting, canShare, onRemove, onShare }: Props) {
  if (entries.length === 0) return null
  return (
    <section className="rounded-sm border border-line bg-panel">
      <header className="flex items-center justify-between border-b border-line px-3 py-2">
        <h2 className="font-mono text-[10px] tracking-[0.16em] text-dim uppercase">Queue</h2>
        <span className="font-mono text-[10px] tabular-nums text-faint">
          {entries.length} file(s)
        </span>
      </header>
      <ul aria-label="Selected files">
        {entries.map((entry) => (
          <FileListItem
            key={entry.id}
            entry={entry}
            disabled={converting}
            canShare={canShare}
            onRemove={onRemove}
            onShare={onShare}
          />
        ))}
      </ul>
    </section>
  )
}
