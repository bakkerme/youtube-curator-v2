'use client';

import { useParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { videoAPI, channelAPI } from '@/lib/api';
import { formatDistanceToNow } from 'date-fns';
import Link from 'next/link';
import { useEffect } from 'react';

// Helper function to extract raw video ID from full format
function extractRawVideoId(fullVideoId: string): string {
  // If the video ID starts with 'yt:video:', extract the raw part
  if (fullVideoId.startsWith('yt:video:')) {
    return fullVideoId.substring('yt:video:'.length);
  }
  // If it's already a raw ID, return as-is
  return fullVideoId;
}

export default function WatchPage() {
  const params = useParams();
  const rawVideoId = params.videoId as string;

  const { data: videosResponse, isLoading: videosLoading } = useQuery({
    queryKey: ['videos'],
    queryFn: () => videoAPI.getAll(),
  });

  const { data: channels, isLoading: channelsLoading } = useQuery({
    queryKey: ['channels'],
    queryFn: () => channelAPI.getAll(),
  });

  // Find video by comparing raw video IDs
  const video = videosResponse?.videos.find((v) => extractRawVideoId(v.id) === rawVideoId);
  const channel = channels?.find((c) => c.id === video?.channelId);

  // Update window title with video title
  useEffect(() => {
    if (video?.title) {
      const originalTitle = document.title;
      document.title = `${video.title} - Curator`;
      
      // Cleanup: restore original title on unmount
      return () => {
        document.title = originalTitle;
      };
    }
  }, [video?.title]);

  if (videosLoading || channelsLoading) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 py-8">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="animate-pulse">
            <div className="bg-gray-300 dark:bg-gray-700 rounded-lg h-96 mb-6"></div>
            <div className="bg-gray-300 dark:bg-gray-700 rounded h-8 w-3/4 mb-4"></div>
            <div className="bg-gray-300 dark:bg-gray-700 rounded h-6 w-1/2"></div>
          </div>
        </div>
      </div>
    );
  }

  if (!video) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 py-8">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">Video Not Found</h1>
            <p className="text-gray-600 dark:text-gray-400 mb-8">The video you&apos;re looking for doesn&apos;t exist.</p>
            <Link 
              href="/"
              className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors"
            >
              Back to Home
            </Link>
          </div>
        </div>
      </div>
    );
  }

  const youtubeVideoId = rawVideoId;
  const publishedDate = new Date(video.published);
  const timeAgo = formatDistanceToNow(publishedDate, { addSuffix: true });

  return (
    <div className="min-h-screen flex flex-col">
      {/* Video player - 16:9 aspect ratio */}
      <div className="relative">
        {youtubeVideoId ? (
          <div className="relative aspect-video">
            <iframe
              src={`https://www.youtube.com/embed/${youtubeVideoId}`}
              title={video.title}
              frameBorder="0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowFullScreen
              className="absolute inset-0 w-full h-full"
            ></iframe>
          </div>
        ) : (
          <div className="relative aspect-video bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
            <p className="text-gray-500 dark:text-gray-400">Unable to embed video</p>
          </div>
        )}
      </div>

      {/* Video metadata */}
      <div className="mt-8">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
            {video.title}
          </h1>
          
          <div className="flex items-center justify-between mb-4">
            <div className="text-gray-600 dark:text-gray-400">
              <p className="font-medium text-lg">{channel?.title || 'Unknown Channel'}</p>
              <p className="text-sm">Published {timeAgo}</p>
            </div>
            
            <div className="flex space-x-3">
              <a
                href={video.link.href}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors"
              >
                Watch on YouTube
              </a>
            </div>
          </div>

          {/* Video description */}
          {video.content && (
            <div className="mt-6">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">Description</h3>
              <div className="text-gray-700 dark:text-gray-300 prose prose-sm max-w-none">
                <p>{video.content}</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}