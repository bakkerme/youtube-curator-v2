import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import VideosPage from './VideosPage';
import { videoAPI, channelAPI } from '@/lib/api'; // To be mocked
import { VideoEntry, Channel, VideosAPIResponse } from '@/lib/types';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    prefetch: jest.fn(),
    back: jest.fn(),
    forward: jest.fn(),
  }),
  useSearchParams: () => ({
    get: jest.fn((key) => {
      if (key === 'page') return '1';
      return null;
    }),
  }),
}));

// Mock APIs
jest.mock('@/lib/api', () => ({
  videoAPI: {
    getAll: jest.fn(),
  },
  channelAPI: {
    getAll: jest.fn(),
  },
}));

const mockChannels: Channel[] = [
  { id: 'channel1', title: 'Channel One', url: 'url1', thumbnailUrl: 'thumb1' },
  { id: 'channel2', title: 'Channel Two', url: 'url2', thumbnailUrl: 'thumb2' },
];

const today = new Date();
const yesterday = new Date(today);
yesterday.setDate(today.getDate() - 1);
const specificPastDate = new Date('2023-03-15T12:00:00'); // Ensure specific time for consistency

const mockVideos: VideoEntry[] = [
  { // Video 1 (Today)
    channelId: 'channel1',
    entry: {
      id: 'video1',
      title: 'Video Today Channel One',
      link: { Href: 'https://example.com/video1', Rel: 'alternate' },
      published: today.toISOString(),
      content: 'Content for video 1',
      author: { name: 'Channel One', uri: 'uri_channel1' },
      mediaGroup: {
        mediaThumbnail: { URL: 'https://images.example.com/thumb1.jpg', Width: '120', Height: '90' },
        mediaTitle: 'Video Today Channel One',
        mediaContent: { URL: 'https://videos.example.com/content1.mp4', Type: 'video/mp4', Width: '640', Height: '360' },
        mediaDescription: 'Description for video 1',
      }
      // url, duration, thumbnailUrl are not part of the 'Entry' type directly,
      // but were in previous mock. If VideoCard relies on them at top level of entry,
      // they might need to be mapped or VideoCard adjusted.
      // For now, assuming VideoCard uses mediaGroup.mediaThumbnail.URL for thumbnail
      // and entry.link.Href for the main link.
    },
    cachedAt: today.toISOString(), // Added cachedAt
  },
  { // Video 2 (Yesterday)
    channelId: 'channel1',
    entry: {
      id: 'video2',
      title: 'Video Yesterday Channel One',
      link: { Href: 'https://example.com/video2', Rel: 'alternate' },
      published: yesterday.toISOString(),
      content: 'Content for video 2',
      author: { name: 'Channel One', uri: 'uri_channel1' },
      mediaGroup: {
        mediaThumbnail: { URL: 'https://images.example.com/thumb2.jpg', Width: '120', Height: '90' },
        mediaTitle: 'Video Yesterday Channel One',
        mediaContent: { URL: 'https://videos.example.com/content2.mp4', Type: 'video/mp4', Width: '640', Height: '360' },
        mediaDescription: 'Description for video 2',
      }
    },
    cachedAt: today.toISOString(), // Added cachedAt
  },
  { // Video 3 (Specific Past Date)
    channelId: 'channel2',
    entry: {
      id: 'video3',
      title: 'Video Specific Date Channel Two',
      link: { Href: 'https://example.com/video3', Rel: 'alternate' },
      published: specificPastDate.toISOString(),
      content: 'Content for video 3',
      author: { name: 'Channel Two', uri: 'uri_channel2' },
      mediaGroup: {
        mediaThumbnail: { URL: 'https://images.example.com/thumb3.jpg', Width: '120', Height: '90' },
        mediaTitle: 'Video Specific Date Channel Two',
        mediaContent: { URL: 'https://videos.example.com/content3.mp4', Type: 'video/mp4', Width: '640', Height: '360' },
        mediaDescription: 'Description for video 3',
      }
    },
    cachedAt: today.toISOString(), // Added cachedAt
  },
  { // Video 4 (Today, different channel, for search testing)
    channelId: 'channel2',
    entry: {
      id: 'video4',
      title: 'Another Video Today Channel Two',
      link: { Href: 'https://example.com/video4', Rel: 'alternate' },
      published: today.toISOString(),
      content: 'Content for video 4',
      author: { name: 'Channel Two', uri: 'uri_channel2' },
      mediaGroup: {
        mediaThumbnail: { URL: 'https://images.example.com/thumb4.jpg', Width: '120', Height: '90' },
        mediaTitle: 'Another Video Today Channel Two',
        mediaContent: { URL: 'https://videos.example.com/content4.mp4', Type: 'video/mp4', Width: '640', Height: '360' },
        mediaDescription: 'Description for video 4',
      }
    },
    cachedAt: today.toISOString(), // Added cachedAt
  },
  // Removed extra trailing comma and brace here that caused syntax error
];

