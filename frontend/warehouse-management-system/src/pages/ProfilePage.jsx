import { Container, Row, Col, Button, Badge, Card, Spinner } from 'react-bootstrap';
import { useNavigate } from "react-router-dom";
import { useUser } from "../auth/context/UserContext";
import AuthApiService from "../auth/service/AuthApiService";
import { useState } from 'react';

const ProfilePage = () => {
    const navigate = useNavigate();
    const { user, clearUser } = useUser();
    const [loading, setLoading] = useState(false);

    const handleLogout = async () => {
        setLoading(true);
        try {
            await AuthApiService.logout();
        } catch (error) {
            console.error("Logout failed on server", error);
        } finally {
            clearUser();
            setLoading(false);
            navigate("/login");
        }
    };

    const getRoleBadge = (role) => {
        // Используем более строгие бейджи без скруглений pill для энтерпрайз вида
        if (role === 'admin') return <Badge bg="danger">Администратор</Badge>;
        if (role === 'worker') return <Badge bg="primary">Сотрудник</Badge>;
        return <Badge bg="secondary">{role}</Badge>;
    };

    if (!user) return null; // Или редирект, обработано в useEffect/AuthGuard

    return (
        <Container className="py-4">
            {/* Заголовок страницы */}
            <div className="d-flex justify-content-between align-items-center mb-4 border-bottom pb-3">
                <h2 className="m-0 fw-light">Профиль пользователя</h2>
                {/* Кнопка выхода аккуратно в углу */}
                <Button 
                    variant="outline-danger" 
                    size="sm"
                    onClick={handleLogout}
                    disabled={loading}
                >
                    {loading ? <Spinner size="sm" animation="border" /> : (
                        <>
                            <i className="bi bi-box-arrow-right me-2"></i> Выйти
                        </>
                    )}
                </Button>
            </div>

            <Row className="g-4">
                {/* Левая колонка: Визитка (Аватар + Роль) */}
                <Col md={4} lg={3}>
                    <Card className="border-0 shadow-sm h-100 bg-light">
                        <Card.Body className="text-center py-5 d-flex flex-column justify-content-center align-items-center">
                            <div className="position-relative mb-3">
                                <div className="bg-white p-1 rounded-circle shadow-sm">
                                    <div 
                                        className="bg-secondary text-white rounded-circle d-flex align-items-center justify-content-center" 
                                        style={{ width: '120px', height: '120px', fontSize: '3.5rem' }}
                                    >
                                        {/* Если есть фото, можно поставить img, иначе инициалы или иконка */}
                                        <i className="bi bi-person"></i>
                                    </div>
                                </div>
                                <div className="position-absolute bottom-0 end-0">
                                    {user.is_active ? 
                                        <i className="bi bi-check-circle-fill text-success fs-4 bg-white rounded-circle" title="Активен"></i> : 
                                        <i className="bi bi-x-circle-fill text-danger fs-4 bg-white rounded-circle" title="Заблокирован"></i>
                                    }
                                </div>
                            </div>
                            
                            <h5 className="fw-bold mb-1">{user.full_name}</h5>
                            <div className="mb-3 text-muted small">{user.email}</div>
                            <div>{getRoleBadge(user.role)}</div>
                        </Card.Body>
                    </Card>
                </Col>

                {/* Правая колонка: Детали и настройки */}
                <Col md={8} lg={9}>
                    <Card className="border-0 shadow-sm h-100">
                        <Card.Header className="bg-white py-3 fw-bold border-bottom">
                            Основные данные
                        </Card.Header>
                        <Card.Body className="p-4">
                            <Row className="g-4">
                                <Col sm={6}>
                                    <div className="text-muted small text-uppercase mb-1">Имя пользователя (Логин)</div>
                                    <div className="fs-5 fw-medium d-flex align-items-center">
                                        <i className="bi bi-person-badge text-primary me-2 opacity-50"></i>
                                        {user.username}
                                    </div>
                                </Col>
                                
                                <Col sm={6}>
                                    <div className="text-muted small text-uppercase mb-1">Email адрес</div>
                                    <div className="fs-5 fw-medium d-flex align-items-center">
                                        <i className="bi bi-envelope text-primary me-2 opacity-50"></i>
                                        {user.email}
                                    </div>
                                </Col>

                                <Col sm={6}>
                                    <div className="text-muted small text-uppercase mb-1">ID пользователя</div>
                                    <div className="fs-6 font-monospace text-dark d-flex align-items-center">
                                        <i className="bi bi-fingerprint text-primary me-2 opacity-50"></i>
                                        {user.id || '—'}
                                    </div>
                                </Col>

                                <Col sm={6}>
                                    <div className="text-muted small text-uppercase mb-1">Статус аккаунта</div>
                                    <div>
                                        {user.is_active ? (
                                            <span className="text-success fw-bold">
                                                Активный
                                            </span>
                                        ) : (
                                            <span className="text-danger fw-bold">
                                                Заблокирован
                                            </span>
                                        )}
                                    </div>
                                </Col>
                            </Row>
                        </Card.Body>
                    </Card>
                </Col>
            </Row>
        </Container>
    );
};

export default ProfilePage;