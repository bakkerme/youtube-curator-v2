import { useQuery, UseQueryResult } from '@tanstack/react-query';

// Types for our API responses
export interface Channel {
  id: string;
  title: string;
  customUrl: string;
  thumbnailUrl: string;
  createdAt: string;
  lastVideoPublishedAt: string;
}

export interface Video {
  id: string;
  title: string;
  channelId: string;
  channelTitle: string;
  thumbnailUrl: string;
  publishedAt: string;
  viewCount: number;
  likeCount: number;  
  commentCount: number;
}

export interface Config {
  host: string;
  port: number;
  username: string;
  password: string;
  fromAddress: string;
  recipientEmail: string;
  emailHour: number;
  emailMinute: number;
  emailTimezone: string;
}

// Hook to fetch channels
export const useChannels = (): UseQueryResult<Channel[], Error> => {
  return useQuery({
    queryKey: ['channels'],
    queryFn: async (): Promise<Channel[]> => {
      const response = await fetch('/api/channels');
      if (!response.ok) {
        throw new Error('Failed to fetch channels');
      }
      return response.json();
    },
  });
};

// Hook to fetch videos with optional refresh
export const useVideos = (options?: { refresh?: boolean }): UseQueryResult<Video[], Error> => {
  return useQuery({
    queryKey: ['videos', options?.refresh],
    queryFn: async (): Promise<Video[]> => {
      const params = new URLSearchParams();
      if (options?.refresh) {
        params.append('refresh', 'true');
      }
      
      const response = await fetch(`/api/videos?${params.toString()}`);
      if (!response.ok) {
        throw new Error('Failed to fetch videos');
      }
      return response.json();
    },
  });
};

// Hook to fetch SMTP configuration
export const useConfig = (): UseQueryResult<Config, Error> => {
  return useQuery({
    queryKey: ['config', 'smtp'],
    queryFn: async (): Promise<Config> => {
      const response = await fetch('/api/config/smtp');
      if (!response.ok) {
        throw new Error('Failed to fetch config');
      }
      return response.json();
    },
  });
}; 