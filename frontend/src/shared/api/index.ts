import axios from 'axios';

// Get the base API URL from environment variables, fallback to localhost:8080 during development
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true, // Crucial for HTTPOnly JWT Cookie authentication
  headers: {
    'Content-Type': 'application/json',
  },
});

// Response interceptor to handle common errors gracefully
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // Server returned a status code outside the 2xx range
      const status = error.response.status;
      if (status === 401) {
        // Unauthorized - session expired or token missing
        console.warn('Session expired or unauthorized. Redirecting to login...');
        // We can handle redirecting or state resetting here if needed
      }
      return Promise.reject(error.response.data || error.message);
    }
    return Promise.reject(error.message || 'Network error');
  }
);
