import { Modal, Button, Form, Spinner } from 'react-bootstrap';
import { useState } from 'react';

const ProductModal = ({ show, onHide, onSave, initialData, loading }) => {
    const [formData, setFormData] = useState({
        sku: '',
        name: ''
    });

    const [validated, setValidated] = useState(false);

    const handleEnter = () => {
        setValidated(false);
        if (initialData) {
            setFormData({
                sku: initialData.sku || '',
                name: initialData.name || ''
            });
        } else {
            setFormData({
                sku: '',
                name: ''
            });
        }
    };

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        const form = e.currentTarget;
        if (form.checkValidity() === false) {
            e.stopPropagation();
            setValidated(true);
            return;
        }

        await onSave(formData);
    };

    const isEditing = !!initialData;

    return (
        <Modal 
            show={show} 
            onHide={onHide} 
            onShow={handleEnter}
            centered
        >
            <Modal.Header closeButton>
                <Modal.Title>{isEditing ? 'Редактирование товара' : 'Новый товар'}</Modal.Title>
            </Modal.Header>
            <Form noValidate validated={validated} onSubmit={handleSubmit}>
                <Modal.Body>
                    <Form.Group className="mb-3" controlId="sku">
                        <Form.Label>Артикул (SKU)</Form.Label>
                        <Form.Control
                            type="text"
                            name="sku"
                            value={formData.sku}
                            onChange={handleChange}
                            required
                            placeholder="Например: PROD-001"
                        />
                        <Form.Control.Feedback type="invalid">
                            Укажите артикул.
                        </Form.Control.Feedback>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="name">
                        <Form.Label>Название</Form.Label>
                        <Form.Control
                            type="text"
                            name="name"
                            value={formData.name}
                            onChange={handleChange}
                            required
                            placeholder="Введите название товара"
                        />
                        <Form.Control.Feedback type="invalid">
                            Укажите название.
                        </Form.Control.Feedback>
                    </Form.Group>
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={onHide} disabled={loading}>
                        Отмена
                    </Button>
                    <Button variant="primary" type="submit" disabled={loading}>
                        {loading ? <Spinner size="sm" animation="border" /> : 'Сохранить'}
                    </Button>
                </Modal.Footer>
            </Form>
        </Modal>
    );
};

export default ProductModal;