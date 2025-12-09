import { useState, useEffect, useCallback } from 'react';
import UserService from '../service/UserService';

const useUsers = (currentPage = 1) => {
    const [users, setUsers] = useState([]);
    const [paging, setPaging] = useState({ size: 10, total: 0 });
    const [loading, setLoading] = useState(false);

    const [showModal, setShowModal] = useState(false);
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    
    const [currentUser, setCurrentUser] = useState(null);

    const fetchData = useCallback(async () => {
        setLoading(true);
        try {
            const data = await UserService.getAll({ 
                page: currentPage, 
                pageSize: paging.size 
            });
            
            setUsers(data.items || []);
            setPaging(data.paging);
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    }, [currentPage, paging.size]);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    const openCreateModal = () => {
        setShowModal(true);
    };

    const openDeleteModal = (user) => {
        setCurrentUser(user);
        setShowDeleteModal(true);
    };

    const handleSave = async (formData) => {
        setLoading(true);
        try {
            await UserService.create(formData);
            setShowModal(false);
            fetchData();
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async () => {
        if (!currentUser) return;
        setLoading(true);
        try {
            await UserService.delete(currentUser.id);
            setShowDeleteModal(false);
            fetchData();
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    return {
        users,
        paging,
        loading,
        showModal,
        showDeleteModal,
        currentUser,
        setShowModal,
        setShowDeleteModal,
        openCreateModal,
        openDeleteModal,
        handleSave,
        handleDelete
    };
};

export default useUsers;