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
  URL: string;
  Width: string;
  Height: string;
}

export interface MediaContent {
  URL: string;
  Type: string;
  Width: string;
  Height: string;
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
  Href: string;
  Rel: string;
}

export interface Entry {
  title: string;
  link: Link;
  id: string;
  published: string;
  content: string;
  author: Author;
  mediaGroup: MediaGroup;
}

// Video Entry from the API
export interface VideoEntry {
  entry: Entry;
  channelId: string;
  cachedAt: string;
  watched: boolean;
}

export interface VideosAPIResponse {
  videos: VideoEntry[];
  lastRefreshedAt: string;
}