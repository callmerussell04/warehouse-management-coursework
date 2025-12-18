import { useState, useEffect, useCallback } from 'react';
import OrderService from '../service/OrderService';
import ProductService from '../../products/service/ProductService';

const useOrders = (currentPage = 1) => {
    const [orders, setOrders] = useState([]);
    const [paging, setPaging] = useState({ size: 10, total: 0 });
    const [loading, setLoading] = useState(false);

    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showViewModal, setShowViewModal] = useState(false);
    const [currentOrder, setCurrentOrder] = useState(null);
    
    const [orderDetails, setOrderDetails] = useState(null);
    const [detailsLoading, setDetailsLoading] = useState(false);

    const fetchData = useCallback(async () => {
        setLoading(true);
        try {
            const response = await OrderService.getAll({ page: currentPage, pageSize: paging.size });
            
            setOrders(response.data || []);
            setPaging({
                total: response.total_count,
                size: response.page_size,
                page: response.page
            });
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    }, [currentPage, paging.size]);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    const openCreateModal = () => setShowCreateModal(true);

    const openViewModal = async (order) => {
        setCurrentOrder(order);
        setShowViewModal(true);
        setDetailsLoading(true);
        
        try {
            const fullOrder = await OrderService.get(order.id);
            
            const itemsWithNames = await Promise.all(fullOrder.items.map(async (item) => {
                try {
                    const product = await ProductService.get(item.product_id);
                    return { ...item, product_name: product.name, sku: product.sku };
                // eslint-disable-next-line no-unused-vars
                } catch (e) {
                    return { ...item, product_name: 'Неизвестный товар', sku: '???' };
                }
            }));

            setOrderDetails({ ...fullOrder, items: itemsWithNames });
        } catch (error) {
            console.error(error);
        } finally {
            setDetailsLoading(false);
        }
    };

    const handleCreate = async (orderData) => {
        setLoading(true);
        try {
            await OrderService.create(orderData);
            setShowCreateModal(false);
            fetchData();
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const handleStatusChange = async (id, newStatus) => {
        setDetailsLoading(true);
        try {
            const updatedOrder = await OrderService.updateStatus(id, newStatus);
            setOrderDetails(prev => ({ ...prev, status: updatedOrder.status }));
            fetchData();
        } catch (error) {
            console.error(error);
        } finally {
            setDetailsLoading(false);
        }
    };

    const handleDelete = async (id) => {
        if (!window.confirm('Вы уверены, что хотите удалить этот заказ?')) return;
        setLoading(true);
        try {
            await OrderService.delete(id);
            setShowViewModal(false);
            fetchData();
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    return {
        orders,
        paging,
        loading,
        showCreateModal,
        setShowCreateModal,
        showViewModal,
        setShowViewModal,
        currentOrder,
        orderDetails,
        detailsLoading,
        openCreateModal,
        openViewModal,
        handleCreate,
        handleStatusChange,
        handleDelete
    };
};

export default useOrders;