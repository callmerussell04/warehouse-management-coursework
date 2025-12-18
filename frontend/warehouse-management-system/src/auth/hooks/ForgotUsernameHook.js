import { useState } from 'react';
import AuthApiService from '../service/AuthApiService';
import { useNavigate } from "react-router-dom";

const useForgotUsernameForm = () => {
    const [validated, setValidated] = useState(false);
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        email: ''
    });

    const navigate = useNavigate();

    const handleSubmit = async (event) => {
        const form = event.currentTarget;
        event.preventDefault();
        event.stopPropagation();

        if (form.checkValidity() === false) {
            setValidated(true);
            return;
        }

        setLoading(true);
        try {
            await AuthApiService.forgotUsername(formData);
            navigate("/login");
        } catch (error) {
            console.error("Failed to recover username", error);
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

    return {
        formData,
        validated,
        loading,
        handleSubmit,
        handleChange
    };
};

export default useForgotUsernameForm;