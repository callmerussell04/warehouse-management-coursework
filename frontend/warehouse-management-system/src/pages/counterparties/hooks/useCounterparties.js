import { useState, useEffect, useCallback } from 'react';
import CounterpartyService from '../service/CounterpartyService';

const useCounterparties = (currentPage = 1) => {
    const [counterparties, setCounterparties] = useState([]);
    const [paging, setPaging] = useState({ size: 10, total: 0 });
    const [loading, setLoading] = useState(false);

    const [showModal, setShowModal] = useState(false);
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    const [currentCounterparty, setCurrentCounterparty] = useState(null);

    const fetchData = useCallback(async () => {
        setLoading(true);
        try {
            const data = await CounterpartyService.getAll({ 
                page: currentPage, 
                pageSize: paging.size 
            });
            
            setCounterparties(data.items || []);
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
        setCurrentCounterparty(null);
        setShowModal(true);
    };

    const openEditModal = (counterparty) => {
        setCurrentCounterparty(counterparty);
        setShowModal(true);
    };

    const openDeleteModal = (counterparty) => {
        setCurrentCounterparty(counterparty);
        setShowDeleteModal(true);
    };

    const handleSave = async (formData) => {
        setLoading(true);
        try {
            if (currentCounterparty) {
                // eslint-disable-next-line no-unused-vars
                const { type, ...updateData } = formData;
                await CounterpartyService.update(currentCounterparty.id, updateData);
            } else {
                await CounterpartyService.create(formData);
            }
            setShowModal(false);
            fetchData();
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async () => {
        if (!currentCounterparty) return;
        setLoading(true);
        try {
            await CounterpartyService.delete(currentCounterparty.id);
            setShowDeleteModal(false);
            fetchData();
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    return {
        counterparties,
        paging,
        loading,
        showModal,
        showDeleteModal,
        currentCounterparty,
        setShowModal,
        setShowDeleteModal,
        openCreateModal,
        openEditModal,
        openDeleteModal,
        handleSave,
        handleDelete
    };
};

export default useCounterparties;