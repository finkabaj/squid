import { useState } from 'react'
import { AuthTypeEnum } from '../../enums/authTypeEnum.ts'
import useAuthCtrl from './hooks/useAuthCtrl.tsx'
import AuthContext from './auth.context.ts'
import styles from './auth.module.css'
import CredentialForm from './components/CredentialForm'
import ValidationForm from '../../shared/Validation'
import Button from '../../shared/Buttons/Button'

const Auth = () => {
  const [authType, setAuthType] = useState(AuthTypeEnum.login)
  const authCtrl = useAuthCtrl({
    actionType: authType,
    setAuthType: () => setAuthType(AuthTypeEnum.login),
  })

  const changeType = () => {
    authType === AuthTypeEnum.register ? setAuthType(AuthTypeEnum.login) : setAuthType(AuthTypeEnum.register)
  }

  const renderContent = () => {
    return (
      <CredentialForm actionType={authType} onChange={changeType} serverError={authCtrl.serverError}>
        {authType === AuthTypeEnum.login && (
          <div className={styles.span_wrapper}>
            New to squid?&nbsp;
            <span className={styles.span} onClick={changeType}>
              create your account
            </span>
          </div>
        )}
        {authType === AuthTypeEnum.register && (
          <div className={styles.span_wrapper}>
            Already have an account?&nbsp;
            <span className={styles.span} onClick={changeType}>
              Sign in
            </span>
          </div>
        )}
        <Button loading={authCtrl.loading} type='submit' className={styles.auth_btn}>
          continue
        </Button>
      </CredentialForm>
    )
  }

  return (
    <AuthContext.Provider
      value={{
        values: authCtrl.authValues,
        onChange: authCtrl.handleChange,
      }}
    >
      <div className={styles.wrapper}>
        <div className={styles.header}>squid.</div>
        <ValidationForm errors={authCtrl.validationCtrl.errors} onSubmit={authCtrl.validationCtrl.handleSubmit}>
          <div className={styles.form}>{renderContent()}</div>
        </ValidationForm>
      </div>
    </AuthContext.Provider>
  )
}

export default Auth
