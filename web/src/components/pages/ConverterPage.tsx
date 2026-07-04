import { useEffect, useMemo, useState } from 'react'
import { useConvert } from '../../hooks/useConvert'
import { useFormats } from '../../hooks/useFormats'
import { createConverterClient, type ConverterClient } from '../../lib/rpc'
import { canShareFiles, isShareSupported, shareFiles, toShareFiles } from '../../lib/share'
import { makeZip } from '../../lib/zip'
import { isConvertible, type FileEntry } from '../../types'
import { DropZone } from '../organisms/DropZone'
import { FileList } from '../organisms/FileList'
import { ConvertPanel } from '../organisms/ConvertPanel'
import { ResultActions } from '../organisms/ResultActions'
import { ConverterLayout } from '../templates/ConverterLayout'

const DEFAULT_QUALITY = 90
const DEFAULT_FORMAT = 'jpg'

interface Props {
  // テストではフェイククライアントを注入する。省略時は実サーバーに接続する
  client?: ConverterClient
}

// 状態(hooks)とorganismsを結線するページ。RPC・ブラウザAPIの知識は
// lib/とhooks/に隔離されており、このページはその配線のみを担う。
export function ConverterPage({ client }: Props) {
  const rpcClient = useMemo(() => client ?? createConverterClient(), [client])
  const { formats, error: formatsError } = useFormats(rpcClient)
  const { entries, converting, addFiles, removeEntry, convert, cancel } = useConvert(rpcClient)
  const [selected, setSelected] = useState<string[]>([])
  const [quality, setQuality] = useState(DEFAULT_QUALITY)

  // 形式一覧の取得後、未選択ならデフォルト形式を選んでおく
  useEffect(() => {
    if (formats.includes(DEFAULT_FORMAT)) {
      setSelected((prev) => (prev.length > 0 ? prev : [DEFAULT_FORMAT]))
    }
  }, [formats])

  const convertibleEntries = entries.filter(isConvertible)
  const doneResults = entries.flatMap((e) => e.results)
  const progress = {
    done: convertibleEntries.filter((e) => e.status === 'done' || e.status === 'error').length,
    total: convertibleEntries.length,
  }
  const canConvert = convertibleEntries.length > 0 && selected.length > 0

  const toggleFormat = (format: string) =>
    setSelected((prev) =>
      prev.includes(format) ? prev.filter((f) => f !== format) : [...prev, format],
    )

  const handleShareEntry = async (entry: FileEntry) => {
    const files = toShareFiles(entry.results)
    if (canShareFiles(files)) await shareFiles(files)
  }

  const handleShareAll = async () => {
    const files = toShareFiles(doneResults)
    if (canShareFiles(files)) await shareFiles(files)
  }

  // 全結果を無圧縮zipにまとめ、一時的なObject URL経由でダウンロードさせる(FR-4)
  const handleDownloadZip = () => {
    const zipData = makeZip(doneResults.map((r) => ({ name: r.filename, data: r.data })))
    const url = URL.createObjectURL(new Blob([zipData], { type: 'application/zip' }))
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = 'converted-images.zip'
    anchor.click()
    URL.revokeObjectURL(url)
  }

  return (
    <ConverterLayout
      dropZone={<DropZone disabled={converting} onFiles={addFiles} />}
      fileList={
        <FileList
          entries={entries}
          converting={converting}
          canShare={isShareSupported()}
          onRemove={removeEntry}
          onShare={handleShareEntry}
        />
      }
      settings={
        <ConvertPanel
          formats={formats}
          formatsError={formatsError}
          selected={selected}
          quality={quality}
          converting={converting}
          canConvert={canConvert}
          progress={progress}
          onToggleFormat={toggleFormat}
          onQualityChange={setQuality}
          onConvert={() => void convert(selected, quality)}
          onCancel={cancel}
        />
      }
      results={
        <ResultActions
          count={doneResults.length}
          canShare={isShareSupported()}
          onDownloadZip={handleDownloadZip}
          onShareAll={handleShareAll}
        />
      }
    />
  )
}
