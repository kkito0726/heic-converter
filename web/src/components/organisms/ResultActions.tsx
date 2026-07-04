import { Button } from '../atoms/Button'

interface Props {
  count: number
  canShare: boolean
  onDownloadZip: () => void
  onShareAll: () => void
}

// 全変換結果の一括ダウンロード(zip)と一括共有(FR-4 / FR-5)。
// 完了を示す唯一の緑ランプを添えた出力トレイ。
export function ResultActions({ count, canShare, onDownloadZip, onShareAll }: Props) {
  if (count === 0) return null
  return (
    <section
      aria-label="Results"
      className="reveal flex flex-col gap-3 rounded-sm border border-ok/30 bg-panel p-4"
    >
      <p className="flex items-center gap-2 font-mono text-xs text-ok">
        <span aria-hidden="true" className="size-1.5 rounded-full bg-ok" />
        {count} image(s) ready
      </p>
      <div className="flex flex-wrap gap-2">
        <Button variant="secondary" className="flex-1" onClick={onDownloadZip}>
          Download all (.zip)
        </Button>
        {canShare && (
          <Button className="flex-1" onClick={onShareAll}>
            Share…
          </Button>
        )}
      </div>
    </section>
  )
}
