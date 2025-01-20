import { useState } from 'react'
import TextInput, { ITextInputProps } from '../TextInput'
import styles from './PasswordInput.module.css'
import { FiEye } from 'react-icons/fi'
import { FiEyeOff } from 'react-icons/fi'

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
    <TextInput {...props}  className={styles.input} type={passwordType}>
      {passwordType === 'password' ? (
        <FiEye className={styles.icon} onClick={togglePassword} />
      ) : (
        <FiEyeOff className={styles.icon} onClick={togglePassword} />
      )}
    </TextInput>
  )
}

export default PasswordInput
