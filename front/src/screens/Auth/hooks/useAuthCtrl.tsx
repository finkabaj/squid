import useHttpLoaderWithServerError from '../../../shared/hooks/httpLoader/useHttpLoaderServerErr.ts'
import {useRecoilState, useSetRecoilState} from 'recoil'
import authAtom from '../auth.atom.ts'
import { generateEmptyAuthState } from '../auth.context.ts'
import { useState } from 'react'
import authApi from '../auth.api.ts'
import { AuthTypeEnum } from '../../../enums/authTypeEnum.ts'
import useValidationCtrl from '../../../shared/Validation/useValidationCtrl.ts'
import validation from '../../../shared/Validation/validation.ts'
import profileApi from '../../Profile/profile.api.ts'
import profileAtom from '../../Profile/profile.atom.ts'

interface IProps {
    actionType: 'register' | 'login'
    setAuthType: () => void
}

const useAuthCtrl = (props: IProps) => {
    const { wait, loading, serverError } = useHttpLoaderWithServerError()
    const setAuthState = useSetRecoilState(authAtom)
    const [authValues, setAuthValues] = useState(generateEmptyAuthState())
    const [profileState, setProfileState] = useRecoilState(profileAtom)

    const handleChange = (value: any, name: string) => {
        setAuthValues((prev) => ({ ...prev, [name]: value }))
    }

    const handleSubmitCredentials = () => {
        const actionName = props.actionType === AuthTypeEnum.register ? 'register' : 'login'
        wait(
            authApi[actionName]({
                username: authValues.username,
                first_name: authValues.first_name,
                last_name: authValues.last_name,
                email: authValues.email,
                password: authValues.password,
                date_of_birth: authValues.date_of_birth,
            }),
            (resp) => {
                if (resp.status === 'success' && actionName === 'register') {
                    props.setAuthType()
                }
                if (resp.status === 'success' && actionName === 'login') {
                    setAuthState((prev) => ({
                        ...prev,
                        token_pair: {
                            access_token: resp.body.result.access_token,
                            refresh_token: resp.body.result.refresh_token,
                        },
                    }))
                    localStorage.setItem('access_token', resp.body.token_pair.access_token)

                    profileApi.getUser(profileState.user_id).then((res: any) => {
                        setProfileState(res.body)
                    })
                }
            }
        )
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

    const validationCtrl = useValidationCtrl(
        handleSubmitCredentials,
        authValues,
        validateAuthType
    )

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
