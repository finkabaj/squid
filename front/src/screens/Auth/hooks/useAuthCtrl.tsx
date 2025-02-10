import useHttpLoaderWithServerError from '../../../shared/hooks/httpLoader/useHttpLoaderServerErr.ts'
import { useSetRecoilState } from 'recoil'
import { generateEmptyAuthState } from '../auth.context.ts'
import { useState } from 'react'
import authApi from '../auth.api.ts'
import { AuthTypeEnum } from '../../../enums/authTypeEnum.ts'
import useValidationCtrl from '../../../shared/Validation/useValidationCtrl.ts'
import validation from '../../../shared/Validation/validation.ts'
import profileAtom from '../../Profile/profile.atom.ts'

interface IProps {
  actionType: 'register' | 'login'
  setAuthType: () => void
}

const useAuthCtrl = (props: IProps) => {
  const { wait, loading, serverError } = useHttpLoaderWithServerError()
  const [authValues, setAuthValues] = useState(generateEmptyAuthState())
  const setProfileState = useSetRecoilState(profileAtom)

  const handleChange = (value: string | Date, name: string) => {
    setAuthValues((prev) => ({ ...prev, [name]: value }))
  }

  const handleSubmitCredentials = () => {
    let data
    const actionName = props.actionType === AuthTypeEnum.register ? 'register' : 'login'

    if (actionName === 'login') {
      data = {
        email: authValues.email,
        password: authValues.password,
      }
      wait(authApi.login(data), (resp) => {
        if (resp.status === 'success') {
          setProfileState(resp.body)
        }
      })
    } else {
      data = {
        username: authValues.username,
        first_name: authValues.first_name,
        last_name: authValues.last_name,
        email: authValues.email,
        password: authValues.password,
        date_of_birth: authValues.date_of_birth,
      }
      wait(authApi.register(data), (resp) => {
        if (resp.status === 'success') {
          setProfileState(resp.body)
        }
      })
    }
  }

  const validateAuthType =
    props.actionType === AuthTypeEnum.register
      ? {
          username: validation.userNameValidate,
          first_name: validation.firstNameValidate,
          last_name: validation.lastNameValidate,
          email: validation.emailValidate,
          password: validation.passwordValidate,
          date_of_birth: validation.dateValidate,
        }
      : {
          email: validation.emailValidate,
          password: validation.passwordValidate,
        }

  const validationCtrl = useValidationCtrl(handleSubmitCredentials, authValues, validateAuthType)

  return {
    validationCtrl,
    handleSubmitCredentials,
    handleChange,
    authValues,
    loading,
    serverError,
  }
}

export default useAuthCtrl
