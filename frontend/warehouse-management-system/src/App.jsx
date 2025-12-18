import { Outlet } from 'react-router-dom'
import './App.css'
import { Toaster } from 'react-hot-toast';
import Navigation from './components/Navigation';
import { UserProvider } from './auth/context/UserContext';

function App() {
  return (
    <div className="d-flex flex-column min-vh-100">
      <UserProvider>
        <Navigation />
        <main className="flex-grow-1 d-flex flex-column">
          <Outlet/>
        </main>
        <Toaster position="top-right" toastOptions={{ duration: 1000 }} />
      </UserProvider>
    </div>
  )
}

export default App