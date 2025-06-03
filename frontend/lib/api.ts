import axios from 'axios';
import getConfig from 'next/config';
import { Channel, ChannelRequest, ConfigInterval, ImportChannelsRequest, ImportChannelsResponse, RunNewsletterRequest, RunNewsletterResponse } from './types';

// Get runtime configuration
const { publicRuntimeConfig } = getConfig();

// Configure axios with base URL from runtime config
const API_BASE_URL = publicRuntimeConfig?.apiUrl || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // The request was made and the server responded with a status code
      // that falls out of the range of 2xx
      const message = error.response.data?.message || error.message;
      error.message = message;
    } else if (error.request) {
      // The request was made but no response was received
      error.message = 'Unable to connect to the server. Please check if the backend is running.';
    } else {
      // Something happened in setting up the request that triggered an Error
      error.message = 'An unexpected error occurred';
    }
    return Promise.reject(error);
  }
);

// Channel APIs
export const channelAPI = {
  getAll: async (): Promise<Channel[]> => {
    const { data } = await api.get('/channels');
    return data;
  },

  add: async (request: ChannelRequest): Promise<Channel> => {
    const { data } = await api.post('/channels', request);
    return data;
  },

  remove: async (channelId: string): Promise<void> => {
    await api.delete(`/channels/${channelId}`);
  },

  import: async (request: ImportChannelsRequest): Promise<ImportChannelsResponse> => {
    const { data } = await api.post('/channels/import', request);
    return data;
  },
};

// Configuration APIs
export const configAPI = {
  getInterval: async (): Promise<ConfigInterval> => {
    const { data } = await api.get('/config/interval');
    return data;
  },

  setInterval: async (interval: string): Promise<ConfigInterval> => {
    const { data } = await api.put('/config/interval', { interval });
    return data;
  },
};

// Newsletter APIs
export const newsletterAPI = {
  run: async (request?: RunNewsletterRequest): Promise<RunNewsletterResponse> => {
    const { data } = await api.post('/newsletter/run', request || {});
    return data;
  },
};

export default api; 