import { Container, Row, Col, Button, Card, Table, Spinner, Badge } from 'react-bootstrap';
import useOrders from './hooks/useOrders';
import Pagination from '../../components/pagination/Pagination';
import usePagination from '../../components/pagination/PaginationHook';
import OrderCreateModal from './form/OrderCreateModal';
import OrderViewModal from './form/OrderViewModal';

const OrdersPage = () => {
    const { currentPage } = usePagination();
    
    const {
        orders,
        paging,
        loading,
        showCreateModal,
        setShowCreateModal,
        showViewModal,
        setShowViewModal,
        orderDetails,
        detailsLoading,
        openCreateModal,
        openViewModal,
        handleCreate,
        handleStatusChange,
        handleDelete
    } = useOrders(currentPage);

    const totalPages = Math.ceil(paging.total / paging.size) || 1;

    const formatDate = (dateString) => {
        return new Date(dateString).toLocaleString('ru-RU', {
            day: '2-digit', month: '2-digit', year: 'numeric',
            hour: '2-digit', minute: '2-digit'
        });
    };

    const getTypeBadge = (type) => {
        return type === 'inbound' 
            ? <Badge bg="primary">Поступление</Badge> 
            : <Badge bg="warning" text="dark">Отправка</Badge>;
    };

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

    return (
        <Container className="py-5">
            <Row className="mb-4 align-items-center">
                <Col><h2>Заказы</h2></Col>
                <Col className="text-end">
                    <Button variant="primary" onClick={openCreateModal}>+ Новый заказ</Button>
                </Col>
            </Row>

            <Card className="shadow-sm border-0 rounded-4 overflow-hidden">
                <Card.Body className="p-0">
                    <Table hover responsive className="mb-0 align-middle">
                        <thead className="bg-light">
                            <tr>
                                <th className="ps-4">Дата</th>
                                <th>Тип</th>
                                <th>Статус</th>
                                <th>Назначение</th>
                                <th className="text-end pe-4">Действия</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading && orders.length === 0 ? (
                                <tr><td colSpan="5" className="text-center py-5"><Spinner animation="border" variant="primary" /></td></tr>
                            ) : orders.length === 0 ? (
                                <tr><td colSpan="5" className="text-center py-5 text-muted">Заказов нет</td></tr>
                            ) : (
                                orders.map((item) => (
                                    <tr key={item.id}>
                                        <td className="ps-4">{formatDate(item.order_date)}</td>
                                        <td>{getTypeBadge(item.order_type)}</td>
                                        <td>{getStatusBadge(item.status)}</td>
                                        <td className="small">
                                            {item.order_type === 'inbound' ? (
                                                <span className="text-muted fst-italic">На этот склад</span>
                                            ) : (
                                                item.destination || <span className="text-danger">Не указано</span>
                                            )}
                                        </td>
                                        <td className="text-end pe-4">
                                            <Button variant="outline-primary" size="sm" onClick={() => openViewModal(item)}>
                                                <i className="bi bi-eye"></i> Просмотр
                                            </Button>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </Table>
                </Card.Body>
            </Card>

            <Pagination totalPages={totalPages} />

            <OrderCreateModal 
                show={showCreateModal}
                onHide={() => setShowCreateModal(false)}
                onCreate={handleCreate}
                loading={loading}
            />

            <OrderViewModal 
                show={showViewModal}
                onHide={() => setShowViewModal(false)}
                order={orderDetails}
                loading={detailsLoading}
                onStatusChange={handleStatusChange}
                onDelete={handleDelete}
            />
        </Container>
    );
};

export default OrdersPage;