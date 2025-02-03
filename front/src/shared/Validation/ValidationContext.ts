import { createContext } from 'react'
import { IValidationFunctionResponse } from './validation.types.ts'

const ValidationContext = createContext<{
  errors: Record<string, IValidationFunctionResponse>
}>({
  errors: {},
})

export default ValidationContext
