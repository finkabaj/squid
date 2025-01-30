import { IRegisterValues } from './auth.types.ts'
import { createContext } from 'react'

interface IAuthContext {
    values: IRegisterValues
    onChange: (value: any, name: string) => void
}

export const generateEmptyAuthState = (): IRegisterValues => ({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    date_of_birth: undefined,
    username: '',
})

const AuthContext = createContext<IAuthContext>({
    values: generateEmptyAuthState(),
    onChange: () => {},
})

export default AuthContext
