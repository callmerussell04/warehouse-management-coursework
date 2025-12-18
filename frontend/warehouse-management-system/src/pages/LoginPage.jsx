import AuthForm from "../auth/form/AuthForm";
import { Link } from "react-router-dom";

const LoginPage = () => {
    return (
        <div className="flex-grow-1 d-flex flex-column justify-content-center align-items-center">
            <AuthForm />
            <div className="mt-3 d-flex flex-column justify-content-center align-items-center">
                <Link to="/forgot-username" className="text-decoration-none">
                    Забыли логин?
                </Link>
                <Link to="/reset-password" className="text-decoration-none">
                    Забыли пароль?
                </Link>
            </div>
        </div>
    );
};

export default LoginPage;