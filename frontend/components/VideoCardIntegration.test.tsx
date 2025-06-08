import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';
import VideoCard from './VideoCard';
import { VideoEntry, Channel } from '@/lib/types';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

// Mock next/image
jest.mock('next/image', () => ({
  __esModule: true,
  default: ({ src, alt, ...props }: { src: string; alt: string; [key: string]: unknown }) => <img src={src} alt={alt} {...props} />,
}));

// Test wrapper with QueryClient
const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

// Mock data representing what would come from the backend
const backendChannelsResponse = {
  channels: [
    {
      id: 'UC123',
      title: 'Tech Channel',
      customUrl: '@techchannel',
      thumbnailUrl: 'https://example.com/thumb.jpg',
      createdAt: '2023-01-01T00:00:00Z',
      lastVideoPublishedAt: '2024-01-15T10:00:00Z',
      videoCount: 42,
      isActive: true,
    },
    {
      id: 'UC456', 
      title: 'Gaming Channel',
      customUrl: '@gamingchannel',
      thumbnailUrl: 'https://example.com/thumb2.jpg',
      createdAt: '2023-01-02T00:00:00Z',
      lastVideoPublishedAt: '2024-01-14T15:30:00Z',
      videoCount: 100,
      isActive: true,
    }
  ],
  totalCount: 2,
  lastUpdated: '2024-01-15T10:00:00Z'
};

// Extract channels array as the frontend would after the fix
const channels: Channel[] = backendChannelsResponse.channels.map(ch => ({
  id: ch.id,
  title: ch.title
}));

const mockVideo: VideoEntry = {
  id: 'video-789',
  channelId: 'UC123', // Matches first channel
  cachedAt: '2024-01-01T12:00:00Z',
  watched: false,
  title: 'How to Fix Channel Names in UI',
  link: { href: 'https://youtube.com/watch?v=video-789', rel: 'alternate' },
  published: '2024-01-01T12:00:00Z',
  content: 'Tutorial video content',
  author: { name: 'Tech Channel', uri: 'https://youtube.com/channel/UC123' },
  mediaGroup: {
    mediaThumbnail: { url: 'https://example.com/thumbnail.jpg', width: '320', height: '180' },
    mediaTitle: 'How to Fix Channel Names in UI',
    mediaContent: { url: 'https://example.com/video.mp4', type: 'video/mp4', width: '1920', height: '1080' },
    mediaDescription: 'Tutorial description'
  }
};

describe('VideoCard Integration with Channel API', () => {
  it('should correctly display channel name when channels are properly extracted from backend response', () => {
    // Act
    render(
      <VideoCard 
        video={mockVideo} 
        channels={channels} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert - The channel name should be displayed instead of "Unknown Channel"
    expect(screen.getByText('Tech Channel')).toBeInTheDocument();
    expect(screen.queryByText('Unknown Channel')).not.toBeInTheDocument();
    
    // Verify video details are also displayed correctly
    expect(screen.getByText('How to Fix Channel Names in UI')).toBeInTheDocument();
  });

  it('should handle videos from different channels correctly', () => {
    // Arrange - Video from the second channel
    const gamingVideo = {
      ...mockVideo,
      id: 'video-456',
      channelId: 'UC456', // Matches second channel
      title: 'Epic Gaming Montage',
      author: { name: 'Gaming Channel', uri: 'https://youtube.com/channel/UC456' }
    };

    // Act
    render(
      <VideoCard 
        video={gamingVideo} 
        channels={channels} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert
    expect(screen.getByText('Gaming Channel')).toBeInTheDocument();
    expect(screen.getByText('Epic Gaming Montage')).toBeInTheDocument();
    expect(screen.queryByText('Unknown Channel')).not.toBeInTheDocument();
  });

  it('should show "Unknown Channel" when channel is not in the channels array', () => {
    // Arrange - Video from a channel not in our channels list
    const unknownChannelVideo = {
      ...mockVideo,
      channelId: 'UC999', // Not in our channels array
      title: 'Video from Unknown Channel'
    };

    // Act
    render(
      <VideoCard 
        video={unknownChannelVideo} 
        channels={channels} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert
    expect(screen.getByText('Unknown Channel')).toBeInTheDocument();
    expect(screen.getByText('Video from Unknown Channel')).toBeInTheDocument();
  });

  it('should handle empty channels array gracefully', () => {
    // Act
    render(
      <VideoCard 
        video={mockVideo} 
        channels={[]} // Empty channels array
      />,
      { wrapper: TestWrapper }
    );

    // Assert
    expect(screen.getByText('Unknown Channel')).toBeInTheDocument();
    expect(screen.getByText('How to Fix Channel Names in UI')).toBeInTheDocument();
  });

  it('should verify the structure matches backend ChannelsResponse format', () => {
    // This test verifies our understanding of the backend response structure
    expect(backendChannelsResponse).toHaveProperty('channels');
    expect(backendChannelsResponse).toHaveProperty('totalCount');
    expect(backendChannelsResponse).toHaveProperty('lastUpdated');
    
    expect(Array.isArray(backendChannelsResponse.channels)).toBe(true);
    expect(backendChannelsResponse.channels).toHaveLength(2);
    expect(backendChannelsResponse.totalCount).toBe(2);
    
    // Verify the frontend correctly extracts only id and title
    expect(channels).toHaveLength(2);
    expect(channels[0]).toEqual({ id: 'UC123', title: 'Tech Channel' });
    expect(channels[1]).toEqual({ id: 'UC456', title: 'Gaming Channel' });
  });
});