import { VideoEntry, Channel } from '@/lib/types';
import { formatDistanceToNow } from 'date-fns';
import Image from 'next/image';

interface VideoCardProps {
  video: VideoEntry;
  channels: Channel[];
}

export default function VideoCard({ video, channels }: VideoCardProps) {
  const { entry, channelId } = video;
  
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
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow">
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
        <h3 className="font-semibold text-gray-900 dark:text-white line-clamp-2 mb-2">
          {title}
        </h3>
        
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
          Watch
        </a>
      </div>
    </div>
  );
} 