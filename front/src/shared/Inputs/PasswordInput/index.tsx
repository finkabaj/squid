import { useState } from 'react'
import TextInput, { ITextInputProps } from '../TextInput'
import styles from './PasswordInput.module.css'
import { IoMdEye } from 'react-icons/io'
import { IoMdEyeOff } from 'react-icons/io'

const PasswordInput = (props: ITextInputProps) => {
  const [passwordType, setPasswordType] = useState('password')
  const togglePassword = () => {
    if (passwordType === 'password') {
      setPasswordType('text')
      return
    }
    setPasswordType('password')
  }
  return (
    <TextInput {...props} className={styles.input} type={passwordType}>
      {passwordType === 'password' ? (
        <IoMdEye className={styles.icon} onClick={togglePassword} />
      ) : (
        <IoMdEyeOff className={styles.icon} onClick={togglePassword} />
      )}
    </TextInput>
  )
}

export default PasswordInput
