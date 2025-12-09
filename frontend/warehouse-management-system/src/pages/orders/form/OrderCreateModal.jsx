import { Modal, Button, Form, Table, Badge, Row, Col } from 'react-bootstrap';
import { useState } from 'react';
import ResourceSelectorModal from './ResourceSelectorModal';
import ProductService from '../../products/service/ProductService';
import CounterpartyService from '../../counterparties/service/CounterpartyService';

const OrderCreateModal = ({ show, onHide, onCreate, loading }) => {
    const [orderType, setOrderType] = useState('inbound');
    const [orderDate, setOrderDate] = useState(new Date().toISOString().slice(0, 16));
    const [destination, setDestination] = useState('');
    
    const [selectedCounterparty, setSelectedCounterparty] = useState(null);
    const [orderItems, setOrderItems] = useState([]);

    const [showCounterpartySelector, setShowCounterpartySelector] = useState(false);
    const [showProductSelector, setShowProductSelector] = useState(false);

    const [validated, setValidated] = useState(false);

    const handleEnter = () => {
        setOrderType('inbound');
        setOrderDate(new Date().toISOString().slice(0, 16));
        setDestination('');
        setSelectedCounterparty(null);
        setOrderItems([]);
        setValidated(false);
    };

    const handleSubmit = (e) => {
        e.preventDefault();
        const form = e.currentTarget;

        if (form.checkValidity() === false) {
            e.stopPropagation();
            setValidated(true);
            return;
        }

        if (!selectedCounterparty || orderItems.length === 0) {
            alert("Необходимо выбрать контрагента и добавить хотя бы один товар.");
            return;
        }

        const payload = {
            counterparty_id: selectedCounterparty.id,
            order_type: orderType,
            order_date: new Date(orderDate).toISOString(),
            destination: orderType === 'outbound' ? destination : null,
            items: orderItems.map(item => ({
                product_id: item.product.id,
                quantity: parseInt(item.quantity)
            }))
        };
        onCreate(payload);
    };

    const addProduct = (product) => {
        if (orderItems.find(i => i.product.id === product.id)) {
            alert('Этот товар уже добавлен');
            return;
        }
        setOrderItems([...orderItems, { product, quantity: 1 }]);
        setShowProductSelector(false);
    };

    const updateQuantity = (index, value) => {
        const newItems = [...orderItems];
        newItems[index].quantity = value;
        setOrderItems(newItems);
    };

    const removeProduct = (index) => {
        const newItems = [...orderItems];
        newItems.splice(index, 1);
        setOrderItems(newItems);
    };

    const handleTypeChange = (e) => {
        setOrderType(e.target.value);
        setSelectedCounterparty(null);
        setDestination('');
        setValidated(false);
    };

    return (
        <>
            <Modal show={show} onHide={onHide} onShow={handleEnter} size="lg" backdrop="static">
                <Modal.Header closeButton>
                    <Modal.Title>Создание заказа</Modal.Title>
                </Modal.Header>
                <Form noValidate validated={validated} onSubmit={handleSubmit}>
                    <Modal.Body>
                        <Row className="mb-3">
                            <Col md={6}>
                                <Form.Group className="mb-3">
                                    <Form.Label>Тип заказа</Form.Label>
                                    <Form.Select value={orderType} onChange={handleTypeChange}>
                                        <option value="inbound">Поступление</option>
                                        <option value="outbound">Отправка</option>
                                    </Form.Select>
                                </Form.Group>
                            </Col>
                            <Col md={6}>
                                <Form.Group className="mb-3">
                                    <Form.Label>Дата заказа</Form.Label>
                                    <Form.Control 
                                        required
                                        type="datetime-local" 
                                        value={orderDate} 
                                        onChange={e => setOrderDate(e.target.value)} 
                                    />
                                </Form.Group>
                            </Col>
                        </Row>

                        <Form.Group className="mb-4">
                            <Form.Label>Контрагент ({orderType === 'inbound' ? 'Поставщик' : 'Клиент'}) <span className="text-danger">*</span></Form.Label>
                            <div className="d-flex gap-2 align-items-center">
                                {selectedCounterparty ? (
                                    <div className="border p-2 rounded flex-grow-1 d-flex justify-content-between align-items-center border-success bg-light">
                                        <span className="fw-bold">{selectedCounterparty.name}</span>
                                        <Button variant="outline-danger" size="sm" onClick={() => setSelectedCounterparty(null)}>
                                            <i className="bi bi-x"></i>
                                        </Button>
                                    </div>
                                ) : (
                                    <Button variant="outline-primary" className="w-100" onClick={() => setShowCounterpartySelector(true)}>
                                        Выбрать контрагента
                                    </Button>
                                )}
                            </div>
                        </Form.Group>

                        {orderType === 'outbound' && (
                            <Form.Group className="mb-3" controlId="destination">
                                <Form.Label>Адрес назначения <span className="text-danger">*</span></Form.Label>
                                <Form.Control 
                                    required
                                    type="text" 
                                    placeholder="Введите адрес доставки или склад получателя"
                                    value={destination}
                                    onChange={e => setDestination(e.target.value)}
                                />
                                <Form.Control.Feedback type="invalid">
                                    Обязательно укажите место назначения для отгрузки.
                                </Form.Control.Feedback>
                            </Form.Group>
                        )}
                        
                        {orderType === 'inbound' && (
                            <div className="mb-3 text-muted fst-italic">
                                <i className="bi bi-info-circle me-2"></i>
                                Товары будут отправлены на текущий склад.
                            </div>
                        )}

                        <hr />
                        
                        <div className="d-flex justify-content-between align-items-center mb-3">
                            <h5>Товары <span className="text-danger">*</span></h5>
                            <Button variant="success" size="sm" onClick={() => setShowProductSelector(true)}>
                                + Добавить товар
                            </Button>
                        </div>

                        <Table bordered size="sm">
                            <thead className="bg-light">
                                <tr>
                                    <th>Товар</th>
                                    <th style={{width: '120px'}}>Кол-во</th>
                                    <th style={{width: '50px'}}></th>
                                </tr>
                            </thead>
                            <tbody>
                                {orderItems.length === 0 ? (
                                    <tr><td colSpan="3" className="text-center text-muted">Нет добавленных товаров</td></tr>
                                ) : (
                                    orderItems.map((item, idx) => (
                                        <tr key={item.product.id}>
                                            <td className="align-middle">
                                                <div>{item.product.name}</div>
                                                <small className="text-muted">{item.product.sku}</small>
                                            </td>
                                            <td>
                                                <Form.Control 
                                                    type="number" 
                                                    min="1" 
                                                    required
                                                    value={item.quantity} 
                                                    onChange={e => updateQuantity(idx, e.target.value)}
                                                />
                                            </td>
                                            <td className="align-middle text-center">
                                                <Button variant="link" className="text-danger p-0" onClick={() => removeProduct(idx)}>
                                                    <i className="bi bi-trash"></i>
                                                </Button>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </Table>

                    </Modal.Body>
                    <Modal.Footer>
                        <Button variant="secondary" onClick={onHide}>Отмена</Button>
                        <Button 
                            variant="primary" 
                            type="submit"
                            disabled={loading}
                        >
                            Создать заказ
                        </Button>
                    </Modal.Footer>
                </Form>
            </Modal>

            <ResourceSelectorModal
                show={showCounterpartySelector}
                onHide={() => setShowCounterpartySelector(false)}
                title={`Выберите ${orderType === 'inbound' ? 'поставщика' : 'клиента'}`}
                service={CounterpartyService}
                filterParams={{ type: orderType === 'inbound' ? 'supplier' : 'client' }}
                columns={[
                    { label: 'Имя', key: 'name' },
                    { label: 'Телефон', key: 'phone_number' }
                ]}
                onSelect={(item) => {
                    setSelectedCounterparty(item);
                    setShowCounterpartySelector(false);
                }}
            />

            <ResourceSelectorModal
                show={showProductSelector}
                onHide={() => setShowProductSelector(false)}
                title="Выберите товар"
                service={ProductService}
                columns={[
                    { label: 'SKU', key: 'sku', render: (i) => <span className="font-monospace">{i.sku}</span> },
                    { label: 'Название', key: 'name' },
                    { label: 'Сток', key: 'quantity', render: (i) => <Badge bg="secondary">{i.quantity}</Badge> }
                ]}
                onSelect={addProduct}
            />
        </>
    );
};

export default OrderCreateModal;