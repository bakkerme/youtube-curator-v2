import '@testing-library/jest-dom';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import VideoCard from './VideoCard';
import { VideoEntry, Channel } from '@/lib/types';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { server } from '@/lib/mocks/server';

// Mock next/image
jest.mock('next/image', () => ({
  __esModule: true,
  default: ({ src, alt, ...props }: { src: string; alt: string; [key: string]: unknown }) => <img src={src} alt={alt} {...props} />,
}));

// Mock the runtime config
jest.mock('@/lib/config', () => {
  const baseURL = 'http://localhost:8080/api';
  return {
    getRuntimeConfig: jest.fn().mockResolvedValue({
      apiUrl: baseURL,
    })
  };
});

const baseURL = 'http://localhost:8080/api';

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

// Mock data
const mockChannel: Channel = {
  id: 'channel-1',
  title: 'Test Channel'
};

const mockVideoEntry: VideoEntry = {
  id: 'video-123',
  channelId: 'channel-1',
  cachedAt: '2024-01-01T12:00:00Z',
  watched: false,
  title: 'Test Video Title',
  link: { href: 'https://youtube.com/watch?v=test', rel: 'alternate' },
  published: '2024-01-01T12:00:00Z',
  content: 'Test video content',
  author: { name: 'Test Author', uri: 'https://youtube.com/channel/test' },
  mediaGroup: {
    mediaThumbnail: { url: 'https://test.com/thumbnail.jpg', width: '320', height: '180' },
    mediaTitle: 'Test Video',
    mediaContent: { url: 'https://test.com/video.mp4', type: 'video/mp4', width: '1920', height: '1080' },
    mediaDescription: 'Test description'
  }
};

