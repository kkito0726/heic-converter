import { Button } from '../atoms/Button'

interface Props {
  count: number
  canShare: boolean
  onDownloadZip: () => void
  onShareAll: () => void
}

// 全変換結果の一括ダウンロード(zip)と一括共有(FR-4 / FR-5)。
export function ResultActions({ count, canShare, onDownloadZip, onShareAll }: Props) {
  if (count === 0) return null
  return (
    <section
      aria-label="Results"
      className="flex flex-wrap items-center gap-2 rounded-2xl border border-emerald-200 bg-emerald-50 p-4"
    >
      <p className="mr-auto text-sm font-medium text-emerald-800">{count} image(s) ready</p>
      <Button variant="secondary" onClick={onDownloadZip}>
        Download all (.zip)
      </Button>
      {canShare && <Button onClick={onShareAll}>Share…</Button>}
    </section>
  )
}
