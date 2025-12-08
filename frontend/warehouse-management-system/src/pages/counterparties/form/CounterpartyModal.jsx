import { Modal, Button, Form, Spinner } from 'react-bootstrap';
import { useState } from 'react'; // useEffect больше не нужен для сброса

const CounterpartyModal = ({ show, onHide, onSave, initialData, loading }) => {
    const [formData, setFormData] = useState({
        name: '',
        type: 'client',
        phone_number: '',
        email: ''
    });

    const [validated, setValidated] = useState(false);

    const handleEnter = () => {
        setValidated(false);
        if (initialData) {
            setFormData({
                name: initialData.name,
                type: initialData.type,
                phone_number: initialData.phone_number || '',
                email: initialData.email || ''
            });
        } else {
            setFormData({
                name: '',
                type: 'client',
                phone_number: '',
                email: ''
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
            onShow={handleEnter} // <--- Инициализация происходит здесь
            centered
        >
            <Modal.Header closeButton>
                <Modal.Title>{isEditing ? 'Редактирование контрагента' : 'Новый контрагент'}</Modal.Title>
            </Modal.Header>
            <Form noValidate validated={validated} onSubmit={handleSubmit}>
                <Modal.Body>
                    <Form.Group className="mb-3" controlId="name">
                        <Form.Label>Название / Имя</Form.Label>
                        <Form.Control
                            type="text"
                            name="name"
                            value={formData.name}
                            onChange={handleChange}
                            required
                            placeholder="Введите название"
                        />
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="type">
                        <Form.Label>Тип</Form.Label>
                        <Form.Select
                            name="type"
                            value={formData.type}
                            onChange={handleChange}
                            disabled={isEditing}
                        >
                            <option value="client">Клиент</option>
                            <option value="supplier">Поставщик</option>
                        </Form.Select>
                        {isEditing && <Form.Text className="text-muted">Тип контрагента нельзя изменить.</Form.Text>}
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="email">
                        <Form.Label>Email</Form.Label>
                        <Form.Control
                            type="email"
                            name="email"
                            value={formData.email}
                            onChange={handleChange}
                            placeholder="example@mail.com"
                        />
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="phone_number">
                        <Form.Label>Телефон</Form.Label>
                        <Form.Control
                            type="text"
                            name="phone_number"
                            value={formData.phone_number}
                            onChange={handleChange}
                            placeholder="+7 (999) 000-00-00"
                        />
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

export default CounterpartyModal;