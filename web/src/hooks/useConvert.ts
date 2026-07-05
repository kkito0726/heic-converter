import { useCallback, useEffect, useRef, useState } from 'react'
import { isHeicFile, isTooLarge, mimeOf, replaceExtension } from '../lib/file'
import type { ConverterClient } from '../lib/rpc'
import { isConvertible, type ConvertedFile, type FileEntry } from '../types'

// クライアント側の同時リクエスト数の上限(FR-3)。
const MAX_CONCURRENCY = 3

export interface UseConvertResult {
  entries: FileEntry[]
  converting: boolean
  addFiles: (files: File[]) => void
  removeEntry: (id: string) => void
  convert: (formats: string[], quality: number) => Promise<void>
  cancel: () => void
}

// ファイルの追加〜並列変換〜揮発性URLの管理までを担うフック。
// 1リクエスト=1画像で並列発行し、1件の失敗で全体を止めない(fail-soft)。
export function useConvert(client: ConverterClient): UseConvertResult {
  const [entries, setEntries] = useState<FileEntry[]>([])
  const [converting, setConverting] = useState(false)
  const abortRef = useRef<AbortController | null>(null)
  const entriesRef = useRef(entries)
  entriesRef.current = entries

  // アンマウント時に実行中の変換を止め、Object URLをすべて解放する
  useEffect(() => {
    return () => {
      abortRef.current?.abort()
      for (const entry of entriesRef.current) revokeResults(entry)
    }
  }, [])

  const patchEntry = useCallback((id: string, patch: Partial<FileEntry>) => {
    setEntries((prev) => prev.map((e) => (e.id === id ? { ...e, ...patch } : e)))
  }, [])

  const addFiles = useCallback((files: File[]) => {
    const added = files.map(newEntry)
    setEntries((prev) => [...prev, ...added])
  }, [])

  const removeEntry = useCallback((id: string) => {
    setEntries((prev) => {
      const target = prev.find((e) => e.id === id)
      if (target) revokeResults(target)
      return prev.filter((e) => e.id !== id)
    })
  }, [])

  const cancel = useCallback(() => {
    abortRef.current?.abort()
  }, [])

  const convertOne = useCallback(
    async (id: string, file: File, formats: string[], quality: number, signal: AbortSignal) => {
      patchEntry(id, { status: 'converting' })
      try {
        const image = new Uint8Array(await file.arrayBuffer())
        const res = await client.convert({ image, formats, quality }, { signal })
        const results = res.images.map((img): ConvertedFile => {
          const data = new Uint8Array(img.data)
          const blob = new Blob([data], { type: mimeOf(img.format) })
          return {
            format: img.format,
            filename: replaceExtension(file.name, img.format),
            data,
            blob,
            url: URL.createObjectURL(blob),
          }
        })
        patchEntry(id, { status: 'done', results, error: undefined })
      } catch (error) {
        const message = signal.aborted
          ? 'Canceled'
          : error instanceof Error
            ? error.message
            : String(error)
        patchEntry(id, { status: 'error', error: message })
      }
    },
    [client, patchEntry],
  )

  const convert = useCallback(
    async (formats: string[], quality: number) => {
      const targets = entriesRef.current.filter(isConvertible)
      if (targets.length === 0 || formats.length === 0) return

      const controller = new AbortController()
      abortRef.current = controller
      setConverting(true)
      // 前回の結果を解放し、対象を待機状態へ戻す
      setEntries((prev) =>
        prev.map((e) => {
          if (!isConvertible(e)) return e
          revokeResults(e)
          return { ...e, status: 'waiting' as const, results: [], error: undefined }
        }),
      )

      const tasks = targets.map(
        (entry) => () => convertOne(entry.id, entry.file, formats, quality, controller.signal),
      )
      await runWithLimit(MAX_CONCURRENCY, tasks)
      setConverting(false)
    },
    [convertOne],
  )

  return { entries, converting, addFiles, removeEntry, convert, cancel }
}

// 選択直後の軽い検証(FR-1)を済ませたエントリを作る。
function newEntry(file: File): FileEntry {
  const unsupported = !isHeicFile(file)
  const tooLarge = isTooLarge(file)
  const error = unsupported
    ? 'Not a HEIC/HEIF file'
    : tooLarge
      ? 'Exceeds the 60 MB upload limit'
      : undefined
  return {
    id: crypto.randomUUID(),
    file,
    tooLarge,
    unsupported,
    status: error ? 'error' : 'waiting',
    results: [],
    error,
  }
}

function revokeResults(entry: FileEntry): void {
  for (const result of entry.results) URL.revokeObjectURL(result.url)
}

// 同時実行数を制限してタスクを消化する。共有イテレータを各ワーカーが順に
// 引くことで、キューの明示的な書き換えなしにタスクを分配する。
async function runWithLimit(limit: number, tasks: Array<() => Promise<void>>): Promise<void> {
  const iterator = tasks.values()
  const workers = Array.from({ length: Math.min(limit, tasks.length) }, async () => {
    for (const task of iterator) await task()
  })
  await Promise.all(workers)
}
