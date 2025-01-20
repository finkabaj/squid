import { IValidationFunctionResponse } from './validation.types.ts'
import { CSSProperties, PropsWithChildren } from 'react'
import ValidationContext from './ValidationContext'

interface IProps {
  errors: Record<string, IValidationFunctionResponse>
  onSubmit: (event: any) => void
  className?: string
  styles?: CSSProperties
}

const ValidationForm = (props: PropsWithChildren<IProps>) => {
  const handleSubmit = (e: any) => {
    e.preventDefault()

    props.onSubmit(e)
  }

  return (
    <ValidationContext.Provider
      value={{
        errors: props.errors,
      }}
    >
      <form onSubmit={handleSubmit} style={props.styles} className={props.className}>
        {props.children}
      </form>
    </ValidationContext.Provider>
  )
}

export default ValidationForm
