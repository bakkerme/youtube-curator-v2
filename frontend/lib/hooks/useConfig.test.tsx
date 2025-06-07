import { renderHook, waitFor } from '@testing-library/react';
import { useConfig } from './useConfig';
import { getRuntimeConfig } from '../config';

// Mock the config module
jest.mock('../config', () => ({
  getRuntimeConfig: jest.fn(),
}));

const mockGetRuntimeConfig = getRuntimeConfig as jest.MockedFunction<typeof getRuntimeConfig>;

describe('useConfig Hook', () => {
  beforeEach(() => {
    mockGetRuntimeConfig.mockClear();
  });

  it('should load config successfully', async () => {
    const mockConfig = { apiUrl: 'http://localhost:8080/api' };
    mockGetRuntimeConfig.mockResolvedValue(mockConfig);

    const { result } = renderHook(() => useConfig());

    // Initially loading
    expect(result.current.loading).toBe(true);
    expect(result.current.config).toBe(null);
    expect(result.current.error).toBe(null);

    // Wait for config to load
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.config).toEqual(mockConfig);
    expect(result.current.error).toBe(null);
    expect(mockGetRuntimeConfig).toHaveBeenCalledTimes(1);
  });

  it('should handle config loading error', async () => {
    const mockError = new Error('Failed to fetch config');
    mockGetRuntimeConfig.mockRejectedValue(mockError);

    const { result } = renderHook(() => useConfig());

    // Initially loading
    expect(result.current.loading).toBe(true);
    expect(result.current.config).toBe(null);
    expect(result.current.error).toBe(null);

    // Wait for error
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.config).toBe(null);
    expect(result.current.error).toBe('Failed to fetch config');
    expect(mockGetRuntimeConfig).toHaveBeenCalledTimes(1);
  });

  it('should handle non-Error rejection', async () => {
    mockGetRuntimeConfig.mockRejectedValue('String error');

    const { result } = renderHook(() => useConfig());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.config).toBe(null);
    expect(result.current.error).toBe('Failed to load configuration');
  });

  it('should handle component unmount during loading', async () => {
    let resolveConfig: (value: { apiUrl: string }) => void;
    const configPromise = new Promise<{ apiUrl: string }>((resolve) => {
      resolveConfig = resolve;
    });
    mockGetRuntimeConfig.mockReturnValue(configPromise);

    const { result, unmount } = renderHook(() => useConfig());

    expect(result.current.loading).toBe(true);

    // Unmount before config loads
    unmount();

    // Resolve the config after unmount
    resolveConfig!({ apiUrl: 'http://localhost:8080/api' });

    // Wait a bit to ensure the cleanup worked
    await new Promise(resolve => setTimeout(resolve, 10));

    // Should not have updated state after unmount
    expect(result.current.loading).toBe(true);
  });

  it('should not reload config on re-render', async () => {
    const mockConfig = { apiUrl: 'http://localhost:8080/api' };
    mockGetRuntimeConfig.mockResolvedValue(mockConfig);

    const { result, rerender } = renderHook(() => useConfig());

    // Wait for initial load
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.config).toEqual(mockConfig);
    expect(mockGetRuntimeConfig).toHaveBeenCalledTimes(1);

    // Re-render the hook
    rerender();

    // Should not call getRuntimeConfig again
    expect(mockGetRuntimeConfig).toHaveBeenCalledTimes(1);
    expect(result.current.config).toEqual(mockConfig);
    expect(result.current.loading).toBe(false);
  });
}); 