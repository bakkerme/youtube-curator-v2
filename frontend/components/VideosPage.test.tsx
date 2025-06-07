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
      url: 'url1',
      published: today.toISOString(),
      duration: 60,
      thumbnailUrl: 'thumb1',
    },
  },
  { // Video 2 (Yesterday)
    channelId: 'channel1',
    entry: {
      id: 'video2',
      title: 'Video Yesterday Channel One',
      url: 'url2',
      published: yesterday.toISOString(),
      duration: 120,
      thumbnailUrl: 'thumb2',
    },
  },
  { // Video 3 (Specific Past Date)
    channelId: 'channel2',
    entry: {
      id: 'video3',
      title: 'Video Specific Date Channel Two',
      url: 'url3',
      published: specificPastDate.toISOString(),
      duration: 180,
      thumbnailUrl: 'thumb3',
    },
  },
  { // Video 4 (Today, different channel, for search testing)
    channelId: 'channel2',
    entry: {
      id: 'video4',
      title: 'Another Video Today Channel Two',
      url: 'url4',
      published: today.toISOString(),
      duration: 240,
      thumbnailUrl: 'thumb4',
    },
  },
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
    const filterButton = screen.getByRole('button', { name: /Today/i });
    expect(filterButton).toHaveClass('bg-red-600'); // Active state for today
  });

  test('"Per Day" mode filters videos for the selected date', async () => {
    render(<VideosPage />);
    await waitFor(() => expect(screen.queryByText('Loading videos...')).not.toBeInTheDocument());

    // Click filter button to cycle from "Today" to "Per Day"
    const filterButton = screen.getByRole('button', { name: /Today/i });
    fireEvent.click(filterButton);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Per Day/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Per Day/i })).toHaveClass('bg-red-600');
    });

    // Date input should now be visible
    const dateInput = screen.getByRole('textbox'); // type="date" is treated as textbox by testing-library in some setups or if not fully specified
    expect(dateInput).toBeVisible();

    // Change the date to our specificPastDate ("2023-03-15")
    // The date input expects 'YYYY-MM-DD' format
    const specificPastDateString = specificPastDate.toISOString().split('T')[0];
    fireEvent.change(dateInput, { target: { value: specificPastDateString } });

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

    // Click filter button twice: "Today" -> "Per Day" -> "All"
    const filterButtonToday = screen.getByRole('button', { name: /Today/i });
    fireEvent.click(filterButtonToday);

    await waitFor(() => expect(screen.getByRole('button', { name: /Per Day/i })).toBeInTheDocument());
    const filterButtonPerDay = screen.getByRole('button', { name: /Per Day/i });
    fireEvent.click(filterButtonPerDay);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /All Videos/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /All Videos/i })).toHaveClass('bg-blue-600');
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

    const filterButton = screen.getByRole('button', { name: /Today/i });
    let dateInput = screen.queryByRole('textbox'); // Query because it might not exist initially

    // Initial state: Today
    expect(filterButton).toHaveTextContent('Today');
    expect(filterButton).toHaveClass('bg-red-600');
    expect(dateInput).not.toBeInTheDocument(); // Date input hidden

    // Click 1: Today -> Per Day
    fireEvent.click(filterButton);
    await waitFor(() => expect(screen.getByRole('button', { name: /Per Day/i })).toBeInTheDocument());
    const perDayButton = screen.getByRole('button', { name: /Per Day/i });
    dateInput = screen.getByRole('textbox'); // Should be visible now
    expect(perDayButton).toHaveTextContent('Per Day');
    expect(perDayButton).toHaveClass('bg-red-600');
    expect(dateInput).toBeVisible();

    // Click 2: Per Day -> All
    fireEvent.click(perDayButton);
    await waitFor(() => expect(screen.getByRole('button', { name: /All Videos/i })).toBeInTheDocument());
    const allVideosButton = screen.getByRole('button', { name: /All Videos/i });
    dateInput = screen.queryByRole('textbox'); // Query again
    expect(allVideosButton).toHaveTextContent('All Videos');
    expect(allVideosButton).toHaveClass('bg-blue-600');
    expect(dateInput).not.toBeInTheDocument(); // Date input hidden

    // Click 3: All -> Today
    fireEvent.click(allVideosButton);
    await waitFor(() => expect(screen.getByRole('button', { name: /Today/i })).toBeInTheDocument());
    const todayButtonAgain = screen.getByRole('button', { name: /Today/i });
    dateInput = screen.queryByRole('textbox'); // Query again
    expect(todayButtonAgain).toHaveTextContent('Today');
    expect(todayButtonAgain).toHaveClass('bg-red-600');
    expect(dateInput).not.toBeInTheDocument(); // Date input hidden
  });

  describe('Search functionality with date filters', () => {
    const searchCases = [
      { mode: 'today' as const, buttonClicks: 0, initialButtonName: /Today/i, expectedVideo: 'Video Today Channel One', searchFor: 'Channel One' },
      { mode: 'perDay' as const, buttonClicks: 1, initialButtonName: /Today/i, expectedVideo: 'Video Specific Date Channel Two', searchFor: 'Specific Date', dateToSelect: specificPastDate.toISOString().split('T')[0] },
      { mode: 'all' as const, buttonClicks: 2, initialButtonName: /Today/i, expectedVideo: 'Video Yesterday Channel One', searchFor: 'Yesterday' },
    ];

    searchCases.forEach(({ mode, buttonClicks, initialButtonName, expectedVideo, searchFor, dateToSelect }) => {
      test(`works with ${mode} filter`, async () => {
        render(<VideosPage />);
        await waitFor(() => expect(screen.queryByText('Loading videos...')).not.toBeInTheDocument());

        let currentButton = screen.getByRole('button', { name: initialButtonName });
        for (let i = 0; i < buttonClicks; i++) {
          fireEvent.click(currentButton);
          // Wait for the button text to change to ensure mode switch is complete
          await waitFor(() => {
            if (i === 0 && mode === 'perDay') expect(screen.getByRole('button', {name: /Per Day/i})).toBeInTheDocument();
            else if (mode === 'all' && i === 1) expect(screen.getByRole('button', {name: /All Videos/i})).toBeInTheDocument();
          });
          currentButton = screen.getByRole('button', { name: (mode === 'perDay' && i===0) ? /Per Day/i : /All Videos/i });
        }

        if (mode === 'perDay' && dateToSelect) {
          const dateInput = screen.getByRole('textbox');
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
