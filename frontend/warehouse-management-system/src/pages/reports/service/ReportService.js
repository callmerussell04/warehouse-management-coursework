import ApiService from '../../../api/ApiService';
import { ApiClient } from '../../../api/ApiClient';

class ReportService extends ApiService {
    constructor() {
        super('api/reports');
    }

    async downloadTurnoverReport(from, to) {
        return ApiClient.get(`${this.url}/turnover`, {
            params: { from, to },
            responseType: 'blob' 
        });
    }
}

export default new ReportService();