import { createContext, useContext, useState } from 'react';

const UserContext = createContext(null);

export const UserProvider = ({ children }) => {
    const [user, setUserState] = useState(() => {
        const savedUser = localStorage.getItem('user_data');
        return savedUser ? JSON.parse(savedUser) : null;
    });

    const setUser = (userData) => {
        if (userData) {
            localStorage.setItem('user_data', JSON.stringify(userData));
        } else {
            localStorage.removeItem('user_data');
        }
        setUserState(userData);
    };

    const clearUser = () => {
        localStorage.removeItem('user_data');
        setUserState(null);
    };

    const hasRole = (role) => user?.role === role;

    return (
        <UserContext.Provider value={{ user, setUser, clearUser, hasRole }}>
            {children}
        </UserContext.Provider>
    );
};

// eslint-disable-next-line react-refresh/only-export-components
export const useUser = () => useContext(UserContext);