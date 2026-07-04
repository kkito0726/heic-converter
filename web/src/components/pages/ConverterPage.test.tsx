import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { fakeClient, heicFile } from '../../test/fakeClient'
import { ConverterPage } from './ConverterPage'

// ファイル選択 → 変換 → ダウンロードリンク表示までの結合テスト。
describe('ConverterPage', () => {
  it('renders formats from the server and selects jpg by default', async () => {
    render(<ConverterPage client={fakeClient()} />)

    const jpg = await screen.findByRole('checkbox', { name: 'jpg' })
    // 形式一覧の描画後、useEffectによるデフォルト選択の反映を待つ
    await waitFor(() => expect(jpg).toBeChecked())
    expect(screen.getByRole('checkbox', { name: 'png' })).not.toBeChecked()
  })

  it('converts selected files and shows download links', async () => {
    const user = userEvent.setup()
    render(<ConverterPage client={fakeClient()} />)
    await screen.findByRole('checkbox', { name: 'jpg' })

    const input = screen.getByLabelText('Select HEIC files')
    await user.upload(input, heicFile('photo.heic'))
    expect(screen.getByText('photo.heic')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Convert' }))

    const link = await screen.findByRole('link', { name: /photo\.jpg/ })
    expect(link).toHaveAttribute('download', 'photo.jpg')
    expect(link).toHaveAttribute('href', expect.stringMatching(/^blob:/))
    expect(screen.getByText('1 image(s) ready')).toBeInTheDocument()
  })

  it('disables Convert until a convertible file and a format are selected', async () => {
    const user = userEvent.setup()
    render(<ConverterPage client={fakeClient()} />)
    await screen.findByRole('checkbox', { name: 'jpg' })

    const convertButton = screen.getByRole('button', { name: 'Convert' })
    expect(convertButton).toBeDisabled()

    const input = screen.getByLabelText('Select HEIC files')
    await user.upload(input, heicFile())
    expect(convertButton).toBeEnabled()

    // 形式をすべて外すと再び無効になる
    await user.click(screen.getByRole('checkbox', { name: 'jpg' }))
    expect(convertButton).toBeDisabled()
  })

  it('shows the failure reason for a broken file (fail-soft)', async () => {
    const user = userEvent.setup()
    const client = fakeClient({
      convert: async () => {
        throw new Error('unsupported image')
      },
    })
    render(<ConverterPage client={client} />)
    await screen.findByRole('checkbox', { name: 'jpg' })

    await user.upload(screen.getByLabelText('Select HEIC files'), heicFile())
    await user.click(screen.getByRole('button', { name: 'Convert' }))

    await waitFor(() => expect(screen.getByText(/unsupported image/)).toBeInTheDocument())
    expect(screen.getByText('Failed')).toBeInTheDocument()
  })

  it('shows the privacy note (FR-7)', () => {
    render(<ConverterPage client={fakeClient()} />)
    expect(screen.getByText(/never stored on the server/)).toBeInTheDocument()
  })

  it('removes a file from the list', async () => {
    const user = userEvent.setup()
    render(<ConverterPage client={fakeClient()} />)
    await user.upload(screen.getByLabelText('Select HEIC files'), heicFile('photo.heic'))

    await user.click(screen.getByRole('button', { name: 'Remove photo.heic' }))

    expect(screen.queryByText('photo.heic')).not.toBeInTheDocument()
  })

  describe('result actions', () => {
    afterEach(() => {
      vi.unstubAllGlobals()
      vi.restoreAllMocks()
    })

    async function convertOneFile(user: ReturnType<typeof userEvent.setup>) {
      await screen.findByRole('checkbox', { name: 'jpg' })
      await user.upload(screen.getByLabelText('Select HEIC files'), heicFile('photo.heic'))
      await user.click(screen.getByRole('button', { name: 'Convert' }))
      await screen.findByText('1 image(s) ready')
    }

    it('downloads all results as a zip (FR-4)', async () => {
      // jsdomはナビゲーションを実装していないためアンカーのclickをスタブする
      const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
      const user = userEvent.setup()
      render(<ConverterPage client={fakeClient()} />)
      await convertOneFile(user)

      await user.click(screen.getByRole('button', { name: 'Download all (.zip)' }))

      expect(click).toHaveBeenCalledOnce()
    })

    it('opens the OS share sheet when supported (FR-5)', async () => {
      const share = vi.fn(async (_data?: ShareData) => {})
      vi.stubGlobal('navigator', { ...navigator, share, canShare: () => true })
      const user = userEvent.setup()
      render(<ConverterPage client={fakeClient()} />)
      await convertOneFile(user)

      await user.click(screen.getByRole('button', { name: 'Share…' }))

      expect(share).toHaveBeenCalledOnce()
      const shared = share.mock.calls[0][0]
      expect(shared?.files?.[0].name).toBe('photo.jpg')
    })

    it('shares a single file from its list item (FR-5)', async () => {
      const share = vi.fn(async () => {})
      vi.stubGlobal('navigator', { ...navigator, share, canShare: () => true })
      const user = userEvent.setup()
      render(<ConverterPage client={fakeClient()} />)
      await convertOneFile(user)

      await user.click(screen.getByRole('button', { name: 'Share' }))

      expect(share).toHaveBeenCalledOnce()
    })
  })
})
