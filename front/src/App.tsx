import {RouterProvider} from "react-router";
import appRouter from "./navigation/app.router.tsx";
import {urls} from "./navigation/app.urls.ts";
import {useEffect} from "react";

const App = () => {
    useEffect(() => {
        if (location.pathname === '/') {
            location.replace(urls.main)
        }
    }, [])

    return (
        <RouterProvider router={appRouter}/>
    );
};

export default App;