import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import { RecoilRoot } from 'recoil'
import './styles/sanitize.css'
import './styles/text.css'
import './styles/variables.css'
import './styles/ui-el.styles.css'
createRoot(document.getElementById('root')!).render(
  <RecoilRoot>
    <App />
  </RecoilRoot>
)
