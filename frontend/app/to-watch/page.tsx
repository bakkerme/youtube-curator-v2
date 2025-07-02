'use client';

import { useState, useEffect, useMemo, useCallback } from 'react';
import { Search, RefreshCw } from 'lucide-react';
import { videoAPI, channelAPI } from '@/lib/api';
import { VideoEntry, Channel } from '@/lib/types';
import VideoCard from '@/components/VideoCard';
import Pagination from '@/components/Pagination';
import { useRouter, useSearchParams } from 'next/navigation';
import { useWindowTitle } from '@/lib/hooks/useWindowTitle';

const VIDEOS_PER_PAGE = 12;

export default function ToWatchPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  
  const [allVideos, setAllVideos] = useState<VideoEntry[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  
  // Update window title during refresh operations
  useWindowTitle('Refreshing...', refreshing);
  
  // Get current page from URL params
  const currentPage = parseInt(searchParams.get('page') || '1', 10);

  // Load data function
  const loadData = useCallback(async (refresh = false) => {
    if (refresh) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError(null);

    try {
      const [videosData, channelsData] = await Promise.all([
        videoAPI.getAll(refresh),
        channelAPI.getAll()
      ]);
      
      // Sort videos by published date, newest first
      const sortedVideos = videosData.videos.sort((a, b) => 
        new Date(b.published).getTime() - new Date(a.published).getTime()
      );
      
      setAllVideos(sortedVideos);
      setChannels(channelsData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  // Initial load
  useEffect(() => {
    loadData();
  }, [loadData]);

  // Callback for when a video's status changes
  const handleVideoStatusChange = useCallback((videoId: string) => {
    // Update the local state to reflect the change
    setAllVideos(prev => prev.map(video => 
      video.id === videoId 
        ? { ...video, toWatch: !video.toWatch }
        : video
    ));
  }, []);

  // Filter videos to show only those marked as "to watch"
  const toWatchVideos = useMemo(() => {
    let filtered = allVideos.filter(video => video.toWatch);
    
    // Apply search filter
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter((video: VideoEntry) => {
        const title = video.title.toLowerCase();
        const channel = channels.find(c => c.id === video.channelId);
        const channelTitle = channel?.title.toLowerCase() || '';
        return title.includes(query) || channelTitle.includes(query);
      });
    }
    
    return filtered;
  }, [allVideos, channels, searchQuery]);

  // Calculate pagination
  const totalPages = toWatchVideos.length > 0 ? Math.ceil(toWatchVideos.length / VIDEOS_PER_PAGE) : 1;
  const startIndex = (currentPage - 1) * VIDEOS_PER_PAGE;
  const paginatedVideos = toWatchVideos.slice(startIndex, startIndex + VIDEOS_PER_PAGE);

  // Handle page change
  const handlePageChange = (page: number) => {
    const params = new URLSearchParams(searchParams);
    params.set('page', page.toString());
    router.push(`/to-watch?${params.toString()}`);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[calc(100vh-200px)]">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600 dark:text-gray-400">Loading videos...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-6">
        <p className="text-red-800 dark:text-red-300">{error}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
        <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            To Watch ({toWatchVideos.length})
          </h1>
          
          <button
            onClick={() => loadData(true)}
            disabled={refreshing}
            className="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            {refreshing ? 'Refreshing...' : 'Refresh'}
          </button>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
          <input
            type="text"
            placeholder="Search videos or channels..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10 pr-4 py-2 w-full border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      </div>

      {/* Videos Grid */}
      {toWatchVideos.length === 0 ? (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-12 text-center">
          <p className="text-gray-500 dark:text-gray-400 text-lg">
            {searchQuery ? 'No videos found matching your search.' : 'No videos in your watch list yet.'}
          </p>
          <p className="text-gray-400 dark:text-gray-500 text-sm mt-2">
            Click the star icon on any video to add it to your watch list.
          </p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {paginatedVideos.map((video) => (
              <VideoCard 
                key={video.id} 
                video={video} 
                channels={channels}
                onWatchedStatusChange={handleVideoStatusChange}
              />
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="mt-8">
              <Pagination
                currentPage={currentPage}
                totalPages={totalPages}
                onPageChange={handlePageChange}
              />
            </div>
          )}
        </>
      )}
    </div>
  );
}