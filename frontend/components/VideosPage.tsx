'use client';

import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { Search, Calendar, RefreshCw, List, ChevronDown } from 'lucide-react';
import { videoAPI, channelAPI } from '@/lib/api';
import { VideoEntry, Channel } from '@/lib/types';
import VideoCard from '@/components/VideoCard';
import Pagination from '@/components/Pagination';
import { useRouter, useSearchParams } from 'next/navigation';
import { useWindowTitle } from '@/lib/hooks/useWindowTitle';

const VIDEOS_PER_PAGE = 12;
const WATCHED_VIDEOS_PER_PAGE = 8;

interface VideosPageProps {
  enableAutoRefresh?: boolean;
}

export default function VideosPage({ enableAutoRefresh = true }: VideosPageProps = {}) {
  const router = useRouter();
  const searchParams = useSearchParams();
  
  const [allVideos, setAllVideos] = useState<VideoEntry[]>([]); // Renamed from 'videos'
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterMode, setFilterMode] = useState<'all' | 'today' | 'perDay'>('today');
  const [selectedDate, setSelectedDate] = useState<string>(() => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    // Format as YYYY-MM-DD in local timezone, not UTC
    const year = yesterday.getFullYear();
    const month = String(yesterday.getMonth() + 1).padStart(2, '0');
    const day = String(yesterday.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  });
  const [lastApiRefreshTimestamp, setLastApiRefreshTimestamp] = useState<string | null>(null);
  const [isWatchedAccordionOpen, setIsWatchedAccordionOpen] = useState(false); // Accordion collapsed by default
  
  // Use refs to access current values in the auto-refresh effect
  const allVideosRef = useRef<VideoEntry[]>([]);
  const loadingRef = useRef(true);
  const refreshingRef = useRef(false);

  // Helper function to normalize dates to local timezone (00:00:00)
  const normalizeToLocalDate = useCallback((date: Date): Date => {
    const newDate = new Date(date);
    newDate.setHours(0, 0, 0, 0);
    return newDate;
  }, []);

  // Helper function to create a local date from YYYY-MM-DD string
  const createLocalDateFromString = useCallback((dateString: string): Date => {
    // Parse as local date by appending T00:00:00 (no timezone offset)
    return new Date(dateString + 'T00:00:00');
  }, []);

  // Helper function to get the start time for "intelligent today" filter
  // Includes videos from 10 PM the previous day in user's local timezone
  const getIntelligentTodayStartTime = useCallback((): Date => {
    const now = new Date();
    const yesterdayAt10PM = new Date(now);
    yesterdayAt10PM.setDate(now.getDate() - 1);
    yesterdayAt10PM.setHours(22, 0, 0, 0); // 10:00 PM yesterday
    return yesterdayAt10PM;
  }, []);
  
  // Update refs whenever state changes
  useEffect(() => {
    allVideosRef.current = allVideos;
  }, [allVideos]);
  
  useEffect(() => {
    loadingRef.current = loading;
  }, [loading]);
  
  useEffect(() => {
    refreshingRef.current = refreshing;
  }, [refreshing]);
  
  // Update window title during refresh operations
  useWindowTitle('Refreshing...', refreshing);
  
  // Get current page from URL params
  const currentPage = parseInt(searchParams.get('page') || '1', 10);
  const currentWatchedPage = parseInt(searchParams.get('watched_page') || '1', 10);

  const handleFilterModeChange = () => {
    if (filterMode === 'today') {
      setFilterMode('perDay');
    } else if (filterMode === 'perDay') {
      setFilterMode('all');
    } else {
      setFilterMode('today');
    }
  };

  // Load data function
  const loadData = useCallback(async (refresh = false) => {
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

      if (videosData && videosData.lastRefresh) {
        setLastApiRefreshTimestamp(videosData.lastRefresh);
      }
      setError(null);
    } catch (err) {
      console.error("Error in loadData:",err);
      setError(err instanceof Error ? err.message : 'Failed to load videos');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Intentionally empty deps to prevent infinite re-renders caused by allVideos.length and loading changes

  // Load data on component mount
  useEffect(() => {
    loadData();
  }, []);

  // Handle refresh button click
  const handleRefresh = useCallback(() => {
    // Clear lastApiRefreshTimestamp to ensure the refresh check logic
    // doesn't immediately re-trigger if the day hasn't changed yet
    // but we want a manual refresh.
    // Or, more simply, loadData(true) will fetch new data and update it.
    loadData(true);
  }, []);

  // Auto-refresh logic
  useEffect(() => {
    // Skip auto-refresh if disabled via props
    if (!enableAutoRefresh) {
      return;
    }

    const checkAndRefreshIfNeeded = () => {
      if (loadingRef.current || refreshingRef.current || !lastApiRefreshTimestamp) {
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

          const hasVideosForNewDay = allVideosRef.current.some(video => {
            const videoPublishedDate = new Date(video.published);
            const videoDay = new Date(videoPublishedDate.getFullYear(), videoPublishedDate.getMonth(), videoPublishedDate.getDate());
            return videoDay.getTime() === currentDay.getTime();
          });

          if (!hasVideosForNewDay) {
            console.log('Auto-refresh: No videos found for the new current day. Triggering refresh.');
            loadData(true);
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
  }, [lastApiRefreshTimestamp]); // Only depend on lastApiRefreshTimestamp

  // Filter videos based on search and date filters, then separate into watched/unwatched
  const { unwatchedVideos, watchedVideos } = useMemo(() => {
    let dateFilteredVideos = allVideos;

    // Apply date filtering based on filterMode
    if (filterMode === 'today') {
      // Intelligent today filter: include videos from 10 PM yesterday onwards
      const intelligentTodayStartTime = getIntelligentTodayStartTime();
      const todayEndTime = new Date();
      todayEndTime.setHours(23, 59, 59, 999); // End of today
      
      dateFilteredVideos = dateFilteredVideos.filter((video: VideoEntry) => {
        const videoDate = new Date(video.published);
        return videoDate >= intelligentTodayStartTime && videoDate <= todayEndTime;
      });
    } else if (filterMode === 'perDay') {
      if (selectedDate) {
        try {
          const perDayDate = createLocalDateFromString(selectedDate);
          if (isNaN(perDayDate.getTime())) {
            console.warn("Invalid selectedDate (parsed as NaN):", selectedDate);
            dateFilteredVideos = [];
          } else {
            const perDayNormalized = normalizeToLocalDate(perDayDate);
            dateFilteredVideos = dateFilteredVideos.filter((video: VideoEntry) => {
              const videoDate = new Date(video.published);
              return normalizeToLocalDate(videoDate).getTime() === perDayNormalized.getTime();
            });
          }
        } catch (e) {
          console.error("Error parsing selectedDate catch:", e);
          dateFilteredVideos = [];
        }
      } else {
        console.warn("selectedDate is empty, showing no videos for perDay mode.");
        dateFilteredVideos = [];
      }
    }
    
    // If filterMode is 'all', no date filtering is applied, dateFilteredVideos remains 'allVideos'.
    let searchFilteredVideos = dateFilteredVideos;
    // Apply search filter
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      searchFilteredVideos = dateFilteredVideos.filter((video: VideoEntry) => {
        const title = video.title.toLowerCase();
        const channel = channels.find(c => c.id === video.channelId);
        const channelTitle = channel?.title.toLowerCase() || '';
        return title.includes(query) || channelTitle.includes(query);
      });
    }

    // Separate into watched and unwatched
    const unwatched = searchFilteredVideos.filter(video => !video.watched);
    const watched = searchFilteredVideos.filter(video => video.watched);

    return { unwatchedVideos: unwatched, watchedVideos: watched };
  }, [allVideos, channels, searchQuery, filterMode, selectedDate, normalizeToLocalDate, createLocalDateFromString, getIntelligentTodayStartTime]);

  // Calculate pagination for unwatched videos
  const totalUnwatchedPages = unwatchedVideos.length > 0 ? Math.ceil(unwatchedVideos.length / VIDEOS_PER_PAGE) : 1;
  const startIndexUnwatched = (currentPage - 1) * VIDEOS_PER_PAGE;
  const paginatedUnwatchedVideos = unwatchedVideos.slice(startIndexUnwatched, startIndexUnwatched + VIDEOS_PER_PAGE);

  // Calculate pagination for watched videos
  const totalWatchedPages = watchedVideos.length > 0 ? Math.ceil(watchedVideos.length / WATCHED_VIDEOS_PER_PAGE) : 1;
  const startIndexWatched = (currentWatchedPage - 1) * WATCHED_VIDEOS_PER_PAGE;
  const paginatedWatchedVideos = watchedVideos.slice(startIndexWatched, startIndexWatched + WATCHED_VIDEOS_PER_PAGE);

  // Handle page change for unwatched videos
  const handleUnwatchedPageChange = (page: number) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set('page', page.toString());
    router.push(`/?${params.toString()}`);
  };

  // Handle page change for watched videos
  const handleWatchedPageChange = (page: number) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set('watched_page', page.toString());
    router.push(`/?${params.toString()}`);
  };

  // Handle video watched status change without refetching from API
  const handleVideoWatchedStatusChange = useCallback((videoId: string) => {
    setAllVideos(prevVideos => 
      prevVideos.map(video => 
        video.id === videoId 
          ? { ...video, watched: true }
          : video
      )
    );
  }, []);

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

        {/* Filter Controls Container */}
        <div className="flex flex-col sm:flex-row gap-2">
          <button
            data-testid="filter-mode-button" // Added data-testid
            onClick={handleFilterModeChange}
            className={`flex items-center justify-center gap-2 px-4 py-2 rounded-lg border transition-colors text-sm sm:text-base cursor-pointer ${
              (filterMode === 'today' || filterMode === 'perDay')
                ? 'bg-red-600 text-white border-red-600' // Active style for 'today' and 'perDay'
                : filterMode === 'all'
                  ? 'bg-blue-600 text-white border-blue-600' // Active style for 'all' (example: blue)
                  : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700' // Default
            }`}
          >
            {filterMode === 'today' && <Calendar className="w-4 h-4" />}
            {filterMode === 'perDay' && <Calendar className="w-4 h-4" />}
            {filterMode === 'all' && <List className="w-4 h-4" />}
            {filterMode === 'today' ? 'Today' : filterMode === 'perDay' ? 'Per Day' : 'All Videos'}
          </button>

          {filterMode === 'perDay' && (
            <input
              type="date"
              data-testid="date-filter-input" // Added data-testid
              value={selectedDate}
              onChange={(e) => setSelectedDate(e.target.value)}
              className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent dark:bg-gray-800 dark:border-gray-700 dark:text-white"
              // Optional: Add max date to prevent selecting future dates if needed
              // max={new Date().toISOString().split('T')[0]}
            />
          )}
        </div>

        {/* Refresh Button */}
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg border transition-colors ${
            refreshing
              ? 'bg-gray-100 text-gray-400 border-gray-200 cursor-not-allowed dark:bg-gray-700 dark:text-gray-500 dark:border-gray-600'
              : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700 cursor-pointer'
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
            {paginatedUnwatchedVideos.map((video) => (
              <VideoCard
                key={video.id}
                video={video}
                channels={channels}
                onWatchedStatusChange={handleVideoWatchedStatusChange}
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
          {/* Accordion Header */}
          <button
            onClick={() => setIsWatchedAccordionOpen(!isWatchedAccordionOpen)}
            className="flex items-center gap-2 w-full text-left group focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 rounded-lg p-1"
            aria-expanded={isWatchedAccordionOpen}
            aria-controls="watched-videos-content"
          >
            <h2 className="text-2xl font-semibold">Watched</h2>
            <ChevronDown 
              className={`w-6 h-6 text-gray-500 transition-transform duration-200 group-hover:text-gray-700 dark:text-gray-400 dark:group-hover:text-gray-200 ${
                isWatchedAccordionOpen ? 'rotate-180' : ''
              }`}
            />
          </button>
          
          {/* Accordion Content */}
          {isWatchedAccordionOpen && (
            <div id="watched-videos-content" className="mt-4">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                {paginatedWatchedVideos.map((video) => (
                  <VideoCard
                    key={video.id}
                    video={video}
                    channels={channels}
                    onWatchedStatusChange={handleVideoWatchedStatusChange}
                  />
                ))}
              </div>
              {/* Pagination for Watched Videos */}
              {watchedVideos.length > WATCHED_VIDEOS_PER_PAGE && (
                <Pagination
                  currentPage={currentWatchedPage}
                  totalPages={totalWatchedPages}
                  onPageChange={handleWatchedPageChange}
                />
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
} 