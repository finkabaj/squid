import {useState} from "react";
import {AuthTypeEnum} from "../../enums/authTypeEnum.ts";
import useAuthCtrl from "./hooks/useAuthCtrl.tsx";

const Auth = () => {
    const [authType, setAuthType ] = useState(AuthTypeEnum.login)
    const authCtrl = useAuthCtrl({
        actionType: authType,
        setAuthType: () => setAuthType(AuthTypeEnum.login),
    })
    return (
        <div>

        </div>
    );
};

export default Auth;