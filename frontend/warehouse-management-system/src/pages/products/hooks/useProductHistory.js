import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import ProductService from '../service/ProductService';

const useProductHistory = (currentPage = 1, pageSize = 20) => {
    const { id } = useParams();
    const [historyItems, setHistoryItems] = useState([]);
    const [productInfo, setProductInfo] = useState(null);
    const [paging, setPaging] = useState({ size: pageSize, total: 0 });
    const [loading, setLoading] = useState(false);

    const [dateFrom, setDateFrom] = useState('');
    const [dateTo, setDateTo] = useState('');

    useEffect(() => {
        const loadProduct = async () => {
            try {
                const data = await ProductService.get(id);
                setProductInfo(data);
            } catch (error) {
                console.error("Не удалось загрузить информацию о товаре", error);
            }
        };
        if (id) loadProduct();
    }, [id]);

    const fetchHistory = useCallback(async () => {
        if (!id) return;
        setLoading(true);
        try {
            const params = {
                page: currentPage,
                pageSize: pageSize,
                from: dateFrom || undefined,
                to: dateTo || undefined
            };

            const data = await ProductService.getHistory(id, params);
            setHistoryItems(data.items || []);
            setPaging(data.paging);
        } catch (error) {
            console.error(error);
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
        dateFrom,
        setDateFrom,
        dateTo,
        setDateTo,
        refresh: fetchHistory
    };
};

export default useProductHistory;