const mockVideoAPIResponse: VideosAPIResponse = {
  videos: mockVideos,
  lastRefreshedAt: new Date().toISOString(),
};

describe('VideosPage', () => {
  beforeEach(() => {
    // Reset mocks before each test
    (videoAPI.getAll as jest.Mock).mockResolvedValue(mockVideoAPIResponse);
    (channelAPI.getAll as jest.Mock).mockResolvedValue(mockChannels);
    // Mock console.warn and console.error to avoid cluttering test output
    jest.spyOn(console, 'warn').mockImplementation(jest.fn());
    jest.spyOn(console, 'error').mockImplementation(jest.fn());
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  test('renders and initially filters videos for the current day (Today Mode)', async () => {
    render(<VideosPage />);

    // Wait for loading to finish
    await waitFor(() => expect(screen.queryByText('Loading videos...')).not.toBeInTheDocument());

    // Check that only today's videos are visible
    expect(screen.getByText('Video Today Channel One')).toBeInTheDocument();
    expect(screen.getByText('Another Video Today Channel Two')).toBeInTheDocument();
    expect(screen.queryByText('Video Yesterday Channel One')).not.toBeInTheDocument();
    expect(screen.queryByText('Video Specific Date Channel Two')).not.toBeInTheDocument();

    // Check filter button state
    const filterButton = screen.getByTestId('filter-mode-button');
    expect(filterButton).toHaveTextContent(/Today/i);
    expect(filterButton).toHaveClass('bg-red-600'); // Active state for today
  });

  test('"Per Day" mode filters videos for the selected date', async () => {
    render(<VideosPage />);
    await waitFor(() => expect(screen.queryByText('Loading videos...')).not.toBeInTheDocument());

    // Click filter button to cycle from "Today" to "Per Day"
    const filterButton = screen.getByTestId('filter-mode-button');
    expect(filterButton).toHaveTextContent(/Today/i); // Initial state check
    fireEvent.click(filterButton);

    let dateInput: HTMLElement | null;
    await waitFor(() => {
      expect(filterButton).toHaveTextContent(/Per Day/i);
      expect(filterButton).toHaveClass('bg-red-600');
      dateInput = screen.getByTestId('date-filter-input');
      expect(dateInput).toBeVisible();
    });

    // Date input should now be visible (re-fetch if needed, or use variable from waitFor scope)
    dateInput = screen.getByTestId('date-filter-input');

    // Change the date to our specificPastDate ("2023-03-15")
    // The date input expects 'YYYY-MM-DD' format
    const specificPastDateString = specificPastDate.toISOString().split('T')[0];
    fireEvent.change(dateInput!, { target: { value: specificPastDateString } }); // Added non-null assertion

    // Wait for re-render and filtering
    await waitFor(() => {
      expect(screen.getByText('Video Specific Date Channel Two')).toBeInTheDocument();
    });

    expect(screen.queryByText('Video Today Channel One')).not.toBeInTheDocument();
    expect(screen.queryByText('Another Video Today Channel Two')).not.toBeInTheDocument();
    expect(screen.queryByText('Video Yesterday Channel One')).not.toBeInTheDocument();
  });

  test('"All" mode displays all videos', async () => {
    render(<VideosPage />);
    await waitFor(() => expect(screen.queryByText('Loading videos...')).not.toBeInTheDocument());

    const filterButton = screen.getByTestId('filter-mode-button');
    expect(filterButton).toHaveTextContent(/Today/i);

    // Click filter button twice: "Today" -> "Per Day" -> "All"
    fireEvent.click(filterButton);

    await waitFor(() => expect(filterButton).toHaveTextContent(/Per Day/i));
    fireEvent.click(filterButton);

    await waitFor(() => {
      expect(filterButton).toHaveTextContent(/All Videos/i);
      expect(filterButton).toHaveClass('bg-blue-600');
    });

    // All videos should be visible
    expect(screen.getByText('Video Today Channel One')).toBeInTheDocument();
    expect(screen.getByText('Another Video Today Channel Two')).toBeInTheDocument();
    expect(screen.getByText('Video Yesterday Channel One')).toBeInTheDocument();
    expect(screen.getByText('Video Specific Date Channel Two')).toBeInTheDocument();
  });

  test('filter button cycles modes and date input visibility is correct', async () => {
    render(<VideosPage />);
    await waitFor(() => expect(screen.queryByText('Loading videos...')).not.toBeInTheDocument());

    const filterButton = screen.getByTestId('filter-mode-button');
    let dateInput = screen.queryByTestId('date-filter-input');

    // Initial state: Today
    expect(filterButton).toHaveTextContent(/Today/i);
    expect(filterButton).toHaveClass('bg-red-600');
    expect(dateInput).not.toBeInTheDocument(); // Date input hidden

    // Click 1: Today -> Per Day
    fireEvent.click(filterButton);
    await waitFor(() => {
      expect(filterButton).toHaveTextContent(/Per Day/i);
      expect(filterButton).toHaveClass('bg-red-600');
      dateInput = screen.getByTestId('date-filter-input');
      expect(dateInput).toBeVisible();
    });

    // Click 2: Per Day -> All
    fireEvent.click(filterButton);
    await waitFor(() => {
      expect(filterButton).toHaveTextContent(/All Videos/i);
      expect(filterButton).toHaveClass('bg-blue-600');
      dateInput = screen.queryByTestId('date-filter-input');
      expect(dateInput).not.toBeInTheDocument(); // Date input hidden
    });

    // Click 3: All -> Today
    fireEvent.click(filterButton);
    await waitFor(() => {
      expect(filterButton).toHaveTextContent(/Today/i);
      expect(filterButton).toHaveClass('bg-red-600');
      dateInput = screen.queryByTestId('date-filter-input');
      expect(dateInput).not.toBeInTheDocument(); // Date input hidden
    });
  });

  describe('Search functionality with date filters', () => {
    const searchCases = [
      { mode: 'today' as const, buttonClicks: 0, expectedButtonText: /Today/i, expectedVideo: 'Video Today Channel One', searchFor: 'Channel One' },
      { mode: 'perDay' as const, buttonClicks: 1, expectedButtonText: /Per Day/i, expectedVideo: 'Video Specific Date Channel Two', searchFor: 'Specific Date', dateToSelect: specificPastDate.toISOString().split('T')[0] },
      { mode: 'all' as const, buttonClicks: 2, expectedButtonText: /All Videos/i, expectedVideo: 'Video Yesterday Channel One', searchFor: 'Yesterday' },
    ];

    searchCases.forEach(({ mode, buttonClicks, expectedButtonText, expectedVideo, searchFor, dateToSelect }) => {
      test(`works with ${mode} filter`, async () => {
        render(<VideosPage />);
        await waitFor(() => expect(screen.queryByText('Loading videos...')).not.toBeInTheDocument());

        const filterButton = screen.getByTestId('filter-mode-button');
        for (let i = 0; i < buttonClicks; i++) {
          fireEvent.click(filterButton);
          // Wait for the button text to change to ensure mode switch is complete
          // This logic could be simplified if we always check against filterButton.toHaveTextContent for the *next* expected state.
          await waitFor(() => {
            if (i === 0 && mode === 'perDay') expect(filterButton).toHaveTextContent(/Per Day/i);
            else if (mode === 'all' && i === 1) expect(filterButton).toHaveTextContent(/All Videos/i);
            // Add other cases if buttonClicks > 2, or if initial state isn't "Today"
          });
        }
        // Final check for button text after all clicks
        await waitFor(() => expect(filterButton).toHaveTextContent(expectedButtonText));

        if (mode === 'perDay' && dateToSelect) {
          const dateInput = screen.getByTestId('date-filter-input');
          fireEvent.change(dateInput, { target: { value: dateToSelect } });
          // Wait for videos to filter based on new date
          await waitFor(() => expect(screen.queryByText(mockVideos[0].entry.title)).not.toBeInTheDocument());
        }

        const searchInput = screen.getByPlaceholderText(/Search videos or channels.../i);
        fireEvent.change(searchInput, { target: { value: searchFor } });

        await waitFor(() => {
          expect(screen.getByText(expectedVideo)).toBeInTheDocument();
        });

        // Check that other videos that don't match search are not present
        mockVideos.filter(v => v.entry.title !== expectedVideo && !v.entry.title.includes(searchFor)).forEach(v => {
             // If the video was supposed to be filtered out by date already, it won't be there.
             // This check is more about search filtering out other videos that *would* match the date filter.
             if (mode === 'today') {
                const videoDate = new Date(v.entry.published).toISOString().split('T')[0];
                const todayString = today.toISOString().split('T')[0];
                if(videoDate === todayString) { // If it's a "today" video but doesn't match search
                    expect(screen.queryByText(v.entry.title)).not.toBeInTheDocument();
                }
             } else if (mode === 'perDay' && dateToSelect) {
                const videoDate = new Date(v.entry.published).toISOString().split('T')[0];
                 if(videoDate === dateToSelect) { // If it's a "selectedDate" video but doesn't match search
                    expect(screen.queryByText(v.entry.title)).not.toBeInTheDocument();
                 }
             } else if (mode === 'all') { // In 'all' mode, if it doesn't match search, it shouldn't be there
                 expect(screen.queryByText(v.entry.title)).not.toBeInTheDocument();
             }
        });
      });
    });
  });
});
