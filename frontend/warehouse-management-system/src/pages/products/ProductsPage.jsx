import { Container, Row, Col, Button, Card, Table, Spinner, Modal, Badge } from 'react-bootstrap';
import useProducts from './hooks/useProducts';
import ProductModal from './form/ProductModal';
import Pagination from '../../components/pagination/Pagination';
import usePagination from '../../components/pagination/PaginationHook';
import { useNavigate } from 'react-router-dom';

const ProductsPage = () => {
    const navigate = useNavigate();
    const { currentPage } = usePagination();

    const {
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
    } = useProducts(currentPage);

    const totalPages = Math.ceil(paging.total / paging.size) || 1;

    const formatDate = (dateString) => {
        if (!dateString) return '-';
        return new Date(dateString).toLocaleString('ru-RU', {
            day: '2-digit', 
            month: '2-digit', 
            year: 'numeric',
            hour: '2-digit', 
            minute: '2-digit'
        });
    };

    return (
        <Container className="py-5">
            <Row className="mb-4 align-items-center">
                <Col>
                    <h2>Товары</h2>
                </Col>
                <Col className="text-end">
                    <Button variant="primary" onClick={openCreateModal}>
                        + Добавить товар
                    </Button>
                </Col>
            </Row>

            <Card className="shadow-sm border-0 rounded-4 overflow-hidden">
                <Card.Body className="p-0">
                    <Table hover responsive className="mb-0 align-middle">
                        <thead className="bg-light">
                            <tr>
                                <th className="ps-4">Артикул (SKU)</th>
                                <th>Название</th>
                                <th>Остаток</th>
                                <th>Обновлено</th>
                                <th className="text-end pe-4">Действия</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading && products.length === 0 ? (
                                <tr>
                                    <td colSpan="5" className="text-center py-5">
                                        <Spinner animation="border" variant="primary" />
                                    </td>
                                </tr>
                            ) : products.length === 0 ? (
                                <tr>
                                    <td colSpan="5" className="text-center py-5 text-muted">
                                        Нет товаров для отображения
                                    </td>
                                </tr>
                            ) : (
                                products.map((item) => (
                                    <tr key={item.id}>
                                        <td className="ps-4 fw-bold font-monospace">{item.sku}</td>
                                        <td>{item.name}</td>
                                        <td>
                                            <Badge bg={item.quantity > 0 ? "success" : "danger"}>
                                                {item.quantity} шт.
                                            </Badge>
                                        </td>
                                        <td className="text-muted small">
                                            {formatDate(item.updated_at)}
                                        </td>
                                        <td className="text-end pe-4">
                                             <Button 
                                                variant="outline-secondary" 
                                                size="sm" 
                                                className="me-2"
                                                onClick={() => navigate(`/products/${item.id}/history`)}
                                                title="История операций"
                                            >
                                                <i className="bi bi-clock-history"></i>
                                            </Button>
                                            <Button 
                                                variant="outline-primary" 
                                                size="sm" 
                                                className="me-2"
                                                onClick={() => openEditModal(item)}
                                            >
                                                <i className="bi bi-pencil"></i> Изменить
                                            </Button>
                                            <Button 
                                                variant="outline-danger" 
                                                size="sm"
                                                onClick={() => openDeleteModal(item)}
                                            >
                                                <i className="bi bi-trash"></i>
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

            <ProductModal 
                show={showModal}
                onHide={() => setShowModal(false)}
                onSave={handleSave}
                initialData={currentProduct}
                loading={loading}
            />

            <Modal show={showDeleteModal} onHide={() => setShowDeleteModal(false)} centered>
                <Modal.Header closeButton>
                    <Modal.Title>Удаление товара</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    Вы действительно хотите удалить товар <strong>{currentProduct?.name}</strong> (SKU: {currentProduct?.sku})? 
                    <br/><br/>
                    <span className="text-danger">Это действие нельзя отменить.</span>
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={() => setShowDeleteModal(false)} disabled={loading}>
                        Отмена
                    </Button>
                    <Button variant="danger" onClick={handleDelete} disabled={loading}>
                        {loading ? <Spinner size="sm" animation="border" /> : 'Удалить'}
                    </Button>
                </Modal.Footer>
            </Modal>
        </Container>
    );
};

export default ProductsPage;