import { Container, Row, Col, Card, Button, Spinner, ProgressBar } from 'react-bootstrap';
import { useUser } from '../auth/context/UserContext';
import { useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

import ProductService from './products/service/ProductService';
import OrderService from './orders/service/OrderService';
import CounterpartyService from './counterparties/service/CounterpartyService';

const HomePage = () => {
    const { user } = useUser();
    const navigate = useNavigate();

    const [loading, setLoading] = useState(true);
    const [stats, setStats] = useState({
        productsTotal: 0,
        ordersTotal: 0,
        counterpartiesTotal: 0,
        recentOrders: []
    });

    const [chartData, setChartData] = useState([]);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [productsData, ordersData, counterpartiesData] = await Promise.all([
                    ProductService.getAll({ page: 1, pageSize: 1 }),
                    OrderService.getAll({ page: 1, pageSize: 50 }),
                    CounterpartyService.getAll({ page: 1, pageSize: 1 })
                ]);

                const rawOrders = ordersData.data || [];
                
                const groupedByDate = {};
                
                const sortedOrders = [...rawOrders].sort((a, b) => new Date(a.order_date) - new Date(b.order_date));

                sortedOrders.forEach(order => {
                    const dateKey = new Date(order.order_date).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' });
                    
                    if (!groupedByDate[dateKey]) {
                        groupedByDate[dateKey] = { name: dateKey, income: 0, expense: 0 };
                    }

                    if (order.order_type === 'inbound') {
                        groupedByDate[dateKey].income += 1;
                    } else {
                        groupedByDate[dateKey].expense += 1;
                    }
                });

                setChartData(Object.values(groupedByDate));

                setStats({
                    productsTotal: productsData.paging.total,
                    ordersTotal: ordersData.total_count,
                    counterpartiesTotal: counterpartiesData.paging.total,
                    recentOrders: rawOrders.slice(0, 5)
                });

            } catch (error) {
                console.error("Ошибка загрузки данных дашборда", error);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

    const StatCard = ({ title, value, icon, color, link, footer }) => (
        <Card 
            className="border-0 shadow-sm h-100" 
            onClick={() => navigate(link)} 
            style={{ cursor: 'pointer', transition: 'transform 0.2s' }}
        >
            <Card.Body>
                <div className="d-flex align-items-center justify-content-between mb-3">
                    <div>
                        <div className="text-muted small text-uppercase fw-bold mb-1">{title}</div>
                        <div className="h2 mb-0 fw-bold text-dark">
                            {loading ? <Spinner animation="border" size="sm" /> : value}
                        </div>
                    </div>
                    <div className={`bg-${color} bg-opacity-10 text-${color} rounded-circle d-flex align-items-center justify-content-center`} style={{ width: '48px', height: '48px' }}>
                        <i className={`bi ${icon} fs-4`}></i>
                    </div>
                </div>
                {footer && <div className="small text-muted">{footer}</div>}
            </Card.Body>
        </Card>
    );

    return (
        <Container className="py-4">
            <div className="mb-4 pb-3 border-bottom d-flex justify-content-between align-items-center">
                <div>
                    <h2 className="fw-light m-0">Обзор склада</h2>
                    <div className="text-muted">Добро пожаловать, <strong>{user?.full_name || 'Пользователь'}</strong>!</div>
                </div>
                <div className="text-end text-muted small">
                    Сегодня: {new Date().toLocaleDateString('ru-RU', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
                </div>
            </div>

            <Row className="g-4 mb-4">
                <Col md={6} xl={3}>
                    <StatCard 
                        title="Всего товаров" 
                        value={stats.productsTotal} 
                        icon="bi-box-seam" 
                        color="primary" 
                        link="/products"
                        footer="Позиций на складе"
                    />
                </Col>
                <Col md={6} xl={3}>
                    <StatCard 
                        title="Всего заказов" 
                        value={stats.ordersTotal} 
                        icon="bi-cart" 
                        color="warning" 
                        link="/orders"
                        footer="За все время"
                    />
                </Col>
                <Col md={6} xl={3}>
                    <StatCard 
                        title="Контрагенты" 
                        value={stats.counterpartiesTotal} 
                        icon="bi-people" 
                        color="info" 
                        link="/counterparties"
                        footer="Клиенты и поставщики"
                    />
                </Col>
                <Col md={6} xl={3}>
                    <StatCard 
                        title="Отчеты" 
                        value="PDF" 
                        icon="bi-file-earmark-text" 
                        color="danger" 
                        link="/reports"
                        footer="Скачать ведомость"
                    />
                </Col>
            </Row>

            <Row className="g-4">
                <Col lg={8}>
                    <Card className="border-0 shadow-sm h-100">
                        <Card.Header className="bg-white py-3 border-bottom">
                            <h5 className="mb-0 fw-bold">Динамика заказов (Количество)</h5>
                        </Card.Header>
                        <Card.Body style={{ minHeight: '350px' }}>
                            {loading ? (
                                <div className="d-flex justify-content-center align-items-center h-100">
                                    <Spinner animation="border" />
                                </div>
                            ) : chartData.length === 0 ? (
                                <div className="text-center text-muted py-5">Нет данных о заказах для отображения графика</div>
                            ) : (
                                <ResponsiveContainer width="100%" height={320}>
                                    <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
                                        <defs>
                                            <linearGradient id="colorIncome" x1="0" y1="0" x2="0" y2="1">
                                                <stop offset="5%" stopColor="#198754" stopOpacity={0.8}/>
                                                <stop offset="95%" stopColor="#198754" stopOpacity={0}/>
                                            </linearGradient>
                                            <linearGradient id="colorExpense" x1="0" y1="0" x2="0" y2="1">
                                                <stop offset="5%" stopColor="#dc3545" stopOpacity={0.8}/>
                                                <stop offset="95%" stopColor="#dc3545" stopOpacity={0}/>
                                            </linearGradient>
                                        </defs>
                                        <XAxis dataKey="name" />
                                        <YAxis allowDecimals={false} />
                                        <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                        <Tooltip />
                                        <Area type="monotone" dataKey="income" name="Приход" stroke="#198754" fillOpacity={1} fill="url(#colorIncome)" />
                                        <Area type="monotone" dataKey="expense" name="Расход" stroke="#dc3545" fillOpacity={1} fill="url(#colorExpense)" />
                                    </AreaChart>
                                </ResponsiveContainer>
                            )}
                        </Card.Body>
                    </Card>
                </Col>

                <Col lg={4}>
                    <div className="d-flex flex-column gap-4 h-100">
                        
                        <Card className="border-0 shadow-sm flex-fill">
                            <Card.Header className="bg-white py-3 border-bottom">
                                <h6 className="mb-0 fw-bold">Быстрые действия</h6>
                            </Card.Header>
                            <Card.Body className="d-flex flex-column gap-2">
                                <Button variant="light" className="text-start d-flex align-items-center p-3 border" onClick={() => navigate('/orders')}>
                                    <div className="bg-primary text-white rounded p-2 me-3">
                                        <i className="bi bi-plus-lg"></i>
                                    </div>
                                    <div>
                                        <div className="fw-bold">Создать заказ</div>
                                        <div className="text-muted small">Оформить приход или расход</div>
                                    </div>
                                </Button>
                                
                                <Button variant="light" className="text-start d-flex align-items-center p-3 border" onClick={() => navigate('/products')}>
                                    <div className="bg-success text-white rounded p-2 me-3">
                                        <i className="bi bi-box-seam"></i>
                                    </div>
                                    <div>
                                        <div className="fw-bold">Добавить товар</div>
                                        <div className="text-muted small">Завести новую карточку</div>
                                    </div>
                                </Button>

                                <Button variant="light" className="text-start d-flex align-items-center p-3 border" onClick={() => navigate('/reports')}>
                                    <div className="bg-secondary text-white rounded p-2 me-3">
                                        <i className="bi bi-printer"></i>
                                    </div>
                                    <div>
                                        <div className="fw-bold">Скачать отчет</div>
                                        <div className="text-muted small">Обороты за месяц</div>
                                    </div>
                                </Button>
                            </Card.Body>
                        </Card>

                        <Card className="border-0 shadow-sm">
                            <Card.Body>
                                <div className="d-flex justify-content-between mb-2">
                                    <span className="fw-bold text-muted small text-uppercase">Статус системы</span>
                                    <span className="fw-bold text-success">Online</span>
                                </div>
                                <ProgressBar now={100} variant="success" className="mb-3" style={{ height: '6px' }} />
                                <div className="small text-muted">
                                    <i className="bi bi-check-circle me-1"></i>
                                    Все сервисы работают исправно.
                                </div>
                            </Card.Body>
                        </Card>
                    </div>
                </Col>
            </Row>
        </Container>
    );
};

export default HomePage;