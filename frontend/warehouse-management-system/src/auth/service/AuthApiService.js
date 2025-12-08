import { ApiClient } from '../../api/ApiClient';

class AuthApiService {
    url = 'auth';

    async login(body) {
        return ApiClient.post(`${this.url}/login`, body);
    }

    async logout() {
        return ApiClient.post(`${this.url}/logout`);
    }

    async requestOTP(body) {
        return ApiClient.post(`${this.url}/request-otp`, body)
    }

    async resetPassword(body) {
        return ApiClient.post(`${this.url}/reset-password`, body)
    }

    async forgotUsername(body) {
        return ApiClient.post(`${this.url}/forgot-username`, body)
    }
}

export default new AuthApiService();