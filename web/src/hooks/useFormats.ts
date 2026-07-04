import { useEffect, useState } from 'react'
import type { ConverterClient } from '../lib/rpc'

export interface FormatsState {
  formats: string[]
  loading: boolean
  error: string | null
}

// 対応形式をサーバーのListFormats RPCから動的に取得する(FR-2)。
// ハードコードしないことで、サーバーの形式追加にUIが自動で追従する。
export function useFormats(client: ConverterClient): FormatsState {
  const [state, setState] = useState<FormatsState>({ formats: [], loading: true, error: null })

  useEffect(() => {
    let active = true
    client
      .listFormats({})
      .then((res) => {
        if (active) setState({ formats: res.formats, loading: false, error: null })
      })
      .catch((error: unknown) => {
        const message = error instanceof Error ? error.message : String(error)
        if (active) {
          setState({ formats: [], loading: false, error: `Failed to load formats: ${message}` })
        }
      })
    return () => {
      active = false
    }
  }, [client])

  return state
}
