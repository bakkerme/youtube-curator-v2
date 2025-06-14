import axios from 'axios';
import { Channel, ChannelRequest, ConfigInterval, ImportChannelsRequest, ImportChannelsResponse, LLMConfigRequest, LLMConfigResponse, RunNewsletterRequest, RunNewsletterResponse, SMTPConfigRequest, SMTPConfigResponse, VideosAPIResponse, VideoSummaryResponse } from './types';
import { getRuntimeConfig } from './config';

// Create axios instance that will be configured with runtime config
const api = axios.create({
  headers: {
    'Content-Type': 'application/json',
  },
});

// Flag to track if API has been initialized
let isInitialized = false;

// Initialize API client with runtime configuration
async function initializeAPI() {
  if (isInitialized) return;
  
  try {
    const config = await getRuntimeConfig();
    api.defaults.baseURL = config.apiUrl;
    isInitialized = true;
  } catch (error) {
    console.error('Failed to initialize API client:', error);
    // Fallback to default URL if config fails
    api.defaults.baseURL = 'http://localhost:8080/api';
    isInitialized = true;
  }
}

// Wrapper function to ensure API is initialized before making requests
async function makeRequest<T>(requestFn: () => Promise<T>): Promise<T> {
  await initializeAPI();
  return requestFn();
}

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
    return makeRequest(async () => {
      const { data } = await api.get('/channels');
      return data.channels || [];
    });
  },

  add: async (request: ChannelRequest): Promise<Channel> => {
    return makeRequest(async () => {
      const { data } = await api.post('/channels', request);
      return data;
    });
  },

  remove: async (channelId: string): Promise<void> => {
    return makeRequest(async () => {
      await api.delete(`/channels/${channelId}`);
    });
  },

  import: async (request: ImportChannelsRequest): Promise<ImportChannelsResponse> => {
    return makeRequest(async () => {
      const { data } = await api.post('/channels/import', request);
      return data;
    });
  },
};

// Configuration APIs
export const configAPI = {
  getInterval: async (): Promise<ConfigInterval> => {
    return makeRequest(async () => {
      const { data } = await api.get('/config/interval');
      return data;
    });
  },

  setInterval: async (interval: string): Promise<ConfigInterval> => {
    return makeRequest(async () => {
      const { data } = await api.put('/config/interval', { interval });
      return data;
    });
  },

  getSMTP: async (): Promise<SMTPConfigResponse> => {
    return makeRequest(async () => {
      const { data } = await api.get('/config/smtp');
      return data;
    });
  },

  setSMTP: async (config: SMTPConfigRequest): Promise<SMTPConfigResponse> => {
    return makeRequest(async () => {
      const { data } = await api.put('/config/smtp', config);
      return data;
    });
  },

  getLLM: async (): Promise<LLMConfigResponse> => {
    return makeRequest(async () => {
      const { data } = await api.get('/config/llm');
      return data;
    });
  },

  setLLM: async (config: LLMConfigRequest): Promise<LLMConfigResponse> => {
    return makeRequest(async () => {
      const { data } = await api.put('/config/llm', config);
      return data;
    });
  },
};

// Newsletter APIs
export const newsletterAPI = {
  run: async (request?: RunNewsletterRequest): Promise<RunNewsletterResponse> => {
    return makeRequest(async () => {
      const { data } = await api.post('/newsletter/run', request || {});
      return data;
    });
  },
};

// Helper function to extract raw video ID from full format
// Centralized conversion utility for consistent video ID handling
function extractRawVideoId(fullVideoId: string): string {
  const prefix = 'yt:video:';
  if (fullVideoId.startsWith(prefix)) {
    return fullVideoId.substring(prefix.length);
  }
  // If it's already a raw ID, return as-is
  return fullVideoId;
}

// Helper function to extract video ID from YouTube URLs or return ID as-is
export function extractVideoId(input: string): string {
  // Remove whitespace
  input = input.trim();
  
  // If it's already a video ID (11 characters, alphanumeric with - and _)
  if (/^[a-zA-Z0-9_-]{11}$/.test(input)) {
    return input;
  }
  
  // YouTube URL patterns
  const patterns = [
    /(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/|youtube\.com\/v\/|youtube\.com\/shorts\/)([a-zA-Z0-9_-]{11})/,
    /youtube\.com\/.*[?&]v=([a-zA-Z0-9_-]{11})/
  ];
  
  for (const pattern of patterns) {
    const match = input.match(pattern);
    if (match) {
      return match[1];
    }
  }
  
  // If no pattern matches, return the input as-is (might still be a valid ID)
  return input;
}

// Video APIs
export const videoAPI = {
  getAll: async (refresh?: boolean): Promise<VideosAPIResponse> => {
    return makeRequest(async () => {
      const url = refresh ? '/videos?refresh=true' : '/videos';
      const { data } = await api.get<VideosAPIResponse>(url);
      return data;
    });
  },

  markAsWatched: async (videoId: string): Promise<void> => {
    return makeRequest(async () => {
      const rawVideoId = extractRawVideoId(videoId);
      await api.post(`/videos/${rawVideoId}/watch`);
    });
  },

  getSummary: async (videoId: string): Promise<VideoSummaryResponse> => {
    return makeRequest(async () => {
      const rawVideoId = extractRawVideoId(videoId);
      const { data } = await api.get<VideoSummaryResponse>(`/videos/${rawVideoId}/summary`);
      return data;
    });
  },
};

export default api;