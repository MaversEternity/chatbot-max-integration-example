import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { MaxUI } from '@maxhub/max-ui';
import '@maxhub/max-ui/dist/styles.css';
import App from './App';
import './index.css';

(window as any).WebApp?.ready();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <MaxUI colorScheme="light">
      <App />
    </MaxUI>
  </StrictMode>,
);
