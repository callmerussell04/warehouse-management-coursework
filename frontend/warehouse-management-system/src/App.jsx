import { Outlet } from 'react-router-dom'
import './App.css'
import { Toaster } from 'react-hot-toast';
import Navigation from './components/Navigation';
import { UserProvider } from './auth/context/UserContext';

function App() {

  return (
    <>
      <UserProvider>
        <Navigation></Navigation>
        <Outlet/>
        <Toaster position="top-right" toastOptions={{ duration: 1000 }} />
      </UserProvider>
    </>
  )
}

export default App
