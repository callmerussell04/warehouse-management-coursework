import { Modal, Button, Table, Badge, Spinner } from 'react-bootstrap';

const OrderViewModal = ({ show, onHide, order, loading, onStatusChange, onDelete }) => {
    if (!order && !loading) return null;

    // Обновленная функция для статусов
    const getStatusBadge = (status) => {
        const variantMap = {
            'pending': 'warning',
            'processing': 'info',
            'completed': 'success',
            'canceled': 'danger'
        };

        const textMap = {
            'pending': 'Ожидает',
            'processing': 'В обработке',
            'completed': 'Выполнен',
            'canceled': 'Отменен'
        };

        return (
            <Badge bg={variantMap[status] || 'secondary'}>
                {textMap[status] || status}
            </Badge>
        );
    };

    const canChangeStatus = order?.status === 'pending' || order?.status === 'processing';
    const canDelete = order?.status === 'pending' || order?.status === 'canceled';

    return (
        <Modal show={show} onHide={onHide} size="lg" centered>
            <Modal.Header closeButton>
                <Modal.Title>Детали заказа</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                {loading || !order ? (
                    <div className="text-center py-5"><Spinner animation="border" /></div>
                ) : (
                    <>
                        <div className="d-flex justify-content-between mb-4">
                            <div>
                                <h5 className="mb-1">{order.order_type === 'inbound' ? 'Поступление' : 'Отправка'}</h5>
                                <div className="text-muted">Дата: {new Date(order.order_date).toLocaleString()}</div>
                                <div className="text-muted small mt-1 font-monospace">ID: {order.id}</div>
                            </div>
                            <div className="text-end">
                                <h5>Статус: {getStatusBadge(order.status)}</h5>
                                <div className="text-muted mt-2">
                                    Назначение: <br/>
                                    {order.order_type === 'inbound' ? (
                                        <strong>На этот склад</strong>
                                    ) : (
                                        <strong>{order.destination}</strong>
                                    )}
                                </div>
                            </div>
                        </div>

                        <h6>Товары</h6>
                        <Table striped bordered size="sm">
                            <thead>
                                <tr>
                                    <th>Товар</th>
                                    <th>Артикул</th>
                                    <th>Количество</th>
                                </tr>
                            </thead>
                            <tbody>
                                {order.items?.map((item, idx) => (
                                    <tr key={idx}>
                                        <td>{item.product_name || 'Загрузка...'}</td>
                                        <td className="font-monospace">{item.sku || '...'}</td>
                                        <td>{item.quantity}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </Table>

                        {canChangeStatus && (
                            <div className="mt-4 p-3 bg-light rounded border">
                                <h6 className="mb-3">Изменить статус</h6>
                                <div className="d-flex gap-2">
                                    {order.status === 'pending' && (
                                        <Button size="sm" variant="info" className="text-white" onClick={() => onStatusChange(order.id, 'processing')}>
                                            В обработку
                                        </Button>
                                    )}
                                    <Button size="sm" variant="success" onClick={() => onStatusChange(order.id, 'completed')}>
                                        Выполнить
                                    </Button>
                                    <Button size="sm" variant="danger" onClick={() => onStatusChange(order.id, 'canceled')}>
                                        Отменить
                                    </Button>
                                </div>
                            </div>
                        )}
                    </>
                )}
            </Modal.Body>
            <Modal.Footer>
                {canDelete && !loading && (
                    <Button variant="outline-danger" className="me-auto" onClick={() => onDelete(order.id)}>
                        Удалить заказ
                    </Button>
                )}
                <Button variant="secondary" onClick={onHide}>Закрыть</Button>
            </Modal.Footer>
        </Modal>
    );
};

export default OrderViewModal;