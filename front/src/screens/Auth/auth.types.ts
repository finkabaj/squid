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
export interface IAuthResponse {
    user: IUser | null,
    token_pair: {
        access_token: string
        refresh_token: string
    }
}

export interface IUser {
    id: string
    username: string
    first_name: string
    last_name: string
    date_of_birth: Date
    email: string
}


