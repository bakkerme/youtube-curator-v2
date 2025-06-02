export interface Channel {
  id: string;
  title: string;
}

export interface ChannelRequest {
  url: string;
  title?: string;
}

export interface ChannelImport {
  url: string;
  title?: string;
}

export interface ImportChannelsRequest {
  channels: ChannelImport[];
}

export interface ImportFailure {
  channel: ChannelImport;
  error: string;
}

export interface ImportChannelsResponse {
  imported: Channel[];
  failed: ImportFailure[];
}

export interface ConfigInterval {
  interval: string;
}

export interface ApiError {
  message: string;
}

export interface RunNewsletterRequest {
  channelId?: string;
}

export interface RunNewsletterResponse {
  message: string;
  channelsProcessed: number;
  channelsWithError: number;
  newVideosFound: number;
  emailSent: boolean;
} 