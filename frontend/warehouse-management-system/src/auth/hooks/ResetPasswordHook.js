import { useState } from 'react';
import AuthApiService from '../service/AuthApiService';
import { useNavigate } from "react-router-dom";

const useResetPasswordForm = () => {
    const [validated, setValidated] = useState(false);
    const [loading, setLoading] = useState(false);
    const [errorMessage, setErrorMessage] = useState('');
    
    const [step, setStep] = useState(1); 

    const [formData, setFormData] = useState({
        email: '',
        otp: '',
        password: '',
        passwordConfirm: '',
    });

    const navigate = useNavigate();

    const handleSubmit = async (event) => {
        const form = event.currentTarget;
        event.preventDefault();
        event.stopPropagation();
        
        setErrorMessage('');

        if (form.checkValidity() === false) {
            setValidated(true);
            return;
        }

        if (step === 3 && formData.password !== formData.passwordConfirm) {
            setErrorMessage('Пароли не совпадают');
            setValidated(true);
            return;
        }

        setLoading(true);

        try {
            if (step === 1) {
                await AuthApiService.requestOTP({ email: formData.email });
                setStep(2);
                setValidated(false);
            } 
            else if (step === 2) {
                setStep(3); 
                setValidated(false);
            } 
            else if (step === 3) {
                const payload = {
                    email: formData.email,
                    otp: formData.otp,
                    new_password: formData.password
                };
                
                await AuthApiService.resetPassword(payload);
                navigate("/login");
            }
        } catch (error) {
            console.error(error);
            setErrorMessage('Произошла ошибка. Проверьте данные и попробуйте снова.');
        } finally {
            setLoading(false);
        }
    };

    const handleChange = (event) => {
        const inputName = event.target.name;
        const inputValue = event.target.type === 'checkbox' ? event.target.checked : event.target.value;
        setFormData({
            ...formData,
            [inputName]: inputValue,
        });
    };

    const goBack = () => {
        if (step > 1) setStep(step - 1);
    }

    return {
        formData,
        validated,
        loading,
        errorMessage,
        step,
        handleSubmit,
        handleChange,
        goBack
    };
};

export default useResetPasswordForm;