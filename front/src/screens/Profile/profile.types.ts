export interface IUserPatchData {
  username: string
  first_name: string
  date_of_birth: Date
}
export interface IUserPatchPassword {
  old_password: string
  password: string
}
export interface IUser {
  id: string
  username: string
  first_name: string
  last_name: string
  date_of_birth: string
  email: string
}
