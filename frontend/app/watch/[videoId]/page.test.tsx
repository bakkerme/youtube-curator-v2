import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import WatchPage from './page';
import { videoAPI, channelAPI } from '@/lib/api';

// Mock the APIs
jest.mock('@/lib/api', () => ({
  videoAPI: {
    getAll: jest.fn(),
  },
  channelAPI: {
    getAll: jest.fn(),
  },
}));

// Mock Next.js router
const mockUseParams = jest.fn();
jest.mock('next/navigation', () => ({
  useParams: () => mockUseParams(),
}));

const mockVideoAPI = videoAPI as jest.Mocked<typeof videoAPI>;
const mockChannelAPI = channelAPI as jest.Mocked<typeof channelAPI>;

describe('WatchPage', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });
    jest.clearAllMocks();
  });

  const renderWithQueryClient = (component: React.ReactElement) => {
    return render(
      <QueryClientProvider client={queryClient}>
        {component}
      </QueryClientProvider>
    );
  };

  it('should render video player with 16:9 aspect ratio', async () => {
    // Mock the router params
    mockUseParams.mockReturnValue({ videoId: 'test-video-id' });

    // Mock API responses
    mockVideoAPI.getAll.mockResolvedValue({
      videos: [
        {
          id: 'yt:video:test-video-id',
          channelId: 'test-channel-id',
          cachedAt: '2023-01-01T00:00:00Z',
          watched: false,
          title: 'Test Video Title',
          link: { href: 'https://youtube.com/watch?v=test-video-id', rel: 'alternate' },
          published: '2023-01-01T00:00:00Z',
          content: 'Test video description',
          author: { name: 'Test Author', uri: 'https://youtube.com/channel/test' },
          mediaGroup: {
            mediaThumbnail: { url: 'https://test.com/thumbnail.jpg', width: '320', height: '180' },
            mediaTitle: 'Test Video',
            mediaContent: { url: 'https://test.com/video.mp4', type: 'video/mp4', width: '1920', height: '1080' },
            mediaDescription: 'Test description',
          },
        },
      ],
    });

    mockChannelAPI.getAll.mockResolvedValue([
      {
        id: 'test-channel-id',
        title: 'Test Channel',
        description: 'Test channel description',
        link: { href: 'https://youtube.com/channel/test', rel: 'alternate' },
        image: { url: 'https://test.com/channel.jpg', title: 'Test Channel', link: 'https://youtube.com/channel/test' },
        feedUrl: 'https://youtube.com/feeds/videos.xml?channel_id=test',
        lastBuildDate: '2023-01-01T00:00:00Z',
      },
    ]);

    renderWithQueryClient(<WatchPage />);

    // Wait for the video to load
    await screen.findByTitle('Test Video Title');

    // Check that the video player container has the correct aspect ratio class
    const videoContainer = screen.getByTitle('Test Video Title').closest('.aspect-video');
    expect(videoContainer).toBeInTheDocument();
    expect(videoContainer).toHaveClass('aspect-video');
  });
});