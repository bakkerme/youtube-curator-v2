'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { channelAPI } from '@/lib/api';
import { Channel, ChannelRequest } from '@/lib/types';
import { X, Plus, Loader2 } from 'lucide-react';

export default function SubscriptionsPage() {
  const [channelInput, setChannelInput] = useState('');
  const [channelTitle, setChannelTitle] = useState('');
  const queryClient = useQueryClient();

  // Fetch channels
  const { data: channels = [], isLoading, error } = useQuery({
    queryKey: ['channels'],
    queryFn: channelAPI.getAll,
  });

  // Add channel mutation
  const addChannelMutation = useMutation({
    mutationFn: (request: ChannelRequest) => channelAPI.add(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['channels'] });
      setChannelInput('');
      setChannelTitle('');
    },
  });

  // Remove channel mutation
  const removeChannelMutation = useMutation({
    mutationFn: (channelId: string) => channelAPI.remove(channelId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['channels'] });
    },
  });

  const handleAddChannel = (e: React.FormEvent) => {
    e.preventDefault();
    if (channelInput.trim()) {
      addChannelMutation.mutate({
        url: channelInput.trim(),
        title: channelTitle.trim() || undefined,
      });
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold mb-2">Manage Subscriptions</h1>
      <p className="text-gray-600 dark:text-gray-400 mb-8">
        Add or remove YouTube channels to track their latest uploads.
      </p>

      {/* Add Channel Form */}
      <form onSubmit={handleAddChannel} className="mb-8 space-y-4">
        <div className="flex gap-4">
          <input
            type="text"
            value={channelInput}
            onChange={(e) => setChannelInput(e.target.value)}
            placeholder="Enter Channel ID or YouTube URL"
            className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                     bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-red-500"
            disabled={addChannelMutation.isPending}
          />
          <button
            type="submit"
            disabled={addChannelMutation.isPending || !channelInput.trim()}
            className="px-6 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 
                     disabled:opacity-50 disabled:cursor-not-allowed transition-colors
                     flex items-center gap-2"
          >
            {addChannelMutation.isPending ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Adding...
              </>
            ) : (
              <>
                <Plus className="w-4 h-4" />
                Add Channel
              </>
            )}
          </button>
        </div>
        
        {/* Optional title input */}
        <input
          type="text"
          value={channelTitle}
          onChange={(e) => setChannelTitle(e.target.value)}
          placeholder="Channel Title (optional - will be fetched if not provided)"
          className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                   bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-red-500"
          disabled={addChannelMutation.isPending}
        />
      </form>

      {/* Error Messages */}
      {addChannelMutation.isError && (
        <div className="mb-4 p-4 bg-red-100 dark:bg-red-900/30 border border-red-300 dark:border-red-700 rounded-lg text-red-700 dark:text-red-300">
          {addChannelMutation.error?.message || 'Failed to add channel'}
        </div>
      )}

      {/* Subscribed Channels */}
      <div className="space-y-4">
        <h2 className="text-xl font-semibold">Subscribed Channels</h2>

        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-gray-400" />
          </div>
        ) : error ? (
          <div className="p-4 bg-red-100 dark:bg-red-900/30 border border-red-300 dark:border-red-700 rounded-lg text-red-700 dark:text-red-300">
            Failed to load channels
          </div>
        ) : channels.length === 0 ? (
          <div className="text-center py-12 text-gray-500 dark:text-gray-400">
            No channels subscribed yet. Add your first channel above!
          </div>
        ) : (
          <div className="space-y-3">
            {channels.map((channel) => (
              <div
                key={channel.id}
                className="flex items-center justify-between p-4 bg-white dark:bg-gray-800 
                         border border-gray-200 dark:border-gray-700 rounded-lg
                         hover:shadow-lg transition-shadow"
              >
                <div className="flex items-center gap-4">
                  <div>
                    <a href={`https://www.youtube.com/channel/${channel.id}`} target="_blank" rel="noopener noreferrer" className="font-semibold">{channel.title}</a>
                  </div>
                </div>
                
                <button
                  onClick={() => removeChannelMutation.mutate(channel.id)}
                  disabled={removeChannelMutation.isPending}
                  className="p-2 text-gray-400 hover:text-red-600 dark:hover:text-red-400
                           disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  aria-label={`Remove ${channel.title}`}
                >
                  {removeChannelMutation.isPending && 
                   removeChannelMutation.variables === channel.id ? (
                    <Loader2 className="w-5 h-5 animate-spin" />
                  ) : (
                    <X className="w-5 h-5" />
                  )}
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
} 