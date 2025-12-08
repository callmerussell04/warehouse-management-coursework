import AuthApiService from "../auth/service/AuthApiService";
import { useNavigate } from "react-router-dom";
import { useUser } from "../auth/context/UserContext";
import { Button } from "react-bootstrap";

//TODO
const ProfilePage = () => {
    const navigate = useNavigate();
    const { clearUser } = useUser();

    return (
        <>
            <Button className="fw-bold mt-3" variant='danger' onClick={() => {AuthApiService.logout(); navigate("/"); clearUser();}}>Выйти</Button>
        </>
    );
};

export default ProfilePage;
