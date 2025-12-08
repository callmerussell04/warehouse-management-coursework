import { Container, Row, Col, Button, Card, Table, Badge, Spinner, Modal } from 'react-bootstrap';
import useCounterparties from './hooks/useCounterparties';
import CounterpartyModal from './form/CounterpartyModal';
import Pagination from '../../components/pagination/Pagination'; 
import usePagination from '../../components/pagination/PaginationHook';

const CounterpartiesPage = () => {
    const { currentPage } = usePagination();

    const {
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
    } = useCounterparties(currentPage);

    const totalPages = Math.ceil(paging.total / paging.size) || 1;

    const getTypeBadge = (type) => {
        return type === 'client' 
            ? <Badge bg="info">Клиент</Badge> 
            : <Badge bg="warning" text="dark">Поставщик</Badge>;
    };

    return (
        <Container className="py-5">
            <Row className="mb-4 align-items-center">
                <Col>
                    <h2>Контрагенты</h2>
                </Col>
                <Col className="text-end">
                    <Button variant="primary" onClick={openCreateModal}>
                        + Добавить
                    </Button>
                </Col>
            </Row>

            <Card className="shadow-sm border-0 rounded-4 overflow-hidden">
                <Card.Body className="p-0">
                    <Table hover responsive className="mb-0 align-middle">
                        <thead className="bg-light">
                            <tr>
                                <th className="ps-4">Имя / Название</th>
                                <th>Тип</th>
                                <th>Email</th>
                                <th>Телефон</th>
                                <th className="text-end pe-4">Действия</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading && counterparties.length === 0 ? (
                                <tr>
                                    <td colSpan="5" className="text-center py-5">
                                        <Spinner animation="border" variant="primary" />
                                    </td>
                                </tr>
                            ) : counterparties.length === 0 ? (
                                <tr>
                                    <td colSpan="5" className="text-center py-5 text-muted">
                                        Нет данных для отображения
                                    </td>
                                </tr>
                            ) : (
                                counterparties.map((item) => (
                                    <tr key={item.id}>
                                        <td className="ps-4 fw-bold">{item.name}</td>
                                        <td>{getTypeBadge(item.type)}</td>
                                        <td>{item.email || '-'}</td>
                                        <td>{item.phone_number || '-'}</td>
                                        <td className="text-end pe-4">
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

            <CounterpartyModal 
                show={showModal}
                onHide={() => setShowModal(false)}
                onSave={handleSave}
                initialData={currentCounterparty}
                loading={loading}
            />

            <Modal show={showDeleteModal} onHide={() => setShowDeleteModal(false)} centered>
                <Modal.Header closeButton>
                    <Modal.Title>Удаление контрагента</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    Вы действительно хотите удалить <strong>{currentCounterparty?.name}</strong>? 
                    Это действие нельзя отменить.
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

export default CounterpartiesPage;