import { Outlet } from 'react-router-dom'
import './App.css'
import { Toaster } from 'react-hot-toast';

function App() {

  return (
    <>
      <Outlet/>
      <Toaster position="top-right" toastOptions={{ duration: 1000 }} />
    </>
  )
}

export default App
