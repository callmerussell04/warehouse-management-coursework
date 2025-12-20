import { Container, Row, Col, Card, Table, Spinner, Form, Button, Badge, Tab, Tabs, Alert } from 'react-bootstrap';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Area, AreaChart } from 'recharts';
import useProductHistory from './hooks/useProductHistory';
import Pagination from '../../components/pagination/Pagination';

const ProductHistoryPage = () => {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    
    const currentPage = parseInt(searchParams.get('page')) || 1;
    
    const {
        productInfo,
        historyItems,
        paging,
        loading,
        error,
        dateFrom,
        setDateFrom,
        dateTo,
        setDateTo,
    } = useProductHistory(currentPage, 50);

    const totalPages = Math.ceil(paging.total / paging.size) || 1;

    const formatDate = (dateString) => {
        return new Date(dateString).toLocaleString('ru-RU', {
            day: '2-digit', month: '2-digit', year: 'numeric',
            hour: '2-digit', minute: '2-digit'
        });
    };

    const chartData = [...historyItems].reverse().map(item => ({
        ...item,
        formattedDate: new Date(item.created_at).toLocaleDateString('ru-RU', {
            day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit'
        }),
    }));

    return (
        <Container className="py-5">
            <div className="mb-4">
                <Button variant="outline-secondary" size="sm" className="mb-2" onClick={() => navigate(-1)}>
                    &larr; Назад
                </Button>
                <Row className="align-items-end">
                    <Col md={6}>
                        <h2>История</h2>
                        <h5 className="text-muted">
                            {productInfo ? `${productInfo.name} (${productInfo.sku})` : <Spinner size="sm" animation="border"/>}
                        </h5>
                    </Col>
                    <Col md={6}>
                        <Row className="g-2">
                            <Col>
                                <Form.Label className="small text-muted mb-1">С даты</Form.Label>
                                <Form.Control 
                                    type="date" 
                                    value={dateFrom} 
                                    onChange={(e) => setDateFrom(e.target.value)} 
                                />
                            </Col>
                            <Col>
                                <Form.Label className="small text-muted mb-1">По дату</Form.Label>
                                <Form.Control 
                                    type="date" 
                                    value={dateTo} 
                                    onChange={(e) => setDateTo(e.target.value)} 
                                />
                            </Col>
                        </Row>
                    </Col>
                </Row>
            </div>

            {error && <Alert variant="danger">{error}</Alert>}

            <Card className="shadow-sm border-0 rounded-4 overflow-hidden">
                <Card.Body>
                    <Tabs defaultActiveKey="table" id="history-tabs" className="mb-3">
                        
                        <Tab eventKey="table" title="Таблица">
                            <Table hover responsive className="mb-0 align-middle">
                                <thead className="bg-light">
                                    <tr>
                                        <th>Дата</th>
                                        <th>Тип операции</th>
                                        <th>Количество</th>
                                        <th>Остаток после</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {loading ? (
                                        <tr><td colSpan="4" className="text-center py-5"><Spinner animation="border" variant="primary" /></td></tr>
                                    ) : historyItems.length === 0 ? (
                                        <tr><td colSpan="4" className="text-center py-5 text-muted">История пуста</td></tr>
                                    ) : (
                                        historyItems.map((item) => (
                                            <tr key={item.id}>
                                                <td>{formatDate(item.created_at)}</td>
                                                <td>
                                                    {item.type === 'income' 
                                                        ? <Badge bg="success">Приход</Badge> 
                                                        : <Badge bg="danger">Расход</Badge>
                                                    }
                                                </td>
                                                <td className={item.type === 'income' ? 'text-success fw-bold' : 'text-danger fw-bold'}>
                                                    {item.type === 'income' ? '+' : '-'}{item.quantity}
                                                </td>
                                                <td><strong>{item.balance_after}</strong></td>
                                            </tr>
                                        ))
                                    )}
                                </tbody>
                            </Table>
                            {!loading && historyItems.length > 0 && (
                                <Pagination totalPages={totalPages} />
                            )}
                        </Tab>

                        <Tab eventKey="chart" title="График">
                            <div style={{ width: '100%', height: 400 }}>
                                {loading ? (
                                    <div className="d-flex justify-content-center align-items-center h-100">
                                        <Spinner animation="border" />
                                    </div>
                                ) : historyItems.length === 0 ? (
                                    <div className="d-flex justify-content-center align-items-center h-100 text-muted">
                                        Нет данных для графика за выбранный период
                                    </div>
                                ) : (
                                    <ResponsiveContainer>
                                        <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
                                            <defs>
                                                <linearGradient id="colorBalance" x1="0" y1="0" x2="0" y2="1">
                                                    <stop offset="5%" stopColor="#0d6efd" stopOpacity={0.8}/>
                                                    <stop offset="95%" stopColor="#0d6efd" stopOpacity={0}/>
                                                </linearGradient>
                                            </defs>
                                            <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                            <XAxis 
                                                dataKey="formattedDate" 
                                                tick={{fontSize: 12}}
                                                minTickGap={30}
                                            />
                                            <YAxis />
                                            <Tooltip />
                                            <Area 
                                                type="monotone" 
                                                dataKey="balance_after" 
                                                name="Остаток"
                                                stroke="#0d6efd" 
                                                fillOpacity={1} 
                                                fill="url(#colorBalance)" 
                                                strokeWidth={2}
                                            />
                                        </AreaChart>
                                    </ResponsiveContainer>
                                )}
                            </div>
                            <div className="text-center text-muted small mt-2">
                                Показаны последние {paging.size} операций. Используйте фильтры дат или пагинацию для просмотра других периодов.
                            </div>
                        </Tab>
                    </Tabs>
                </Card.Body>
            </Card>
        </Container>
    );
};

export default ProductHistoryPage;
