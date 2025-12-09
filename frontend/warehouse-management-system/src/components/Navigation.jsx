import { Container, Nav, Navbar, Dropdown } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { useUser } from '../auth/context/UserContext';

const Navigation = () => {
    const { user } = useUser();
    const userRole = user?.role ?? null;

    return (
        <header className="text-center">
            <Navbar expand="lg" bg="dark" data-bs-theme="dark">
            <Container fluid>
                <Navbar.Brand>Склад</Navbar.Brand>
                <Navbar.Toggle aria-controls="basic-navbar-nav" />
                <Navbar.Collapse id="basic-navbar-nav">
                <Nav className="me-auto">
                        <Nav.Link as={Link} to="/">Главная</Nav.Link>
                        {userRole === 'admin' &&
                            <Nav.Link as={Link} to="/users">Пользователи</Nav.Link>
                        }
                        {user ? (<>
                        <Nav.Link as={Link} to="/products">Товары</Nav.Link>
                        <Nav.Link as={Link} to="/counterparties">Контрагенты</Nav.Link>
                        <Nav.Link as={Link} to="/orders">Заказы</Nav.Link>
                        <Nav.Link as={Link} to="/profile">Профиль</Nav.Link>
                        </>
                        ) : (
                            <Nav.Link as={Link} to="/login">Вход</Nav.Link>
                        )}
                    </Nav>
                </Navbar.Collapse>
            </Container>
            </Navbar>
        </header>

    );
};

export default Navigation;