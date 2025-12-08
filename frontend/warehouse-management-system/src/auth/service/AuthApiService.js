import { ApiClient } from '../../api/ApiClient';

class AuthApiService {
    url = 'auth';

    async login(body) {
        return ApiClient.post(`${this.url}/login`, body);
    }

    async logout() {
        return ApiClient.post(`${this.url}/logout`);
    }
}

export default new AuthApiService();