import { ApiClient } from '../../api/ApiClient';

class AuthApiService {
    url = 'auth';

    async login(body) {
        return ApiClient.post(`${this.url}/login`, body);
    }
}

export default new AuthApiService();