import axios from 'axios';
import toast from 'react-hot-toast';

export class HttpError extends Error {
    constructor(message = '') {
        super(message);
        this.name = 'HttpError';
        Object.setPrototypeOf(this, new.target.prototype);
        toast.error(message, { id: 'HttpError' });
    }
}

let accessToken = null;

export const TokenService = {
    setAccessToken: (token) => {
        accessToken = token;
    },
    getAccessToken: () => accessToken,
    clear: () => {
        accessToken = null;
        localStorage.removeItem('user_data');
    },
};

function responseHandler(response) {
    if (response.status >= 200 && response.status < 300) {
        if (response.status === 204) {
            return null;
        }

        const contentType = response.headers?.['content-type']?.toLowerCase() || '';
        const expectsBlob = response.config?.responseType === 'blob';
        if (!expectsBlob && contentType.includes('text/html')) {
            throw new HttpError('API Error. Unexpected HTML response!');
        }

        const data = response?.data;
        if (!data) {
            throw new HttpError('API Error. No data!');
        }
        
        return data;
    }
    throw new HttpError(`API Error! Invalid status code ${response.status}!`);
}

function responseErrorHandler(error) {
    if (error === null) {
        throw new Error('Unrecoverable error!! Error is null!');
    }
    toast.error(error.message, { id: 'AxiosError' });
    return Promise.reject(error.message);
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL?.replace(/\/+$/, '') || window.location.origin;

export const ApiClient = axios.create({
    baseURL: API_BASE_URL,
    timeout: 10000,
    withCredentials: true,
    headers: {
        Accept: 'application/json',
    },
});

let isRefreshing = false;
let refreshSubscribers = [];

function onRefreshed(newToken) {
    refreshSubscribers.forEach((callback) => callback(newToken));
    refreshSubscribers = [];
}

function addRefreshSubscriber(callback) {
    refreshSubscribers.push(callback);
}

ApiClient.interceptors.request.use(
    (config) => {
        const token = TokenService.getAccessToken();
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => Promise.reject(error)
);

ApiClient.interceptors.response.use(
    responseHandler,
    async (error) => {
        const originalRequest = error.config ?? {};

        if (
            error.response?.status === 401 &&
            !originalRequest._retry &&
            !originalRequest.skipAuthRefresh
        ) {
            originalRequest._retry = true;

            if (isRefreshing) {
                return new Promise((resolve) => {
                    addRefreshSubscriber((newToken) => {
                        originalRequest.headers.Authorization = `Bearer ${newToken}`;
                        resolve(ApiClient(originalRequest));
                    });
                });
            }

            isRefreshing = true;

            try {
                const refreshResponse = await axios.post(
                    `${API_BASE_URL}/auth/refresh`,
                    {},
                    { withCredentials: true }
                );

                const newAccessToken = refreshResponse.data?.access_token;
                if (!newAccessToken) throw new Error('No access token in refresh response');

                TokenService.setAccessToken(newAccessToken);
                onRefreshed(newAccessToken);

                originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
                return ApiClient(originalRequest);
            } catch (refreshError) {
                TokenService.clear();
                toast.error('Сессия истекла. Войдите снова.', { id: 'SessionExpired' });
                window.location.href = '/login';
                return Promise.reject(refreshError);
            } finally {
                isRefreshing = false;
            }
        }

        if (originalRequest.suppressErrorToast) {
            return Promise.reject(error);
        }

        return responseErrorHandler(error);
    }
);
