const decodeJWT = (token: string): string | null => {
  if (token) {
    const slices = token.split('.')
    if (slices.length !== 3) {
      return null
    }
    const data = slices[1]
    const decodedData = JSON.parse(atob(data))
    if (!decodedData.user_id) {
      return null
    }
    return decodedData.user_id
  }
  return null
}

export default decodeJWT
