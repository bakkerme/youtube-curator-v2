import { http, HttpResponse } from 'msw';
import { QueryClient } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { createWrapper } from '../mocks/QueryWrapper';
import { useChannels, useVideos, useConfig } from './react-query';
import { server } from '../mocks/server';

describe('React Query Hooks Integration Tests', () => {
  let queryClient: QueryClient;
  let wrapper: React.ComponentType<{ children: React.ReactNode }>;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          gcTime: 0, // Renamed from cacheTime in React Query v5
          staleTime: 0,
        },
        mutations: {
          retry: false,
        },
      },
    });
    wrapper = createWrapper(queryClient);
  });

  afterEach(() => {
    queryClient.clear();
  });

  describe('useChannels Hook', () => {
    test('successfully fetches channels data', async () => {
      const { result } = renderHook(() => useChannels(), { wrapper });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.isSuccess).toBe(true);
      expect(result.current.data).toHaveLength(2);
      
      // Check first channel
      const firstChannel = result.current.data?.[0];
      expect(firstChannel).toEqual({
        id: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
        title: 'Google Developers',
        customUrl: '@GoogleDevelopers',
        thumbnailUrl: 'https://yt3.ggpht.com/mock-thumbnail',
        createdAt: '2023-01-01T00:00:00Z',
        lastVideoPublishedAt: '2024-01-15T10:00:00Z',
      });
    });

    test('handles loading state correctly', async () => {
      const { result } = renderHook(() => useChannels(), { wrapper });

      // Check loaded data from MSW
      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.isSuccess).toBe(true);
      expect(result.current.error).toBeNull();
    });

    test('handles error state when MSW returns error', async () => {
      // Override MSW to return error
      server.use(
        http.get('/api/channels', () => {
          return new HttpResponse(null, { status: 500 })
        })
      );

      const { result } = renderHook(() => useChannels(), { wrapper });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.isError).toBe(true);
      expect(result.current.data).toBeUndefined();

      // Reset MSW handlers
      server.resetHandlers();
    });
  });

  describe('useVideos Hook', () => {
    test('successfully fetches videos data', async () => {
      const { result } = renderHook(() => useVideos(), { wrapper });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.isSuccess).toBe(true);
      expect(result.current.data).toHaveLength(2);
      
      // Check first video with new VideoEntry structure
      const firstVideo = result.current.data?.[0];
      expect(firstVideo).toEqual({
        id: 'dQw4w9WgXcQ',
        channelId: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
        cachedAt: '2024-01-15T10:00:00Z',
        watched: false,
        title: 'Introduction to React Testing',
        link: {
          Href: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
          Rel: 'alternate',
        },
        published: '2024-01-15T10:00:00Z',
        content: 'Learn how to test React components effectively with modern testing tools.',
        author: {
          name: 'Google Developers',
          uri: 'https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw',
        },
        mediaGroup: {
          mediaThumbnail: {
            URL: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg',
            Width: '1280',
            Height: '720',
          },
          mediaTitle: 'Introduction to React Testing',
          mediaContent: {
            URL: 'https://www.youtube.com/v/dQw4w9WgXcQ?version=3',
            Type: 'application/x-shockwave-flash',
            Width: '640',
            Height: '390',
          },
          mediaDescription: 'Learn how to test React components effectively with modern testing tools.',
        },
      });
    });

    test('handles refresh parameter correctly', async () => {
      // Override to test refresh behavior
      server.use(
        http.get('/api/videos', ({ request }) => {
          const url = new URL(request.url);
          const refresh = url.searchParams.get('refresh');
          
          if (refresh === 'true') {
            return HttpResponse.json({
              videos: [{
                id: 'refreshed123',
                channelId: 'UC_refresh',
                cachedAt: '2024-01-20T00:00:00Z',
                watched: false,
                title: 'Refreshed Video Title',
                link: {
                  Href: 'https://www.youtube.com/watch?v=refreshed123',
                  Rel: 'alternate',
                },
                published: '2024-01-20T00:00:00Z',
                content: 'Refreshed video content.',
                author: {
                  name: 'Refresh Channel',
                  uri: 'https://www.youtube.com/channel/UC_refresh',
                },
                mediaGroup: {
                  mediaThumbnail: {
                    URL: 'https://refresh.example.com/thumb.jpg',
                    Width: '1280',
                    Height: '720',
                  },
                  mediaTitle: 'Refreshed Video Title',
                  mediaContent: {
                    URL: 'https://www.youtube.com/v/refreshed123?version=3',
                    Type: 'application/x-shockwave-flash',
                    Width: '640',
                    Height: '390',
                  },
                  mediaDescription: 'Refreshed video content.',
                },
              }],
              lastRefresh: '2024-01-20T00:00:00Z',
              totalCount: 1,
            });
          }
          
          return HttpResponse.json({
            videos: [],
            lastRefresh: '2024-01-15T10:00:00Z',
            totalCount: 0,
          });
        })
      );

      const { result } = renderHook(() => useVideos({ refresh: true }), { wrapper });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.isSuccess).toBe(true);
      expect(result.current.data).toHaveLength(1);
      expect(result.current.data?.[0].title).toBe('Refreshed Video Title');

      // Reset MSW handlers
      server.resetHandlers();
    });
  });

  describe('useConfig Hook', () => {
    test('successfully fetches SMTP config', async () => {
      const { result } = renderHook(() => useConfig(), { wrapper });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.isSuccess).toBe(true);
      expect(result.current.data).toEqual({
        host: 'smtp.gmail.com',
        port: 587,
        username: 'test@example.com',
        password: '',
        fromAddress: 'test@example.com',
        recipientEmail: 'recipient@example.com',
        emailHour: 9,
        emailMinute: 0,
        emailTimezone: 'America/New_York',
      });
    });
  });

  describe('React Query Caching and State Management', () => {
    test('caches data correctly between renders', async () => {
      // First render
      const { result: result1, unmount } = renderHook(() => useChannels(), { wrapper });

      await waitFor(() => {
        expect(result1.current.isLoading).toBe(false);
      });

      const firstRenderData = result1.current.data;
      unmount();

      // Second render should use cached data
      const { result: result2 } = renderHook(() => useChannels(), { wrapper });

      // Should immediately have data from cache
      expect(result2.current.data).toBeDefined();
      // Data should be the same (MSW returns consistent mock data)
      expect(result2.current.data).toEqual(firstRenderData);
    });

    test('handles stale-while-revalidate correctly', async () => {
      // Create a query client with stale time
      const staleQueryClient = new QueryClient({
        defaultOptions: {
          queries: {
            retry: false,
            gcTime: 1000 * 60 * 5, // 5 minutes (renamed from cacheTime in React Query v5)
            staleTime: 1000 * 60, // 1 minute
          },
        },
      });

      const staleWrapper = createWrapper(staleQueryClient);

      // First fetch
      const { result, rerender } = renderHook(() => useChannels(), { 
        wrapper: staleWrapper 
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.isSuccess).toBe(true);
      const initialData = result.current.data;

      // Rerender - should use cached data
      rerender();

      expect(result.current.data).toEqual(initialData);
      expect(result.current.isStale).toBe(false);

      staleQueryClient.clear();
    });

    test('refetch functionality works correctly', async () => {
      const { result } = renderHook(() => useChannels(), { wrapper });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.isSuccess).toBe(true);
      
      // Trigger refetch
      await result.current.refetch();

      expect(result.current.isSuccess).toBe(true);
      // Should still have data after refetch
      expect(result.current.data).toHaveLength(2);
    });

    test('handles invalidation correctly', async () => {
      const { result } = renderHook(() => useChannels(), { wrapper });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      // Invalidate the query
      queryClient.invalidateQueries({ queryKey: ['channels'] });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.data).toHaveLength(2); // Fresh data from MSW
    });
  });

  describe('Query Key Management', () => {
    test('different query keys create separate cache entries', async () => {
      const { result: channelsResult } = renderHook(() => useChannels(), { wrapper });
      const { result: videosResult } = renderHook(() => useVideos(), { wrapper });

      await waitFor(() => {
        expect(channelsResult.current.isLoading).toBe(false);
        expect(videosResult.current.isLoading).toBe(false);
      });

      expect(channelsResult.current.data).toHaveLength(2);
      expect(videosResult.current.data).toHaveLength(2);
      
      // Data should be different (channels vs videos)
      expect(channelsResult.current.data).not.toEqual(videosResult.current.data);
    });

    test('parametrized queries create separate cache entries', async () => {
      const { result: normalVideos } = renderHook(() => useVideos(), { wrapper });
      const { result: refreshVideos } = renderHook(() => useVideos({ refresh: true }), { wrapper });

      await waitFor(() => {
        expect(normalVideos.current.isLoading).toBe(false);
        expect(refreshVideos.current.isLoading).toBe(false);
      });

      // Both should have data but potentially different due to different query keys
      expect(normalVideos.current.data).toBeDefined();
      expect(refreshVideos.current.data).toBeDefined();
    });
  });

  describe('Error Recovery and Retry', () => {
    test('does not retry on error when retry is disabled', async () => {
      const { server } = await import('../mocks/server');
      
      // Set up error response
      server.use(
        http.get('/api/channels', () => {
          return new HttpResponse(null, { status: 500 })
        })
      );

      const { result } = renderHook(() => useChannels(), { wrapper });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.isError).toBe(true);
      expect(result.current.error).toBeDefined();

      // Reset handlers
      server.resetHandlers();
    });

    test('can recover from error state', async () => {
      const { server } = await import('../mocks/server');
      
      // First: Set up error
      server.use(
        http.get('/api/channels', () => {
          return new HttpResponse(null, { status: 500 })
        })
      );

      const { result } = renderHook(() => useChannels(), { wrapper });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      // Then: Fix the error by resetting handlers
      server.resetHandlers();

      // Refetch should now succeed
      await result.current.refetch();

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toHaveLength(2);
    });
  });
}); 