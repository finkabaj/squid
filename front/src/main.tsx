import {createRoot} from 'react-dom/client'
import App from './App.tsx'
import {RecoilRoot} from "recoil";
import './styles/sanitize.css'
import './styles/text.css'
createRoot(document.getElementById('root')!).render(
    <RecoilRoot>
        <App/>
    </RecoilRoot>
)
