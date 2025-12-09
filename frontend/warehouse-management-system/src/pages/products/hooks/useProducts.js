import { useState, useEffect, useCallback } from 'react';
import ProductService from '../service/ProductService';

const useProducts = (currentPage = 1) => {
    const [products, setProducts] = useState([]);
    const [paging, setPaging] = useState({ size: 10, total: 0 });
    const [loading, setLoading] = useState(false);

    const [showModal, setShowModal] = useState(false);
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    const [currentProduct, setCurrentProduct] = useState(null);

    const fetchData = useCallback(async () => {
        setLoading(true);
        try {
            const data = await ProductService.getAll({ 
                page: currentPage, 
                pageSize: paging.size 
            });
            
            setProducts(data.items || []);
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
        setCurrentProduct(null);
        setShowModal(true);
    };

    const openEditModal = (product) => {
        setCurrentProduct(product);
        setShowModal(true);
    };

    const openDeleteModal = (product) => {
        setCurrentProduct(product);
        setShowDeleteModal(true);
    };

    const handleSave = async (formData) => {
        setLoading(true);
        try {
            if (currentProduct) {
                await ProductService.update(currentProduct.id, formData);
            } else {
                await ProductService.create(formData);
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
        if (!currentProduct) return;
        setLoading(true);
        try {
            await ProductService.delete(currentProduct.id);
            setShowDeleteModal(false);
            fetchData();
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    return {
        products,
        paging,
        loading,
        showModal,
        showDeleteModal,
        currentProduct,
        setShowModal,
        setShowDeleteModal,
        openCreateModal,
        openEditModal,
        openDeleteModal,
        handleSave,
        handleDelete
    };
};

export default useProducts;