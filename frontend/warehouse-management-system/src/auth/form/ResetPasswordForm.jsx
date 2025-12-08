import { Button, Form, Spinner, Card } from 'react-bootstrap';
import useResetPasswordForm from '../hooks/ResetPasswordHook';
import { useState } from 'react';

const ResetPasswordForm = () => {
    const { 
        formData,
        validated,
        loading,
        errorMessage,
        step,
        handleSubmit,
        handleChange,
        goBack
    } = useResetPasswordForm();
    
    const [showPassword, setShowPassword] = useState(false);

    const renderStepContent = () => {
        switch (step) {
            case 1:
                return (
                    <>
                        <h4 className="text-center mb-4">Сброс пароля</h4>
                        <Form.Group className="mb-3" controlId='email'>
                            <Form.Label>Email</Form.Label>
                            <Form.Control 
                                type='email' 
                                name='email' 
                                required
                                placeholder="Введите email"
                                value={formData.email} 
                                onChange={handleChange} 
                            />
                            <Form.Control.Feedback type="invalid">
                                Введите корректный email.
                            </Form.Control.Feedback>
                        </Form.Group>
                        <Button className="w-100" variant='primary' type='submit' disabled={loading}>
                            {loading ? <Spinner size="sm" animation="border" /> : 'Получить код'}
                        </Button>
                    </>
                );
            case 2:
                return (
                    <>
                        <h4 className="text-center mb-4">Ввод кода</h4>
                        <div className="text-center text-muted mb-3" style={{fontSize: '0.9rem'}}>
                            Код отправлен на {formData.email}
                        </div>
                        <Form.Group className="mb-3" controlId='otp'>
                            <Form.Label>Код подтверждения</Form.Label>
                            <Form.Control 
                                type='text' 
                                name='otp' 
                                required
                                placeholder="Введите 6-значный код"
                                value={formData.otp} 
                                onChange={handleChange} 
                            />
                        </Form.Group>
                        <Button className="w-100" variant='primary' type='submit' disabled={loading}>
                            {loading ? <Spinner size="sm" animation="border" /> : 'Продолжить'}
                        </Button>
                        <Button variant="link" className="w-100 mt-2 text-decoration-none" onClick={goBack} disabled={loading}>
                            Назад
                        </Button>
                    </>
                );
            case 3:
                return (
                    <>
                        <h4 className="text-center mb-4">Новый пароль</h4>
                        <Form.Group className="mb-3" controlId='password'>
                            <Form.Label>Пароль</Form.Label>
                            <Form.Control 
                                type={showPassword ? 'text' : 'password'} 
                                name='password' 
                                required
                                placeholder="Введите новый пароль"
                                value={formData.password} 
                                onChange={handleChange} 
                            />
                        </Form.Group>
                        <Form.Group className="mb-3" controlId='passwordConfirm'>
                            <Form.Label>Подтверждение пароля</Form.Label>
                            <Form.Control 
                                type={showPassword ? 'text' : 'password'} 
                                name='passwordConfirm' 
                                required
                                placeholder="Повторите пароль"
                                value={formData.passwordConfirm} 
                                onChange={handleChange} 
                            />
                        </Form.Group>
                        
                        <Form.Group className="mb-3" controlId="showPasswordCheckbox">
                            <Form.Check 
                                type="checkbox" 
                                label="Показать пароль" 
                                onChange={() => setShowPassword(!showPassword)}
                            />
                        </Form.Group>

                        <Button className="w-100" variant='primary' type='submit' disabled={loading}>
                            {loading ? <Spinner size="sm" animation="border" /> : 'Сохранить пароль'}
                        </Button>
                    </>
                );
            default:
                return null;
        }
    };

    return (
        <div className="d-flex justify-content-center align-items-center">
            <Card className="p-4 shadow" style={{ width: '380px', borderRadius: '16px' }}>
                <Form noValidate validated={validated} onSubmit={handleSubmit}>
                    {renderStepContent()}
                    
                    {errorMessage && (
                        <div className="alert alert-danger mt-3 py-2 text-center mb-0" role="alert">
                            {errorMessage}
                        </div>
                    )}
                </Form>
            </Card>
        </div>
    );
};

export default ResetPasswordForm;