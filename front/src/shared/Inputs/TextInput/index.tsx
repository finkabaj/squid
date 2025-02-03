import { ChangeEvent, CSSProperties, PropsWithChildren, useContext, useEffect, useRef } from 'react'
import ValidationContext from '../../Validation/ValidationContext.ts'
import cn from '../../../utils/cn.ts'
export interface ITextInputProps {
  style?: CSSProperties
  inputStyle?: CSSProperties
  value: string
  name: string
  onChange: (val: string, name: string) => void
  disabled?: boolean
  type?: string
  className?: string
  placeholder?: string
  maxLength?: number
  minLength?: number
  label?: string
  autoComplete?: 'off' | 'on'
  serverError?: string
  autofocus?: boolean
  onFocus?: () => void
  onBlur?: () => void
  error?: boolean
  errorText?: string
  size?: string
}

const TextInput = ({
  onChange,
  value,
  children,
  className,
  style,
  size,
  serverError,
  autofocus,
  onFocus,
  onBlur,
  type,
  ...props
}: PropsWithChildren<ITextInputProps>) => {
  const inputRef = useRef<HTMLInputElement | null>(null)
  const context = useContext(ValidationContext)
  const error = context.errors[props.name]?.message || props.error
  const isError = Boolean(error || serverError)

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value, props.name)
  }
  useEffect(() => {
    if (autofocus && inputRef.current) {
      inputRef.current.focus({ preventScroll: true })
    }
  }, [])
  return (
    <div className='input-wrapper' style={style}>
      {children}
      <input
        ref={inputRef}
        data-error={isError}
        type={type ?? 'text'}
        className={cn('input', className || 0)}
        value={value}
        onChange={handleChange}
        data-size={size}
        autoComplete='false'
        style={props.inputStyle}
        onBlur={onBlur}
        {...props}
      />
      {isError && <span className={'text-error'}>{error}</span>}
    </div>
  )
}

export default TextInput
