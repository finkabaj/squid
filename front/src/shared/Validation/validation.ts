import { IValidationFunctionResponse } from './validation.types.ts'

const emailValidate = (email: string): IValidationFunctionResponse | null => {
  if (!email) {
    return { key: 'email', message: 'Enter email' }
  } else {
    if (!/^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w\w+)+$/.test(email)) {
      return { key: 'email', message: 'Email entered incorrectly' }
    }
  }
  return null
}
const userNameValidate = (username: string): IValidationFunctionResponse | null => {
  if (!username) {
    return { key: 'username', message: 'Enter username' }
  }
  return null
}
const firstNameValidate = (firstName: string): IValidationFunctionResponse | null => {
  if (!firstName) {
    return { key: 'first_name', message: 'Enter first name' }
  }
  return null
}
const lastNameValidate = (lastName: string): IValidationFunctionResponse | null => {
  if (!lastName) {
    return { key: 'last_name', message: 'Enter last name' }
  }
  return null
}
const dateValidate = (date_of_birth: Date): IValidationFunctionResponse | null => {
  if (!date_of_birth) {
    return { key: 'date_of_birth', message: 'Enter date of birth' }
  }
  return null
}

const passwordValidate = (password: string): IValidationFunctionResponse | null => {
  if (!password) {
    return { key: 'password', message: 'Enter password' }
  }
  return null
}

const validation = {
  emailValidate,
  passwordValidate,
  firstNameValidate,
  lastNameValidate,
  dateValidate,
  userNameValidate,
}

export default validation
