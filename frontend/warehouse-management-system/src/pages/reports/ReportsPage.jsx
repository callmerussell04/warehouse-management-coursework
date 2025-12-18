import { Container, Row, Col, Card, Form, Button, Spinner } from 'react-bootstrap';
import { useState } from 'react';
import ReportService from './service/ReportService';

const ReportsPage = () => {
    const today = new Date();
    const firstDay = new Date(today.getFullYear(), today.getMonth(), 1);

    const [dateFrom, setDateFrom] = useState(firstDay.toISOString().slice(0, 10));
    const [dateTo, setDateTo] = useState(today.toISOString().slice(0, 10));
    const [loading, setLoading] = useState(false);

    const handleDownload = async (e) => {
        e.preventDefault();
        setLoading(true);

        try {
            const blob = await ReportService.downloadTurnoverReport(dateFrom, dateTo);

            const url = window.URL.createObjectURL(new Blob([blob]));
            
            const link = document.createElement('a');
            link.href = url;
            
            link.setAttribute('download', `turnover_report_${dateFrom}_${dateTo}.pdf`);
            
            document.body.appendChild(link);
            link.click();
            link.parentNode.removeChild(link);

            window.URL.revokeObjectURL(url);

        } catch (error) {
            console.error("Ошибка при скачивании отчета:", error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <Container className="py-5">
            <h2 className="mb-4">Отчеты</h2>

            <Row>
                <Col md={6} lg={5}>
                    <Card className="shadow-sm border-0 rounded-4">
                        <Card.Header className="bg-white border-0 pt-4 px-4">
                            <h5 className="mb-0">
                                <i className="bi bi-file-earmark-pdf me-2 text-danger"></i>
                                Отчет по оборотам
                            </h5>
                            <small className="text-muted">Движение товаров за выбранный период</small>
                        </Card.Header>
                        <Card.Body className="p-4">
                            <Form onSubmit={handleDownload}>
                                <Form.Group className="mb-3">
                                    <Form.Label>Дата начала</Form.Label>
                                    <Form.Control 
                                        type="date" 
                                        required
                                        value={dateFrom} 
                                        onChange={(e) => setDateFrom(e.target.value)} 
                                    />
                                </Form.Group>

                                <Form.Group className="mb-4">
                                    <Form.Label>Дата конца</Form.Label>
                                    <Form.Control 
                                        type="date" 
                                        required
                                        value={dateTo} 
                                        onChange={(e) => setDateTo(e.target.value)} 
                                    />
                                </Form.Group>

                                <div className="d-grid">
                                    <Button 
                                        variant="primary" 
                                        type="submit" 
                                        disabled={loading}
                                        className="py-2 fw-bold"
                                    >
                                        {loading ? (
                                            <>
                                                <Spinner as="span" animation="border" size="sm" role="status" aria-hidden="true" className="me-2" />
                                                Генерация PDF...
                                            </>
                                        ) : (
                                            <>
                                                <i className="bi bi-download me-2"></i>
                                                Скачать отчет
                                            </>
                                        )}
                                    </Button>
                                </div>
                            </Form>
                        </Card.Body>
                    </Card>
                </Col>
            </Row>
        </Container>
    );
};

export default ReportsPage;