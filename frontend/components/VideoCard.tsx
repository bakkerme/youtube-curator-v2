import { VideoEntry, Channel } from '@/lib/types';
import { videoAPI } from '@/lib/api'; // Import videoAPI
import { useState, useEffect } from 'react'; // Import useState and useEffect
import { formatDistanceToNow } from 'date-fns';
import Image from 'next/image';
import Link from 'next/link';

// Helper function to extract raw video ID from full format
function extractRawVideoId(fullVideoId: string): string {
  // If the video ID starts with 'yt:video:', extract the raw part
  if (fullVideoId.startsWith('yt:video:')) {
    return fullVideoId.substring('yt:video:'.length);
  }
  // If it's already a raw ID, return as-is
  return fullVideoId;
}

interface VideoCardProps {
  video: VideoEntry;
  channels: Channel[];
  onWatchedStatusChange?: (videoId: string) => void; // Add callback prop with video ID
}

export default function VideoCard({ video, channels, onWatchedStatusChange }: VideoCardProps) {
  const [isChecked, setIsChecked] = useState(video.watched);

  // Effect to synchronize isChecked with prop changes
  useEffect(() => {
    setIsChecked(video.watched);
  }, [video.watched]);

  const handleCheckboxChange = async () => {
    const originalCheckedState = isChecked;
    setIsChecked(!originalCheckedState); // Optimistic update

    try {
      await videoAPI.markAsWatched(video.id);
      if (onWatchedStatusChange) {
        onWatchedStatusChange(video.id); // Pass video ID to callback
      }
    } catch (error) {
      console.error('Failed to mark video as watched:', error);
      setIsChecked(originalCheckedState); // Revert on error
    }
  };
  
  // Find the channel title
  const channel = channels.find(c => c.id === video.channelId);
  const channelTitle = channel?.title || 'Unknown Channel';
  
  // Format the published date
  const publishedDate = new Date(video.published);
  const timeAgo = formatDistanceToNow(publishedDate, { addSuffix: true });
  
  // Get thumbnail URL with fallback
  const thumbnailUrl = video.mediaGroup?.mediaThumbnail?.url || '/placeholder-video.svg';
  
  // Clean up title
  const title = video.title || 'Untitled Video';
  
  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-all min-h-[400px] flex flex-col ${isChecked ? 'opacity-60' : ''}`}>
      {/* Thumbnail */}
      <a
        href={video.link.href}
        target="_blank"
        rel="noopener noreferrer"
        className="block"
        aria-label={`${title} - thumbnail`}
      >
        <div className="relative aspect-video">
          <Image
            src={thumbnailUrl}
            alt={title}
            fill
            className="object-cover"
            sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
          />
        </div>
      </a>
      
      {/* Content */}
      <div className="p-4 flex flex-col justify-between flex-1">
        <div>
          <div className="mb-2">
            <Link
              href={`/watch/${extractRawVideoId(video.id)}`}
              aria-label={title}
            >
              <h3 className="font-semibold text-gray-900 dark:text-white line-clamp-2 hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
                {title}
              </h3>
            </Link>
          </div>
          
          <div className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
            <p className="font-medium">{channelTitle}</p>
            <p>{timeAgo}</p>
          </div>
        </div>
        
        {/* Watch buttons and Watched checkbox */}
        <div className="mt-3 space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex space-x-2">
              <Link
                href={`/watch/${extractRawVideoId(video.id)}`}
                className="inline-flex items-center justify-center px-2 py-1.5 text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors whitespace-nowrap"
              >
                Watch in Curator
              </Link>
              <a
                href={video.link.href}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center px-2 py-1.5 text-xs font-medium text-white bg-red-600 hover:bg-red-700 rounded-md transition-colors whitespace-nowrap"
              >
                Watch on YouTube
              </a>
            </div>
            
            <label htmlFor={`watched-${video.id}`} className="flex items-center space-x-1 cursor-pointer text-xs text-gray-500 dark:text-gray-400">
              <span>Watched</span>
              <input
                type="checkbox"
                id={`watched-${video.id}`}
                name={`watched-${video.id}`}
                checked={isChecked}
                onChange={handleCheckboxChange}
                className="form-checkbox h-4 w-4 text-red-600 border-gray-300 rounded focus:ring-red-500 dark:border-gray-600 dark:bg-gray-700 dark:focus:ring-red-600 dark:ring-offset-gray-800"
              />
            </label>
          </div>
        </div>
      </div>
    </div>
  );
} 