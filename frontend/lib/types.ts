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

export interface SMTPConfigRequest {
  server: string;
  port: string;
  username: string;
  password: string;
  recipientEmail: string;
}

export interface SMTPConfigResponse {
  server: string;
  port: string;
  username: string;
  recipientEmail: string;
  passwordSet: boolean;
}

export interface LLMConfigRequest {
  endpoint: string;
  apiKey: string;
  model: string;
}

export interface LLMConfigResponse {
  endpoint: string;
  model: string;
  apiKeySet: boolean;
}

export interface ApiError {
  message: string;
}

export interface RunNewsletterRequest {
  channelId?: string;
  ignoreLastChecked?: boolean;
  maxItems?: number;
}

export interface RunNewsletterResponse {
  message: string;
  channelsProcessed: number;
  channelsWithError: number;
  newVideosFound: number;
  emailSent: boolean;
}

// RSS Entry types
export interface MediaThumbnail {
  url: string;
  width: string;
  height: string;
}

export interface MediaContent {
  url: string;
  type: string;
  width: string;
  height: string;
}

export interface MediaGroup {
  mediaThumbnail: MediaThumbnail;
  mediaTitle: string;
  mediaContent: MediaContent;
  mediaDescription: string;
}

export interface Author {
  name: string;
  uri: string;
}

export interface Link {
  href: string;
  rel: string;
}

// Video Entry from the API
export interface VideoEntry {
  id: string;
  channelId: string;
  cachedAt: string;
  watched: boolean;
  title: string;
  link: Link;
  published: string;
  content: string;
  author: Author;
  mediaGroup: MediaGroup;
}

export interface VideosAPIResponse {
  videos: VideoEntry[];
  lastRefresh: string;
  totalCount: number;
}