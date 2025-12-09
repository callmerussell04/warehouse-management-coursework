import ApiService from '../../../api/ApiService';
import { ApiClient } from '../../../api/ApiClient';

class OrderService extends ApiService {
    constructor() {
        super('api/orders');
    }

    async updateStatus(id, status) {
        return ApiClient.put(`${this.url}/${id}`, { status });
    }
}

export default new OrderService();