import '@testing-library/jest-dom/vitest'

// jsdomはcreateObjectURL / revokeObjectURLを実装していないためスタブする。
if (typeof URL.createObjectURL !== 'function') {
  let counter = 0
  URL.createObjectURL = () => {
    counter += 1
    return `blob:mock-${counter}`
  }
  URL.revokeObjectURL = () => {}
}
