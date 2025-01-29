import { PropsWithChildren, useContext } from 'react'
import authContext from '../../auth.context.ts'
import styles from './CredentialForm.module.css'
import TextInput from '../../../../shared/Inputs/TextInput'
import PasswordInput from '../../../../shared/Inputs/PasswordInput'
import DateInput from '../../../../shared/Inputs/DateInput'

interface IFormProps {
  actionType: 'register' | 'login'
  onChange: () => void
  serverError: string
}

const CredentialForm = (props: PropsWithChildren<IFormProps>) => {
  const context = useContext(authContext)
  return (
    <>
      {props.actionType === 'login' && (
        <div>
          <div className={styles.form_header}>Sign in</div>
          <div className={styles.body}>
            <TextInput
              serverError={props.serverError}
              name='email'
              value={context.values.email}
              onChange={context.onChange}
              placeholder='email'
              type='text'
              autoComplete='off'
              autofocus='true'
            />
            <PasswordInput
              serverError={props.serverError}
              label='enter your password'
              name='password'
              value={context.values.password}
              onChange={context.onChange}
              placeholder='password'
              autoComplete='off'
            />
          </div>
        </div>
      )}
      {props.actionType === 'register' && (
        <div>
          <div className={styles.form_header}>Create account</div>
          <div className={styles.body}>
            <TextInput
              serverError={props.serverError}
              name='username'
              value={context.values.username}
              onChange={context.onChange}
              placeholder='username'
              type='text'
              autoComplete='off'
              autofocus='true'
            />
            <TextInput
              serverError={props.serverError}
              name='first_name'
              value={context.values.first_name}
              onChange={context.onChange}
              placeholder='first name'
              type='text'
              autoComplete='off'
            />
            <TextInput
              serverError={props.serverError}
              name='last_name'
              value={context.values.last_name}
              onChange={context.onChange}
              placeholder='last name'
              type='text'
              autoComplete='off'
            />
            <TextInput
              serverError={props.serverError}
              name='email'
              value={context.values.email}
              onChange={context.onChange}
              placeholder='email'
              type='text'
              autoComplete='off'
              inputStyle={{
                color: '2F3D53'
              }}
            />
            <DateInput
              name={'date_of_birth'}
              onChange={context.onChange}
              placeholder={'date of birth'}
            />
            <PasswordInput
              serverError={props.serverError}
              label='enter your password'
              name='password'
              value={context.values.password}
              onChange={context.onChange}
              placeholder='password'
              autoComplete='off'
            />
          </div>
        </div>
      )}
      {props.children}
    </>
  )
}

export default CredentialForm
