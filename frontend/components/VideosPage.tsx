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
  
  const [allVideos, setAllVideos] = useState<VideoEntry[]>([]); // Renamed from 'videos'
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

      if (videosData && videosData.videos) { // Allow empty array to be set
        setAllVideos(videosData.videos);
      } else if (refresh) {
        // If refreshing and no videos came back, it could be that all videos expired
        // or there are genuinely no videos for any channel.
        // Keep existing videos in this case unless it's an initial load.
        // If it's an initial load and no videos, videos will be an empty array.
        if (allVideos.length > 0 && !loading) { // only clear if not initial load
             // Keep stale data on refresh error
        } else {
            setAllVideos([]);
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

          const hasVideosForNewDay = allVideos.some(video => {
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
  }, [lastApiRefreshTimestamp, allVideos, loading, refreshing]); // Added loading and refreshing to deps

  // Filter videos based on search and today filter
  const { unwatchedVideos, watchedVideos } = useMemo(() => {
    let filtered = allVideos;

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

    const unwatched = filtered.filter(video => !video.watched);
    const watched = filtered.filter(video => video.watched);

    return { unwatchedVideos: unwatched, watchedVideos: watched };
  }, [allVideos, channels, searchQuery, showTodayOnly]);

  // Calculate pagination for unwatched videos
  const totalUnwatchedPages = unwatchedVideos.length > 0 ? Math.ceil(unwatchedVideos.length / VIDEOS_PER_PAGE) : 1;
  const startIndexUnwatched = (currentPage - 1) * VIDEOS_PER_PAGE;
  const paginatedUnwatchedVideos = unwatchedVideos.slice(startIndexUnwatched, startIndexUnwatched + VIDEOS_PER_PAGE);

  // For watched videos, we'll display a smaller, non-paginated list or paginated if needed.
  // For simplicity in this step, let's show all watched videos or a fixed number.
  // Let's paginate watched videos as well for consistency.
  const WATCHED_VIDEOS_PER_PAGE = 8; // Can be different from unwatched
  const currentWatchedPage = parseInt(searchParams.get('watched_page') || '1', 10);
  const totalWatchedPages = watchedVideos.length > 0 ? Math.ceil(watchedVideos.length / WATCHED_VIDEOS_PER_PAGE) : 1;
  const startIndexWatched = (currentWatchedPage - 1) * WATCHED_VIDEOS_PER_PAGE;
  const paginatedWatchedVideos = watchedVideos.slice(startIndexWatched, startIndexWatched + WATCHED_VIDEOS_PER_PAGE);


  // Handle page change for unwatched videos
  const handleUnwatchedPageChange = (page: number) => {
    const params = new URLSearchParams(searchParams);
    params.set('page', page.toString());
    router.push(`/?${params.toString()}`);
  };

  // Handle page change for watched videos
  const handleWatchedPageChange = (page: number) => {
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
        <h1 className="text-3xl font-bold">Unwatched Videos</h1>
        <div className="text-sm text-gray-600 dark:text-gray-400">
          {unwatchedVideos.length} video{unwatchedVideos.length !== 1 ? 's' : ''} found
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

      {/* Unwatched Videos Grid */}
      <h2 className="text-2xl font-semibold mt-8 mb-4">Unwatched</h2>
      {paginatedUnwatchedVideos.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            {unwatchedVideos.length === 0 && allVideos.filter(v => !v.watched).length === 0
              ? 'No unwatched videos available. Great job!'
              : 'No unwatched videos match your current filters.'}
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
            {paginatedUnwatchedVideos.map((video) => (
              <VideoCard
                key={video.entry.id}
                video={video}
                channels={channels}
                // Pass loadData to refresh the list when a video is marked as watched
                onWatchedStatusChange={() => loadData(false)}
              />
            ))}
          </div>

          {/* Pagination for Unwatched Videos */}
          {unwatchedVideos.length > VIDEOS_PER_PAGE && (
            <Pagination
              currentPage={currentPage}
              totalPages={totalUnwatchedPages}
              onPageChange={handleUnwatchedPageChange}
            />
          )}
        </>
      )}

      {/* Watched Videos Section */}
      {watchedVideos.length > 0 && (
        <div className="mt-12">
          <h2 className="text-2xl font-semibold mb-4">Watched</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {paginatedWatchedVideos.map((video) => (
              <VideoCard
                key={video.entry.id}
                video={video}
                channels={channels}
                onWatchedStatusChange={() => loadData(false)}
              />
            ))}
          </div>
          {/* Pagination for Watched Videos */}
          {watchedVideos.length > WATCHED_VIDEOS_PER_PAGE && (
            <Pagination
              currentPage={currentWatchedPage}
              totalPages={totalWatchedPages}
              onPageChange={handleWatchedPageChange}
              pageParamName="watched_page" // To distinguish from unwatched pagination
            />
          )}
        </div>
      )}
    </div>
  );
} 