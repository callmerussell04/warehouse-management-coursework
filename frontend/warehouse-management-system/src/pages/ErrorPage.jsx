import { Alert, Button, Container } from 'react-bootstrap';
import { isRouteErrorResponse, useNavigate, useRouteError } from 'react-router-dom';

const ErrorPage = () => {
    const navigate = useNavigate();
    const error = useRouteError();
    const isNotFound = isRouteErrorResponse(error) && error.status === 404;

    return (
        <Container fluid className="p-2 row justify-content-center">
            <Container className='col-md-6'>
                <Alert variant="danger">
                    {isNotFound
                        ? 'Страница не найдена'
                        : 'При отображении страницы произошла ошибка'}
                </Alert>
                <Button className="w-25 mt-2" variant="primary" onClick={() => navigate(-1)}>Назад</Button>
            </Container>
        </Container>
    );
};

export default ErrorPage;
