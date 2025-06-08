import { VideoEntry, Channel } from '@/lib/types';
import { videoAPI } from '@/lib/api'; // Import videoAPI
import { useState, useEffect } from 'react'; // Import useState and useEffect
import { formatDistanceToNow } from 'date-fns';
import Image from 'next/image';

interface VideoCardProps {
  video: VideoEntry;
  channels: Channel[];
  onWatchedStatusChange?: () => void; // Add callback prop
}

export default function VideoCard({ video, channels, onWatchedStatusChange }: VideoCardProps) {
  const { entry, channelId, watched } = video; // Destructure watched
  const [isChecked, setIsChecked] = useState(watched);

  // Effect to synchronize isChecked with prop changes
  useEffect(() => {
    setIsChecked(watched);
  }, [watched]);

  const handleCheckboxChange = async () => {
    const originalCheckedState = isChecked;
    setIsChecked(!originalCheckedState); // Optimistic update

    try {
      await videoAPI.markAsWatched(entry.id);
      if (onWatchedStatusChange) {
        onWatchedStatusChange(); // Call the callback to refresh data in parent
      }
    } catch (error) {
      console.error('Failed to mark video as watched:', error);
      setIsChecked(originalCheckedState); // Revert on error
    }
  };
  
  // Find the channel title
  const channel = channels.find(c => c.id === channelId);
  const channelTitle = channel?.title || 'Unknown Channel';
  
  // Format the published date
  const publishedDate = new Date(entry.published);
  const timeAgo = formatDistanceToNow(publishedDate, { addSuffix: true });
  
  // Get thumbnail URL with fallback
  const thumbnailUrl = entry.mediaGroup?.mediaThumbnail?.URL || '/placeholder-video.svg';
  
  // Clean up title
  const title = entry.title || 'Untitled Video';
  
  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-all ${isChecked ? 'opacity-60' : ''}`}>
      {/* Thumbnail */}
      <div className="relative aspect-video">
        <Image
          src={thumbnailUrl}
          alt={title}
          fill
          className="object-cover"
          sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
        />
      </div>
      
      {/* Content */}
      <div className="p-4">
        <div className="flex justify-between items-start mb-2">
          <h3 className="font-semibold text-gray-900 dark:text-white line-clamp-2 flex-1">
            {title}
          </h3>
          <div className="ml-2 flex-shrink-0">
            <label htmlFor={`watched-${entry.id}`} className="flex items-center space-x-1 cursor-pointer text-xs text-gray-500 dark:text-gray-400">
              <span>Watched</span>
              <input
                type="checkbox"
                id={`watched-${entry.id}`}
                name={`watched-${entry.id}`}
                checked={isChecked}
                onChange={handleCheckboxChange}
                className="form-checkbox h-4 w-4 text-red-600 border-gray-300 rounded focus:ring-red-500 dark:border-gray-600 dark:bg-gray-700 dark:focus:ring-red-600 dark:ring-offset-gray-800"
              />
            </label>
          </div>
        </div>
        
        <div className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
          <p className="font-medium">{channelTitle}</p>
          <p>{timeAgo}</p>
        </div>
        
        {/* Watch button */}
        <a
          href={entry.link.Href}
          target="_blank"
          rel="noopener noreferrer"
          className="mt-3 inline-flex items-center px-3 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors"
        >
          Watch on YouTube
        </a>
      </div>
    </div>
  );
} 