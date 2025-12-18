import ApiService from '../../../api/ApiService';
import { ApiClient } from '../../../api/ApiClient';

class ProductService extends ApiService {
    constructor() {
        super('api/products');
    }

    async getHistory(id, params = {}) {
        return ApiClient.get(`${this.url}/${id}/history`, { params });
    }
}

export default new ProductService();