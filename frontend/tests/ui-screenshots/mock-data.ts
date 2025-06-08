import { VideoEntry, Channel, SMTPConfigResponse } from '@/lib/types';

export const mockVideoEntries: VideoEntry[] = [
  {
    id: 'dQw4w9WgXcQ',
    channelId: 'UC1',
    cachedAt: '2025-01-07T10:30:00Z',
    watched: false,
    title: 'Advanced React Patterns and Best Practices',
    link: { href: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ', rel: 'alternate' },
    published: '2025-01-07T10:30:00Z',
    content: 'Learn advanced React patterns that will make your code more maintainable and scalable.',
    author: { name: 'React Mastery', uri: 'https://www.youtube.com/channel/UC1' },
    mediaGroup: {
      mediaThumbnail: { url: 'https://via.placeholder.com/320x180/ff0000/ffffff?text=React+Tutorial', width: '320', height: '180' },
      mediaTitle: 'Advanced React Patterns and Best Practices',
      mediaContent: { url: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ', type: 'application/x-shockwave-flash', width: '480', height: '360' },
      mediaDescription: 'Learn advanced React patterns that will make your code more maintainable and scalable.',
    },
  },
  {
    id: 'abc123def45',
    channelId: 'UC2',
    cachedAt: '2025-01-06T14:15:00Z',
    watched: false,
    title: 'Node.js Performance Optimization Techniques',
    link: { href: 'https://www.youtube.com/watch?v=abc123def45', rel: 'alternate' },
    published: '2025-01-06T14:15:00Z',
    content: 'Discover proven techniques to optimize your Node.js applications for maximum performance.',
    author: { name: 'Backend Developer Hub', uri: 'https://www.youtube.com/channel/UC2' },
    mediaGroup: {
      mediaThumbnail: { url: 'https://via.placeholder.com/320x180/00ff00/000000?text=Node.js+Tips', width: '320', height: '180' },
      mediaTitle: 'Node.js Performance Optimization Techniques',
      mediaContent: { url: 'https://www.youtube.com/watch?v=abc123def45', type: 'application/x-shockwave-flash', width: '480', height: '360' },
      mediaDescription: 'Discover proven techniques to optimize your Node.js applications for maximum performance.',
    },
  },
  {
    id: 'xyz789ghi01',
    channelId: 'UC3',
    cachedAt: '2025-01-05T09:45:00Z',
    watched: true,
    title: 'Building Scalable Microservices with Docker',
    link: { href: 'https://www.youtube.com/watch?v=xyz789ghi01', rel: 'alternate' },
    published: '2025-01-05T09:45:00Z',
    content: 'Learn how to build and deploy scalable microservices using Docker containers.',
    author: { name: 'DevOps Central', uri: 'https://www.youtube.com/channel/UC3' },
    mediaGroup: {
      mediaThumbnail: { url: 'https://via.placeholder.com/320x180/0000ff/ffffff?text=Docker+Guide', width: '320', height: '180' },
      mediaTitle: 'Building Scalable Microservices with Docker',
      mediaContent: { url: 'https://www.youtube.com/watch?v=xyz789ghi01', type: 'application/x-shockwave-flash', width: '480', height: '360' },
      mediaDescription: 'Learn how to build and deploy scalable microservices using Docker containers.',
    },
  },
  {
    id: 'mno456pqr78',
    channelId: 'UC4',
    cachedAt: '2025-01-04T16:20:00Z',
    watched: false,
    title: 'CSS Grid Layout: Complete Tutorial for Beginners',
    link: { href: 'https://www.youtube.com/watch?v=mno456pqr78', rel: 'alternate' },
    published: '2025-01-04T16:20:00Z',
    content: 'Master CSS Grid layout with this comprehensive tutorial for beginners.',
    author: { name: 'Frontend Focus', uri: 'https://www.youtube.com/channel/UC4' },
    mediaGroup: {
      mediaThumbnail: { url: 'https://via.placeholder.com/320x180/ffaa00/000000?text=CSS+Grid', width: '320', height: '180' },
      mediaTitle: 'CSS Grid Layout: Complete Tutorial for Beginners',
      mediaContent: { url: 'https://www.youtube.com/watch?v=mno456pqr78', type: 'application/x-shockwave-flash', width: '480', height: '360' },
      mediaDescription: 'Master CSS Grid layout with this comprehensive tutorial for beginners.',
    },
  },
  {
    id: 'stu901vwx23',
    channelId: 'UC5',
    cachedAt: '2025-01-03T11:10:00Z',
    watched: false,
    title: 'Database Design Principles Every Developer Should Know',
    link: { href: 'https://www.youtube.com/watch?v=stu901vwx23', rel: 'alternate' },
    published: '2025-01-03T11:10:00Z',
    content: 'Essential database design principles that every developer should understand.',
    author: { name: 'Database Expert', uri: 'https://www.youtube.com/channel/UC5' },
    mediaGroup: {
      mediaThumbnail: { url: 'https://via.placeholder.com/320x180/aa00ff/ffffff?text=Database+Design', width: '320', height: '180' },
      mediaTitle: 'Database Design Principles Every Developer Should Know',
      mediaContent: { url: 'https://www.youtube.com/watch?v=stu901vwx23', type: 'application/x-shockwave-flash', width: '480', height: '360' },
      mediaDescription: 'Essential database design principles that every developer should understand.',
    },
  },
];

export const mockChannels: Channel[] = [
  {
    id: 'UC1',
    title: 'React Mastery',
  },
  {
    id: 'UC2', 
    title: 'Backend Developer Hub',
  },
  {
    id: 'UC3',
    title: 'DevOps Central',
  },
  {
    id: 'UC4',
    title: 'Frontend Focus',
  },
  {
    id: 'UC5',
    title: 'Database Expert',
  },
];

export const mockSMTPConfig: SMTPConfigResponse = {
  server: 'smtp.gmail.com',
  port: '587',
  username: 'curator@example.com',
  recipientEmail: 'team@example.com',
  passwordSet: true,
};

export const emptyMockData = {
  videos: [],
  channels: [],
  smtpConfig: {
    server: '',
    port: '',
    username: '',
    recipientEmail: '',
    passwordSet: false,
  },
};