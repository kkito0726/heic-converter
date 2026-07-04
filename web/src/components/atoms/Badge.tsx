import type { ConversionStatus } from '../../types'

const TONES: Record<ConversionStatus, { text: string; dot: string }> = {
  waiting: { text: 'text-faint', dot: 'bg-faint' },
  converting: {
    text: 'text-accent-bright',
    dot: 'bg-accent animate-[blink_1s_ease-in-out_infinite]',
  },
  done: { text: 'text-ok', dot: 'bg-ok' },
  error: { text: 'text-err', dot: 'bg-err' },
}

const LABELS: Record<ConversionStatus, string> = {
  waiting: 'Waiting',
  converting: 'Converting',
  done: 'Done',
  error: 'Failed',
}

// インジケータランプ+モノスペースの機材風ステータス表示(FR-3)。
export function Badge({ status }: { status: ConversionStatus }) {
  const tone = TONES[status]
  return (
    <span
      className={`inline-flex w-24 shrink-0 items-center gap-1.5 font-mono text-[10px] tracking-[0.14em] uppercase ${tone.text}`}
    >
      <span aria-hidden="true" className={`size-1.5 rounded-full ${tone.dot}`} />
      {LABELS[status]}
    </span>
  )
}
