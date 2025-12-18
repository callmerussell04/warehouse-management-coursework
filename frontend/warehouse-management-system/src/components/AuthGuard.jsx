import { Navigate, useLocation } from 'react-router-dom';
import { useUser } from '../auth/context/UserContext';

export const ProtectedRoute = ({ children }) => {
    const { user } = useUser();
    const location = useLocation();

    if (!user) {
        return <Navigate to="/login" state={{ from: location }} replace />;
    }

    return children;
};

export const AdminRoute = ({ children }) => {
    const { user } = useUser();
    const location = useLocation();

    if (!user) {
        return <Navigate to="/login" state={{ from: location }} replace />;
    }

    if (user.role !== 'admin') {
        return <Navigate to="/" replace />;
    }

    return children;
};