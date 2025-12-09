import { Modal, Button, Form, Spinner } from 'react-bootstrap';
import { useState } from 'react';

const UserModal = ({ show, onHide, onSave, loading }) => {
    const [formData, setFormData] = useState({
        username: '',
        email: '',
        full_name: '',
        role: 'worker'
    });

    const [validated, setValidated] = useState(false);

    const handleEnter = () => {
        setValidated(false);
        setFormData({
            username: '',
            email: '',
            full_name: '',
            role: 'worker'
        });
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

    return (
        <Modal 
            show={show} 
            onHide={onHide} 
            onShow={handleEnter}
            centered
        >
            <Modal.Header closeButton>
                <Modal.Title>Новый пользователь</Modal.Title>
            </Modal.Header>
            <Form noValidate validated={validated} onSubmit={handleSubmit}>
                <Modal.Body>
                    <Form.Group className="mb-3" controlId="username">
                        <Form.Label>Логин (Username)</Form.Label>
                        <Form.Control
                            type="text"
                            name="username"
                            value={formData.username}
                            onChange={handleChange}
                            required
                            placeholder="user123"
                        />
                        <Form.Control.Feedback type="invalid">
                            Укажите логин.
                        </Form.Control.Feedback>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="full_name">
                        <Form.Label>ФИО</Form.Label>
                        <Form.Control
                            type="text"
                            name="full_name"
                            value={formData.full_name}
                            onChange={handleChange}
                            required
                            placeholder="Иванов Иван Иванович"
                        />
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="email">
                        <Form.Label>Email</Form.Label>
                        <Form.Control
                            type="email"
                            name="email"
                            value={formData.email}
                            onChange={handleChange}
                            required
                            placeholder="example@mail.com"
                        />
                        <Form.Control.Feedback type="invalid">
                            Введите корректный email.
                        </Form.Control.Feedback>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="role">
                        <Form.Label>Роль</Form.Label>
                        <Form.Select
                            name="role"
                            value={formData.role}
                            onChange={handleChange}
                        >
                            <option value="worker">Сотрудник</option>
                            <option value="admin">Администратор</option>
                        </Form.Select>
                    </Form.Group>
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={onHide} disabled={loading}>
                        Отмена
                    </Button>
                    <Button variant="primary" type="submit" disabled={loading}>
                        {loading ? <Spinner size="sm" animation="border" /> : 'Создать'}
                    </Button>
                </Modal.Footer>
            </Form>
        </Modal>
    );
};

export default UserModal;