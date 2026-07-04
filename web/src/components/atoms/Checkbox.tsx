interface Props {
  label: string
  checked: boolean
  disabled?: boolean
  onChange: (checked: boolean) => void
}

// ラベル全体がタップ領域になるチェックボックス(FR-6)。
export function Checkbox({ label, checked, disabled, onChange }: Props) {
  const border = checked
    ? 'border-indigo-500 bg-indigo-50 text-indigo-700'
    : 'border-slate-300 text-slate-600'
  return (
    <label
      className={`flex min-h-11 cursor-pointer items-center gap-2 rounded-lg border px-3 text-sm ${border} ${disabled ? 'pointer-events-none opacity-40' : ''}`}
    >
      <input
        type="checkbox"
        className="accent-indigo-600"
        checked={checked}
        disabled={disabled}
        onChange={(event) => onChange(event.target.checked)}
      />
      {label}
    </label>
  )
}
