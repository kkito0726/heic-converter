import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { heicFile } from '../../test/fakeClient'
import { DropZone } from './DropZone'

describe('DropZone', () => {
  it('passes picked files to onFiles', async () => {
    const user = userEvent.setup()
    const onFiles = vi.fn()
    render(<DropZone onFiles={onFiles} />)

    const file = heicFile()
    await user.upload(screen.getByLabelText('Select HEIC files'), file)

    expect(onFiles).toHaveBeenCalledWith([file])
  })

  it('passes dropped files to onFiles', () => {
    const onFiles = vi.fn()
    render(<DropZone onFiles={onFiles} />)

    const file = heicFile()
    fireEvent.drop(screen.getByText(/Tap to choose/i), {
      dataTransfer: { files: [file] },
    })

    expect(onFiles).toHaveBeenCalledWith([file])
  })

  it('ignores drops while disabled', () => {
    const onFiles = vi.fn()
    render(<DropZone disabled onFiles={onFiles} />)

    fireEvent.drop(screen.getByText(/Tap to choose/i), {
      dataTransfer: { files: [heicFile()] },
    })

    expect(onFiles).not.toHaveBeenCalled()
  })
})
