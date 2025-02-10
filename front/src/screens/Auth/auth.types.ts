export interface ILoginValues {
  email: string
  password: string
}

export interface IRegisterValues {
  username: string
  first_name: string
  last_name: string
  email: string
  password: string
  date_of_birth: Date | undefined
}

export interface IUser {
  id: string
  username: string
  first_name: string
  last_name: string
  date_of_birth: string
  email: string
}

export interface IRefreshResponse {
  status: string
  result: {
    id: string
    username: string
    first_name: string
    last_name: string
    date_of_birth: string
    email: string
  }
}
