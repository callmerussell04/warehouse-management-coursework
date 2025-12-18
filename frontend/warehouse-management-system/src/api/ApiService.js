import { ApiClient } from './ApiClient';

class ApiService {
    constructor(url) {
        this.url = url;
    }

    async getAll(params = {}) {
        return ApiClient.get(`${this.url}`, { params });
    }

    async create(body) {
        return ApiClient.post(this.url, body);
    }

    async get(id) {
        return ApiClient.get(`${this.url}/${id}`);
    }

    async update(id, body) {
        return ApiClient.put(`${this.url}/${id}`, body)
    }

    async delete(id) {
        return ApiClient.delete(`${this.url}/${id}`);
    }
}

export default ApiService;
