import type { ConversionStatus } from '../../types'

const STYLES: Record<ConversionStatus, string> = {
  waiting: 'bg-slate-100 text-slate-600',
  converting: 'bg-amber-100 text-amber-700',
  done: 'bg-emerald-100 text-emerald-700',
  error: 'bg-rose-100 text-rose-700',
}

const LABELS: Record<ConversionStatus, string> = {
  waiting: 'Waiting',
  converting: 'Converting',
  done: 'Done',
  error: 'Failed',
}

// ファイルごとの変換ステータス表示(FR-3)。
export function Badge({ status }: { status: ConversionStatus }) {
  return (
    <span className={`shrink-0 rounded-full px-2 py-0.5 text-xs font-medium ${STYLES[status]}`}>
      {LABELS[status]}
    </span>
  )
}
