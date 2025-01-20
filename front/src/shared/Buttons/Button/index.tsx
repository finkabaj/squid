import { ButtonHTMLAttributes, DetailedHTMLProps, PropsWithChildren } from 'react'

export interface IButtonProps extends DetailedHTMLProps<ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement> {
  loading?: boolean
}

const Button = ({ children, loading, ...props }: PropsWithChildren<IButtonProps>) => {
  return (
    <button
      {...props}
      type={props.type ? props.type : 'button'}
      className={props.className}
      data-loading={loading ? 'true' : 'false'}
    >
      <div>{children}</div>
    </button>
  )
}

export default Button
