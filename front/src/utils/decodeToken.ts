export interface ITokenData{
    user_id: string;
    email: string;
    created_at: string;
    expires_at: string;
}

const decodeJWT = (token: string): Pick<ITokenData, 'user_id'> | null => {
    if (token){
        const slices = token.split('.');
        if (slices.length !== 3) {
            return null;
        }
        const data = slices[1];
        const decodedData = JSON.parse(atob(data));
        if (!decodedData.user_id) {
            return null;
        }
        return {user_id: decodedData.user_id};
    }


}
export default decodeJWT