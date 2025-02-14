const isObjectEmpty = (obj: Record<string, any>) => {
  if (!obj) return false

  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-expect-error
  for (const key in obj) return false

  return true
}

const objectUtils = { isObjectEmpty }
export default objectUtils
