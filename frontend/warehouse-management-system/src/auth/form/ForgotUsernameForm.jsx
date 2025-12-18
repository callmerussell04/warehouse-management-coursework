import { Button, Form, Spinner, Card } from 'react-bootstrap';
import useForgotUsernameForm from '../hooks/ForgotUsernameHook';

const ForgotUsernameForm = () => {
    const { 
        formData,
        validated,
        loading,
        handleSubmit,
        handleChange,
    } = useForgotUsernameForm();

    return (
        <div className="d-flex justify-content-center align-items-center">
            <Card className="p-4 shadow" style={{ width: '380px', borderRadius: '16px' }}>
                <h4 className="text-center mb-4">Восстановление логина</h4>
                <Form noValidate validated={validated} onSubmit={handleSubmit}>
                    <Form.Group className="mb-3" controlId='email'>
                        <Form.Label>Email</Form.Label>
                        <Form.Control 
                            type='email' 
                            name='email' 
                            required
                            value={formData.email} 
                            onChange={handleChange}
                            placeholder="Введите ваш email"
                        />
                        <Form.Control.Feedback type="invalid">
                            Пожалуйста, введите корректный email.
                        </Form.Control.Feedback>
                    </Form.Group>

                    <Button 
                        className="w-100 mt-2" 
                        variant='primary' 
                        type='submit'
                        disabled={loading}
                    >
                        {loading ? <Spinner size="sm" animation="border" /> : 'Получить логин'}
                    </Button>
                </Form>
            </Card>
        </div>
    );
};

export default ForgotUsernameForm;