export interface Channel {
  id: string;
  title: string;
}

export interface ChannelRequest {
  url: string;
  title?: string;
}

export interface ConfigInterval {
  interval: string;
}

export interface ApiError {
  message: string;
} 