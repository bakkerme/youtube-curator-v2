import '@testing-library/jest-dom';
import { render, screen, waitFor } from '@testing-library/react';
import WatchPage from './page';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { videoAPI, channelAPI } from '@/lib/api';
import { useParams } from 'next/navigation';

// Mock the runtime config
jest.mock('@/lib/config', () => {
  return {
    getRuntimeConfig: jest.fn().mockResolvedValue({
      apiUrl: 'http://localhost:8080/api',
    })
  };
});

// Mock the Next.js useParams hook
jest.mock('next/navigation', () => ({
  useParams: jest.fn(),
}));

// Mock the API modules
jest.mock('@/lib/api', () => ({
  videoAPI: {
    getAll: jest.fn(),
  },
  channelAPI: {
    getAll: jest.fn(),
  },
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

describe('WatchPage', () => {
  const mockVideo = {
    id: 'yt:video:test-video-id',
    title: 'Test Video Title',
    channelId: 'channel1',
    watched: false,
    link: { href: 'https://youtube.com/watch?v=test-video-id', rel: 'alternate' },
    published: new Date().toISOString(),
    content: 'Test video description',
    author: { name: 'Test Channel', uri: 'uri_channel1' },
    mediaGroup: {
      mediaThumbnail: { url: 'https://example.com/thumb.jpg', width: '120', height: '90' },
      mediaTitle: 'Test Video Title',
      mediaContent: { url: 'https://example.com/video.mp4', type: 'video/mp4', width: '640', height: '360' },
      mediaDescription: 'Test video description',
    },
    cachedAt: new Date().toISOString(),
  };

  const mockChannel = {
    id: 'channel1',
    title: 'Test Channel',
    link: { href: 'https://youtube.com/channel/channel1', rel: 'alternate' },
    published: new Date().toISOString(),
    updated: new Date().toISOString(),
    author: { name: 'Test Channel', uri: 'uri_channel1' },
    thumbnail: 'https://example.com/channel-thumb.jpg',
    description: 'Test channel description',
  };

  beforeEach(() => {
    // Clear all mocks before each test
    jest.clearAllMocks();
    
    // Set initial document title
    document.title = 'Curator';
    
    // Mock useParams to return the test video ID
    (useParams as jest.Mock).mockReturnValue({
      videoId: 'test-video-id'
    });
  });

  afterEach(() => {
    // Reset document title after each test
    document.title = 'Curator';
  });

  it('updates window title with video title when video loads', async () => {
    // Mock API responses
    (videoAPI.getAll as jest.Mock).mockResolvedValue({
      videos: [mockVideo],
      lastRefreshedAt: new Date().toISOString(),
    });
    (channelAPI.getAll as jest.Mock).mockResolvedValue([mockChannel]);

    render(
      <TestWrapper>
        <WatchPage />
      </TestWrapper>
    );

    // Wait for the video to load and title to update
    await waitFor(() => {
      expect(document.title).toBe('Test Video Title - Curator');
    });

    // Verify the video content is displayed
    expect(screen.getByText('Test Video Title')).toBeInTheDocument();
  });

  it('displays loading state without changing title', () => {
    // Mock API responses to return pending promises (loading state)
    (videoAPI.getAll as jest.Mock).mockReturnValue(new Promise(() => {}));
    (channelAPI.getAll as jest.Mock).mockReturnValue(new Promise(() => {}));

    render(
      <TestWrapper>
        <WatchPage />
      </TestWrapper>
    );

    // Title should remain unchanged during loading
    expect(document.title).toBe('Curator');
    
    // Verify we have the loading skeleton structure
    const loadingContainer = document.querySelector('.animate-pulse');
    expect(loadingContainer).toBeInTheDocument();
  });

  it('displays not found state without changing title when video does not exist', async () => {
    // Mock API responses with empty video list
    (videoAPI.getAll as jest.Mock).mockResolvedValue({
      videos: [],
      lastRefreshedAt: new Date().toISOString(),
    });
    (channelAPI.getAll as jest.Mock).mockResolvedValue([]);

    render(
      <TestWrapper>
        <WatchPage />
      </TestWrapper>
    );

    // Wait for the API calls to complete
    await waitFor(() => {
      expect(screen.getByText('Video Not Found')).toBeInTheDocument();
    });

    // Title should remain unchanged when video is not found
    expect(document.title).toBe('Curator');
  });

  it('should render video player with 16:9 aspect ratio', async () => {
    // Mock API responses
    (videoAPI.getAll as jest.Mock).mockResolvedValue({
      videos: [mockVideo],
      lastRefreshedAt: new Date().toISOString(),
    });
    (channelAPI.getAll as jest.Mock).mockResolvedValue([mockChannel]);

    render(
      <TestWrapper>
        <WatchPage />
      </TestWrapper>
    );

    // Wait for the video to load
    await screen.findByTitle('Test Video Title');

    // Check that the video player container has the correct aspect ratio class
    const videoContainer = screen.getByTitle('Test Video Title').closest('.aspect-video');
    expect(videoContainer).toBeInTheDocument();
    expect(videoContainer).toHaveClass('aspect-video');
  });
});