import { Checkbox } from '../atoms/Checkbox'

interface Props {
  formats: string[]
  selected: string[]
  disabled?: boolean
  onToggle: (format: string) => void
}

// ListFormats RPCから取得した形式をチェックボックスで複数選択する(FR-2)。
export function FormatSelector({ formats, selected, disabled, onToggle }: Props) {
  return (
    <div className="grid grid-cols-3 gap-2">
      {formats.map((format) => (
        <Checkbox
          key={format}
          label={format}
          checked={selected.includes(format)}
          disabled={disabled}
          onChange={() => onToggle(format)}
        />
      ))}
    </div>
  )
}
