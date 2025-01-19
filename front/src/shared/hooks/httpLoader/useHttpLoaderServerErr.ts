import { useState } from 'react'
import useHttpLoader from '../httpLoader/useHttpLoader.ts'
import { HTTPResponse } from '../../../services/http/http.types.ts'
import { AxiosError } from 'axios'

const useHttpLoaderWithServerError = () => {
    const [serverError, setServerError] = useState('')
    const { wait: baseWait, loading } = useHttpLoader()

    const wait = <ResponseBody, T extends HTTPResponse<ResponseBody>>(
        p: Promise<T>,
        onLoad?: (v: T) => any,
        onError?: (err: AxiosError) => any
    ) => {
        setServerError('')

        baseWait(
            p,
            (resp) => {
                if (resp.status === 'error') {
                    setServerError(resp.body?.message || resp.message)
                }

                if (onLoad) onLoad(resp)
            },
            onError
        )
    }

    return { loading, serverError, wait }
}

export default useHttpLoaderWithServerError
