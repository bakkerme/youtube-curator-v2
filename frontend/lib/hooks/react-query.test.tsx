import { renderHook, waitFor, act } from '@testing-library/react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rest } from 'msw';
import { server } from '../mocks/server';
import { createTestQueryClient } from '../test-utils';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import axios from 'axios';

// Simple test API client that works with MSW
const testAPI = axios.create({
  baseURL: '', // Let MSW intercept requests
});

// Test wrapper component for React Query
const createWrapper = (queryClient?: QueryClient) => {
  const client = queryClient || createTestQueryClient();
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>
      {children}
    </QueryClientProvider>
  );
};

describe('React Query Integration Tests', () => {
  describe('useQuery Integration', () => {
    it('should fetch and cache data successfully', async () => {
      const { result } = renderHook(
        () => useQuery({
          queryKey: ['channels'],
          queryFn: async () => {
            const response = await testAPI.get('/api/channels');
            return response.data;
          },
        }),
        { wrapper: createWrapper() }
      );

      // Initially loading
      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();

      // Wait for data to load
      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Check loaded data from MSW
      expect(result.current.data).toHaveLength(2);
      expect(result.current.data[0]).toEqual({
        id: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
        title: 'Google Developers',
        customUrl: '@GoogleDevelopers',
        thumbnailUrl: 'https://yt3.ggpht.com/mock-thumbnail',
        createdAt: '2023-01-01T00:00:00Z',
        lastVideoPublishedAt: '2024-01-15T10:00:00Z',
      });
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBe(null);
    });

    it('should handle query errors gracefully', async () => {
      // Override with error handler
      server.use(
        rest.get('/api/channels', (req, res, ctx) => {
          return res(ctx.status(500), ctx.json({ message: 'Internal server error' }));
        })
      );

      const { result } = renderHook(
        () => useQuery({
          queryKey: ['channels-error-test'],
          queryFn: async () => {
            const response = await testAPI.get('/api/channels');
            return response.data;
          },
          retry: false, // Don't retry for this test
        }),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toBeTruthy();
      expect(result.current.data).toBeUndefined();
    });

    it('should refetch data when invalidated', async () => {
      const queryClient = createTestQueryClient();
      
      const { result } = renderHook(
        () => useQuery({
          queryKey: ['channels-refetch'],
          queryFn: async () => {
            const response = await testAPI.get('/api/channels');
            return response.data;
          },
        }),
        { wrapper: createWrapper(queryClient) }
      );

      // Wait for initial data
      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      const initialData = result.current.data;

      // Invalidate the query and wait for refetch to complete
      await act(async () => {
        await queryClient.invalidateQueries({ queryKey: ['channels-refetch'] });
      });

      // Wait for the refetch to complete
      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Data should be the same (MSW returns consistent mock data)
      expect(result.current.data).toEqual(initialData);
    });
  });

  describe('useMutation Integration', () => {
    it('should execute mutations successfully', async () => {
      const { result } = renderHook(
        () => useMutation({
          mutationFn: async (channelData: { channelId: string }) => {
            const response = await testAPI.post('/api/channels', channelData);
            return response.data;
          },
        }),
        { wrapper: createWrapper() }
      );

      // Execute mutation
      act(() => {
        result.current.mutate({ channelId: 'UCtest123' });
      });

      // Wait for mutation to complete
      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual({
        id: 'UCtest123',
        title: 'New Test Channel',
        customUrl: '@newtestchannel',
        thumbnailUrl: 'https://yt3.ggpht.com/new-mock',
        createdAt: expect.any(String),
        lastVideoPublishedAt: expect.any(String),
      });
    });

    it('should handle mutation errors', async () => {
      // Override with error response
      server.use(
        rest.post('/api/channels', (req, res, ctx) => {
          return res(ctx.status(400), ctx.json({ message: 'Invalid channel data' }));
        })
      );

      const { result } = renderHook(
        () => useMutation({
          mutationFn: async (channelData: { channelId: string }) => {
            const response = await testAPI.post('/api/channels', channelData);
            return response.data;
          },
        }),
        { wrapper: createWrapper() }
      );

      act(() => {
        result.current.mutate({ channelId: 'invalid' });
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toBeTruthy();
    });

    it('should invalidate queries on successful mutation', async () => {
      const queryClient = createTestQueryClient();
      
      // First, set up a query
      const { result: queryResult } = renderHook(
        () => useQuery({
          queryKey: ['channels-mutation-test'],
          queryFn: async () => {
            const response = await testAPI.get('/api/channels');
            return response.data;
          },
        }),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(queryResult.current.isSuccess).toBe(true);
      });

      // Set up mutation that invalidates the query
      const { result: mutationResult } = renderHook(
        () => useMutation({
          mutationFn: async (channelData: { channelId: string }) => {
            const response = await testAPI.post('/api/channels', channelData);
            return response.data;
          },
          onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['channels-mutation-test'] });
          },
        }),
        { wrapper: createWrapper(queryClient) }
      );

      // Execute mutation
      act(() => {
        mutationResult.current.mutate({ channelId: 'UCtest456' });
      });

      await waitFor(() => {
        expect(mutationResult.current.isSuccess).toBe(true);
      });

      // Wait for query invalidation to trigger refetch and complete
      await waitFor(() => {
        expect(queryResult.current.isSuccess).toBe(true);
      });
    });
  });

  describe('Cache Management', () => {
    it('should share cache between multiple query hooks', async () => {
      const queryClient = createTestQueryClient();
      
      // First hook
      const { result: result1 } = renderHook(
        () => useQuery({
          queryKey: ['shared-cache'],
          queryFn: async () => {
            const response = await testAPI.get('/api/channels');
            return response.data;
          },
        }),
        { wrapper: createWrapper(queryClient) }
      );

      await waitFor(() => {
        expect(result1.current.isSuccess).toBe(true);
      });

      // Second hook with same query key
      const { result: result2 } = renderHook(
        () => useQuery({
          queryKey: ['shared-cache'],
          queryFn: async () => {
            const response = await testAPI.get('/api/channels');
            return response.data;
          },
        }),
        { wrapper: createWrapper(queryClient) }
      );

      // Should immediately have cached data
      expect(result2.current.data).toEqual(result1.current.data);
      expect(result2.current.isLoading).toBe(false);
    });

    it('should handle stale-while-revalidate pattern', async () => {
      const queryClient = createTestQueryClient();
      
      // Set initial cache data
      queryClient.setQueryData(['stale-test'], [
        { id: 'cached', title: 'Cached Data' }
      ]);

      const { result } = renderHook(
        () => useQuery({
          queryKey: ['stale-test'],
          queryFn: async () => {
            const response = await testAPI.get('/api/channels');
            return response.data;
          },
          staleTime: 0, // Immediately stale
        }),
        { wrapper: createWrapper(queryClient) }
      );

      // Should show stale data immediately
      expect(result.current.data).toEqual([{ id: 'cached', title: 'Cached Data' }]);
      expect(result.current.isLoading).toBe(false);

      // Wait for background refetch
      await waitFor(() => {
        expect(result.current.data).toHaveLength(2); // Fresh data from MSW
      });

      expect(result.current.data[0].title).toBe('Google Developers');
    });
  });

  describe('Query Configuration', () => {
    it('should respect enabled option', async () => {
      let enabled = false;
      
      const { result, rerender } = renderHook(
        ({ enabled }) => useQuery({
          queryKey: ['conditional-query'],
          queryFn: async () => {
            const response = await testAPI.get('/api/channels');
            return response.data;
          },
          enabled,
        }),
        { 
          wrapper: createWrapper(),
          initialProps: { enabled }
        }
      );

      // Should not fetch when disabled
      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toBeUndefined();

      // Enable the query
      enabled = true;
      rerender({ enabled });

      // Should start fetching
      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toHaveLength(2);
    });

    it('should handle different query keys correctly', async () => {
      const { result: channelsResult } = renderHook(
        () => useQuery({
          queryKey: ['channels'],
          queryFn: async () => {
            const response = await testAPI.get('/api/channels');
            return response.data;
          },
        }),
        { wrapper: createWrapper() }
      );

      const { result: videosResult } = renderHook(
        () => useQuery({
          queryKey: ['videos'],
          queryFn: async () => {
            const response = await testAPI.get('/api/videos');
            return response.data;
          },
        }),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(channelsResult.current.isSuccess).toBe(true);
        expect(videosResult.current.isSuccess).toBe(true);
      });

      // Should have different data
      expect(channelsResult.current.data).toHaveLength(2);
      expect(videosResult.current.data).toHaveLength(2);
      expect(channelsResult.current.data[0].title).toBe('Google Developers');
      expect(videosResult.current.data[0].title).toBe('Introduction to React Testing');
    });
  });
}); 