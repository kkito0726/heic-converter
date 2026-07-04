import type { FileEntry } from '../../types'
import { FileListItem } from '../molecules/FileListItem'

interface Props {
  entries: FileEntry[]
  converting: boolean
  canShare: boolean
  onRemove: (id: string) => void
  onShare: (entry: FileEntry) => void
}

export function FileList({ entries, converting, canShare, onRemove, onShare }: Props) {
  if (entries.length === 0) return null
  return (
    <ul aria-label="Selected files" className="flex flex-col gap-2">
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
  )
}
