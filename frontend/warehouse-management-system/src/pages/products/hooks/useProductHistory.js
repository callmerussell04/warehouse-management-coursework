import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import ProductService from '../service/ProductService';

const useProductHistory = (currentPage = 1, pageSize = 20) => {
    const { id } = useParams();
    const [historyItems, setHistoryItems] = useState([]);
    const [productInfo, setProductInfo] = useState(null);
    const [paging, setPaging] = useState({ page: currentPage, size: pageSize, total: 0 });
    const [loading, setLoading] = useState(false);
    const [productError, setProductError] = useState('');
    const [historyError, setHistoryError] = useState('');

    const [dateFrom, setDateFrom] = useState('');
    const [dateTo, setDateTo] = useState('');

    useEffect(() => {
        const loadProduct = async () => {
            setProductError('');
            try {
                const data = await ProductService.get(id);
                if (!data || typeof data !== 'object' || !data.name || !data.sku) {
                    throw new Error('Invalid product response');
                }
                setProductInfo(data);
            } catch (error) {
                console.error("Не удалось загрузить информацию о товаре", error);
                setProductInfo(null);
                setProductError('Не удалось загрузить информацию о товаре.');
            }
        };
        if (id) loadProduct();
    }, [id]);

    const fetchHistory = useCallback(async () => {
        if (!id) return;
        setLoading(true);
        setHistoryError('');
        try {
            const params = {
                page: currentPage,
                pageSize: pageSize,
                from: dateFrom || undefined,
                to: dateTo || undefined
            };

            const data = await ProductService.getHistory(id, params);
            if (!data || !Array.isArray(data.items) || !data.paging) {
                throw new Error('Invalid product history response');
            }

            setHistoryItems(data.items);
            setPaging({
                page: Number.isInteger(data.paging.page) ? data.paging.page : currentPage,
                size: Number.isInteger(data.paging.size) && data.paging.size > 0 ? data.paging.size : pageSize,
                total: Number.isInteger(data.paging.total) && data.paging.total >= 0 ? data.paging.total : 0,
            });
        } catch (error) {
            console.error(error);
            setHistoryItems([]);
            setPaging({ page: currentPage, size: pageSize, total: 0 });
            setHistoryError('Не удалось загрузить историю операций.');
        } finally {
            setLoading(false);
        }
    }, [id, currentPage, pageSize, dateFrom, dateTo]);

    useEffect(() => {
        fetchHistory();
    }, [fetchHistory]);

    return {
        productInfo,
        historyItems,
        paging,
        loading,
        error: productError || historyError,
        dateFrom,
        setDateFrom,
        dateTo,
        setDateTo,
        refresh: fetchHistory,
    };
};

export default useProductHistory;