describe('VideoCard', () => {
  it('should render video information correctly', () => {
    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert
    expect(screen.getByText('Test Video Title')).toBeInTheDocument();
    expect(screen.getByText('Test Channel')).toBeInTheDocument();
    expect(screen.getByText('Watch on YouTube')).toBeInTheDocument();
    expect(screen.getByLabelText(/watched/i)).toBeInTheDocument();
  });

  it('should display watched checkbox as unchecked for unwatched video', () => {
    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert
    const checkbox = screen.getByRole('checkbox');
    expect(checkbox).not.toBeChecked();
  });

  it('should display watched checkbox as checked for watched video', () => {
    // Arrange
    const watchedVideo = { ...mockVideoEntry, watched: true };

    // Act
    render(
      <VideoCard 
        video={watchedVideo} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert
    const checkbox = screen.getByRole('checkbox');
    expect(checkbox).toBeChecked();
  });

  it('should call markAsWatched API when checkbox is clicked', async () => {
    // Arrange
    let apiCalled = false;
    server.use(
      http.post(`${baseURL}/videos/video-123/watch`, () => {
        apiCalled = true;
        return new HttpResponse(null, { status: 200 });
      })
    );

    const mockCallback = jest.fn();

    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
        onWatchedStatusChange={mockCallback}
      />,
      { wrapper: TestWrapper }
    );

    const checkbox = screen.getByRole('checkbox');
    fireEvent.click(checkbox);

    // Assert
    await waitFor(() => {
      expect(apiCalled).toBe(true);
      expect(mockCallback).toHaveBeenCalledWith('video-123');
    });
  });

  it('should optimistically update checkbox state', async () => {
    // Arrange
    server.use(
      http.post(`${baseURL}/api/videos/video-123/watch`, async () => {
        // Delay to simulate network request
        await new Promise(resolve => setTimeout(resolve, 100));
        return new HttpResponse(null, { status: 200 });
      })
    );

    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    const checkbox = screen.getByRole('checkbox');
    
    // Initially unchecked
    expect(checkbox).not.toBeChecked();
    
    // Click the checkbox
    fireEvent.click(checkbox);
    
    // Should be immediately checked (optimistic update)
    expect(checkbox).toBeChecked();

    // Wait for API call to complete
    await waitFor(() => {
      expect(checkbox).toBeChecked();
    });
  });

  it('should revert checkbox state on API error', async () => {
    // Arrange
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
    
    server.use(
      http.post(`${baseURL}/videos/video-123/watch`, () => {
        return HttpResponse.json(
          { message: 'Server error' },
          { status: 500 }
        );
      })
    );

    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    const checkbox = screen.getByRole('checkbox');
    
    // Initially unchecked
    expect(checkbox).not.toBeChecked();
    
    // Click the checkbox
    fireEvent.click(checkbox);
    
    // Should be immediately checked (optimistic update)
    expect(checkbox).toBeChecked();

    // Wait for API error and revert
    await waitFor(() => {
      expect(checkbox).not.toBeChecked();
    });

    // Assert
    expect(consoleSpy).toHaveBeenCalledWith('Failed to mark video as watched:', expect.any(Error));
    
    consoleSpy.mockRestore();
  });

  it('should handle unknown channel gracefully', () => {
    // Arrange
    const videoWithUnknownChannel = { ...mockVideoEntry, channelId: 'unknown-channel' };

    // Act
    render(
      <VideoCard 
        video={videoWithUnknownChannel} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert
    expect(screen.getByText('Unknown Channel')).toBeInTheDocument();
  });

  it('should display time ago correctly', () => {
    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert - Should show "about 1 year ago" or similar for 2024-01-01
    expect(screen.getByText(/ago$/)).toBeInTheDocument();
  });

  it('should link to YouTube correctly', () => {
    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert
    const youtubeLink = screen.getByRole('link', { name: /watch on youtube/i });
    expect(youtubeLink).toHaveAttribute('href', 'https://youtube.com/watch?v=test');
    expect(youtubeLink).toHaveAttribute('target', '_blank');
    expect(youtubeLink).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('should make thumbnail clickable and link to YouTube', () => {
    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert
    const thumbnailLink = screen.getByRole('link', { name: /test video title.*thumbnail/i });
    expect(thumbnailLink).toHaveAttribute('href', 'https://youtube.com/watch?v=test');
    expect(thumbnailLink).toHaveAttribute('target', '_blank');
    expect(thumbnailLink).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('should make title clickable and link to YouTube', () => {
    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert
    const titleLink = screen.getByRole('link', { name: /test video title$/i });
    expect(titleLink).toHaveAttribute('href', 'https://youtube.com/watch?v=test');
    expect(titleLink).toHaveAttribute('target', '_blank');
    expect(titleLink).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('should position watched checkbox and buttons correctly', () => {
    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert - Find the buttons and checkbox
    const watchInCuratorButton = screen.getByRole('link', { name: /watch in curator/i });
    const watchOnYouTubeButton = screen.getByRole('link', { name: /watch on youtube/i });
    const watchedCheckbox = screen.getByRole('checkbox');
    
    // Verify both buttons exist
    expect(watchInCuratorButton).toBeInTheDocument();
    expect(watchOnYouTubeButton).toBeInTheDocument();
    expect(watchedCheckbox).toBeInTheDocument();
    
    // Get the container that has both buttons
    const buttonsContainer = watchInCuratorButton.closest('div');
    const youtubeButtonContainer = watchOnYouTubeButton.closest('div');
    
    // Both buttons should be in the same container
    expect(buttonsContainer).toBe(youtubeButtonContainer);
    
    // Verify the buttons container has flex layout classes
    expect(buttonsContainer).toHaveClass('flex', 'space-x-2');
  });

  it('should use flexbox layout to position controls at bottom of card', () => {
    // Act
    render(
      <VideoCard 
        video={mockVideoEntry} 
        channels={[mockChannel]} 
      />,
      { wrapper: TestWrapper }
    );

    // Assert - Find the main content container
    const title = screen.getByText('Test Video Title');
    const contentContainer = title.closest('div[class*="p-4"]');
    expect(contentContainer).not.toBeNull(); // Ensure the container exists
    
    // Verify the content container has flexbox classes for bottom positioning
    expect(contentContainer).toHaveClass('flex', 'flex-col', 'justify-between', 'flex-1');
  });
});