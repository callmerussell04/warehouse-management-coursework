import { Modal, Button, Table, Spinner, Pagination, Badge } from 'react-bootstrap';
import { useState, useEffect } from 'react';

const ResourceSelectorModal = ({ show, onHide, title, service, columns, onSelect, filterParams = {} }) => {
    const [items, setItems] = useState([]);
    const [paging, setPaging] = useState({ page: 1, size: 5, total: 0 });
    const [loading, setLoading] = useState(false);

    const fetchItems = async (page) => {
        setLoading(true);
        try {
            const params = { page, pageSize: paging.size, ...filterParams };
            const data = await service.getAll(params);
            setItems(data.items || []);
            setPaging(data.paging);
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (show) {
            fetchItems(1);
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [show, JSON.stringify(filterParams)]);

    const totalPages = Math.ceil(paging.total / paging.size) || 1;

    return (
        <Modal show={show} onHide={onHide} size="lg" centered>
            <Modal.Header closeButton>
                <Modal.Title>{title}</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                {loading ? (
                    <div className="text-center py-4"><Spinner animation="border" /></div>
                ) : (
                    <>
                        <Table hover size="sm">
                            <thead>
                                <tr>
                                    {columns.map((col, idx) => <th key={idx}>{col.label}</th>)}
                                    <th>Выбор</th>
                                </tr>
                            </thead>
                            <tbody>
                                {items.map(item => (
                                    <tr key={item.id}>
                                        {columns.map((col, idx) => (
                                            <td key={idx}>{col.render ? col.render(item) : item[col.key]}</td>
                                        ))}
                                        <td>
                                            <Button size="sm" variant="outline-primary" onClick={() => onSelect(item)}>
                                                Выбрать
                                            </Button>
                                        </td>
                                    </tr>
                                ))}
                                {items.length === 0 && <tr><td colSpan={columns.length + 1} className="text-center">Нет данных</td></tr>}
                            </tbody>
                        </Table>
                        
                        {totalPages > 1 && (
                            <div className="d-flex justify-content-center">
                                <Pagination size="sm">
                                    <Pagination.Prev disabled={paging.page === 1} onClick={() => fetchItems(paging.page - 1)} />
                                    <Pagination.Item active>{paging.page}</Pagination.Item>
                                    <Pagination.Next disabled={paging.page === totalPages} onClick={() => fetchItems(paging.page + 1)} />
                                </Pagination>
                            </div>
                        )}
                    </>
                )}
            </Modal.Body>
        </Modal>
    );
};

export default ResourceSelectorModal;