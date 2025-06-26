import axios from "axios";
import config from "@/config";

const api = axios.create({
  baseURL: config.apiUrl,
  // You can add headers or interceptors here if needed
});

export default api;

// POST request
export const post = (url, data, config) => api.post(url, data, config);
