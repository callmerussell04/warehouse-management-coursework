import 'bootstrap/dist/css/bootstrap.min.css';
import './index.css'
import App from './App.jsx'
import ReactDOM from 'react-dom/client'
import { RouterProvider, createBrowserRouter } from 'react-router-dom';
import React from 'react';
import ErrorPage from './pages/ErrorPage';
import Homepage from './pages/Homepage';
import LoginPage from './pages/LoginPage.jsx';
import ProfilePage from './pages/ProfilePage.jsx';
import ForgotUsernamePage from './pages/ForgotUsernamePage.jsx';
import ResetPasswordPage from './pages/ResetPasswordPage.jsx';
import CounterpartiesPage from './pages/counterparties/CounterpartiesPage.jsx';

const routes = [
  {
    index: true,
    path: '/',
    element: <Homepage />,
  },
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/forgot-username',
    element: <ForgotUsernamePage />,
  },
  {
    path: '/reset-password',
    element: <ResetPasswordPage />,
  },
  {
    path: '/profile',
    element: <ProfilePage />,
  },
  {
    path: '/counterparties',
    element: <CounterpartiesPage />,
  },
];

const router = createBrowserRouter([
  {
    path: '/',
    element: <App routes={routes} />,
    children: routes,
    errorElement: <ErrorPage />,
  },
]);

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>,
);
