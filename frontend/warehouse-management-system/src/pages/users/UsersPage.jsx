import { Container, Row, Col, Button, Card, Table, Spinner, Modal, Badge } from 'react-bootstrap';
import useUsers from './hooks/useUsers';
import UserModal from './form/UserModal';
import Pagination from '../../components/pagination/Pagination';
import usePagination from '../../components/pagination/PaginationHook';

const UsersPage = () => {
    const { currentPage } = usePagination();

    const {
        users,
        paging,
        loading,
        showModal,
        showDeleteModal,
        currentUser,
        setShowModal,
        setShowDeleteModal,
        openCreateModal,
        openDeleteModal,
        handleSave,
        handleDelete
    } = useUsers(currentPage);

    const totalPages = Math.ceil(paging.total / paging.size) || 1;

    const getRoleBadge = (role) => {
        return role === 'admin' 
            ? <Badge bg="danger">Администратор</Badge> 
            : <Badge bg="primary">Сотрудник</Badge>;
    };

    const getStatusBadge = (isActive) => {
        return isActive 
            ? <Badge bg="success">Активен</Badge> 
            : <Badge bg="secondary">Не активен</Badge>;
    };

    return (
        <Container className="py-5">
            <Row className="mb-4 align-items-center">
                <Col>
                    <h2>Пользователи</h2>
                </Col>
                <Col className="text-end">
                    <Button variant="primary" onClick={openCreateModal}>
                        + Создать пользователя
                    </Button>
                </Col>
            </Row>

            <Card className="shadow-sm border-0 rounded-4 overflow-hidden">
                <Card.Body className="p-0">
                    <Table hover responsive className="mb-0 align-middle">
                        <thead className="bg-light">
                            <tr>
                                <th className="ps-4">Логин</th>
                                <th>ФИО</th>
                                <th>Email</th>
                                <th>Роль</th>
                                <th>Статус</th>
                                <th className="text-end pe-4">Действия</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading && users.length === 0 ? (
                                <tr>
                                    <td colSpan="6" className="text-center py-5">
                                        <Spinner animation="border" variant="primary" />
                                    </td>
                                </tr>
                            ) : users.length === 0 ? (
                                <tr>
                                    <td colSpan="6" className="text-center py-5 text-muted">
                                        Пользователи не найдены
                                    </td>
                                </tr>
                            ) : (
                                users.map((item) => (
                                    <tr key={item.id}>
                                        <td className="ps-4 fw-bold">{item.username}</td>
                                        <td>{item.full_name}</td>
                                        <td>{item.email}</td>
                                        <td>{getRoleBadge(item.role)}</td>
                                        <td>{getStatusBadge(item.is_active)}</td>
                                        <td className="text-end pe-4">
                                            {/* Кнопки редактирования нет, т.к. нет метода в API */}
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

            <UserModal 
                show={showModal}
                onHide={() => setShowModal(false)}
                onSave={handleSave}
                loading={loading}
            />

            <Modal show={showDeleteModal} onHide={() => setShowDeleteModal(false)} centered>
                <Modal.Header closeButton>
                    <Modal.Title>Удаление пользователя</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    Вы действительно хотите удалить пользователя <strong>{currentUser?.username}</strong>? 
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

export default UsersPage;