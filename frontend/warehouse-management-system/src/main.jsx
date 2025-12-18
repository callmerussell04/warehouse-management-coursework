import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap-icons/font/bootstrap-icons.css'
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
import ProductsPage from './pages/products/ProductsPage.jsx';
import UsersPage from './pages/users/UsersPage.jsx';
import OrdersPage from './pages/orders/OrdersPage.jsx';
import ProductHistoryPage from './pages/products/ProductHistoryPage.jsx';
import ReportsPage from './pages/reports/ReportsPage.jsx';

import { ProtectedRoute, AdminRoute } from './components/AuthGuard';

const routes = [
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
    index: true,
    path: '/',
    element: (
      <ProtectedRoute>
        <Homepage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/profile',
    element: (
      <ProtectedRoute>
        <ProfilePage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/counterparties',
    element: (
      <ProtectedRoute>
        <CounterpartiesPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/products',
    element: (
      <ProtectedRoute>
        <ProductsPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/orders',
    element: (
      <ProtectedRoute>
        <OrdersPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/products/:id/history',
    element: (
      <ProtectedRoute>
        <ProductHistoryPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/reports',
    element: (
      <ProtectedRoute>
        <ReportsPage />
      </ProtectedRoute>
    ),
  },

  {
    path: '/users',
    element: (
      <AdminRoute>
        <UsersPage />
      </AdminRoute>
    ),
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