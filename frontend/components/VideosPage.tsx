'use client';

import { useState, useEffect, useMemo } from 'react';
import { Search, Calendar, RefreshCw } from 'lucide-react';
import { videoAPI, channelAPI } from '@/lib/api';
import { VideoEntry, Channel, VideosAPIResponse } from '@/lib/types';
import VideoCard from '@/components/VideoCard';
import Pagination from '@/components/Pagination';
import { useRouter, useSearchParams } from 'next/navigation';

const VIDEOS_PER_PAGE = 12;

export default function VideosPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  
  const [videos, setVideos] = useState<VideoEntry[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [showTodayOnly, setShowTodayOnly] = useState(true);
  const [lastApiRefreshTimestamp, setLastApiRefreshTimestamp] = useState<string | null>(null);
  
  // Get current page from URL params
  const currentPage = parseInt(searchParams.get('page') || '1', 10);

  // Load data function
  const loadData = async (refresh = false) => {
    try {
      if (refresh) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      
      const [videosData, channelsData] = await Promise.all([
        videoAPI.getAll(refresh),
        channelAPI.getAll()
      ]);

      if (videosData && videosData.videos && videosData.videos.length > 0) {
        setVideos(videosData.videos);
      } else if (refresh) {
        // If refreshing and no videos came back, it could be that all videos expired
        // or there are genuinely no videos for any channel.
        // Keep existing videos in this case unless it's an initial load.
        // If it's an initial load and no videos, videos will be an empty array.
        if (videos.length > 0 && !loading) { // only clear if not initial load
             // Keep stale data on refresh error
        } else {
            setVideos([]);
        }
      }


      if (channelsData && channelsData.length > 0) {
        setChannels(channelsData);
      }

      if (videosData && videosData.lastRefreshedAt) {
        setLastApiRefreshTimestamp(videosData.lastRefreshedAt);
      }
      setError(null);
    } catch (err) {
      console.error("Error in loadData:",err);
      setError(err instanceof Error ? err.message : 'Failed to load videos');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  // Load data on component mount
  useEffect(() => {
    loadData();
  }, []);

  // Handle refresh button click
  const handleRefresh = () => {
    // Clear lastApiRefreshTimestamp to ensure the refresh check logic
    // doesn't immediately re-trigger if the day hasn't changed yet
    // but we want a manual refresh.
    // Or, more simply, loadData(true) will fetch new data and update it.
    loadData(true);
  };

  // Auto-refresh logic
  useEffect(() => {
    const checkAndRefreshIfNeeded = () => {
      if (loading || refreshing || !lastApiRefreshTimestamp) {
        console.log('Auto-refresh check: Skipping due to loading, refreshing, or no timestamp.');
        return;
      }

      try {
        const lastRefreshDate = new Date(lastApiRefreshTimestamp);
        const currentDate = new Date();

        // Normalize to compare dates only (YYYY-MM-DD)
        const lastRefreshDay = new Date(lastRefreshDate.getFullYear(), lastRefreshDate.getMonth(), lastRefreshDate.getDate());
        const currentDay = new Date(currentDate.getFullYear(), currentDate.getMonth(), currentDate.getDate());

        console.log(`Auto-refresh check: Last refresh on ${lastRefreshDay.toDateString()}, current day is ${currentDay.toDateString()}`);

        if (currentDay.getTime() > lastRefreshDay.getTime()) {
          console.log('Auto-refresh: Day has changed since last API refresh.');

          const hasVideosForNewDay = videos.some(video => {
            const videoPublishedDate = new Date(video.entry.published);
            const videoDay = new Date(videoPublishedDate.getFullYear(), videoPublishedDate.getMonth(), videoPublishedDate.getDate());
            return videoDay.getTime() === currentDay.getTime();
          });

          if (!hasVideosForNewDay) {
            console.log('Auto-refresh: No videos found for the new current day. Triggering refresh.');
            handleRefresh();
          } else {
            console.log('Auto-refresh: Videos already exist for the new current day. No refresh needed.');
          }
        } else {
          console.log('Auto-refresh: Still the same day as last API refresh. No refresh needed based on day change.');
        }
      } catch (e) {
        console.error("Error during auto-refresh check:", e);
      }
    };

    // Initial check
    checkAndRefreshIfNeeded();

    // Check when tab becomes visible
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        console.log('Auto-refresh: Tab became visible. Checking for refresh.');
        checkAndRefreshIfNeeded();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);

    // Periodic check interval (e.g., every 1 minute)
    const intervalId = setInterval(() => {
      console.log('Auto-refresh: Interval check.');
      checkAndRefreshIfNeeded();
    }, 60 * 1000); // 1 minute

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      clearInterval(intervalId);
      console.log('Auto-refresh: Cleaned up visibility listener and interval.');
    };
  }, [lastApiRefreshTimestamp, videos, loading, refreshing]); // Added loading and refreshing to deps

  // Filter videos based on search and today filter
  const filteredVideos = useMemo(() => {
    let filtered = videos;

    // Apply search filter
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(video => {
        const title = video.entry.title.toLowerCase();
        const channel = channels.find(c => c.id === video.channelId);
        const channelTitle = channel?.title.toLowerCase() || '';
        return title.includes(query) || channelTitle.includes(query);
      });
    }

    // Apply today filter
    if (showTodayOnly) {
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      filtered = filtered.filter(video => {
        const videoDate = new Date(video.entry.published);
        videoDate.setHours(0, 0, 0, 0);
        return videoDate.getTime() === today.getTime();
      });
    }

    if (!filtered || filtered.length === 0) {
        return [];
    }

    return filtered;
  }, [videos, channels, searchQuery, showTodayOnly]);

  // Calculate pagination
  const totalPages = filteredVideos.length > 0 ? Math.ceil(filteredVideos.length / VIDEOS_PER_PAGE) : 1;
  const startIndex = (currentPage - 1) * VIDEOS_PER_PAGE;
  const paginatedVideos = filteredVideos.slice(startIndex, startIndex + VIDEOS_PER_PAGE);

  // Handle page change
  const handlePageChange = (page: number) => {
    const params = new URLSearchParams(searchParams);
    params.set('page', page.toString());
    router.push(`/?${params.toString()}`);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <div className="w-8 h-8 border-4 border-red-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600 dark:text-gray-400">Loading videos...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <p className="text-red-600 dark:text-red-400 mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <h1 className="text-3xl font-bold">Latest Videos</h1>
        <div className="text-sm text-gray-600 dark:text-gray-400">
          {filteredVideos.length} video{filteredVideos.length !== 1 ? 's' : ''} found
        </div>
      </div>

      {/* Search and Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        {/* Search */}
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
          <input
            type="text"
            placeholder="Search videos or channels..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent dark:bg-gray-800 dark:border-gray-700 dark:text-white"
          />
        </div>

        {/* Today Filter */}
        <button
          onClick={() => setShowTodayOnly(!showTodayOnly)}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg border transition-colors ${
            showTodayOnly
              ? 'bg-red-600 text-white border-red-600'
              : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700'
          }`}
        >
          <Calendar className="w-4 h-4" />
          Today
        </button>

        {/* Refresh Button */}
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg border transition-colors ${
            refreshing
              ? 'bg-gray-100 text-gray-400 border-gray-200 cursor-not-allowed dark:bg-gray-700 dark:text-gray-500 dark:border-gray-600'
              : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700'
          }`}
        >
          <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
          {refreshing ? 'Refreshing...' : 'Refresh'}
        </button>
      </div>

      {/* Videos Grid */}
      {paginatedVideos.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            {filteredVideos.length === 0 && videos.length === 0
              ? 'No videos available. Add some channels to start seeing videos!'
              : 'No videos match your current filters.'}
          </p>
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
            >
              Clear search
            </button>
          )}
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {paginatedVideos.map((video) => (
              <VideoCard
                key={video.entry.id}
                video={video}
                channels={channels}
              />
            ))}
          </div>

          {/* Pagination */}
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={handlePageChange}
          />
        </>
      )}
    </div>
  );
} 