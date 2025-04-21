import './App.css';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ActivationPage } from './ActivationPage';
import { HomePage } from './HomePage';

export const API_URL =
  import.meta.env.VITE_API_URL || 'http://localhost:8080/v1';

function App() {
  return (
    <>
      <BrowserRouter>
        <Routes>
          <Route path='/' element={<HomePage />} />
          <Route path='/activate/:token' element={<ActivationPage />} />
        </Routes>
      </BrowserRouter>
    </>
  );
}

export default App;
