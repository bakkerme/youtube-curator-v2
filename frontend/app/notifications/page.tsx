'use client';

import { useState, useEffect } from 'react';
import { newsletterAPI, channelAPI } from '@/lib/api';
import { Channel } from '@/lib/types';

export default function NotificationsPage() {
  const [channels, setChannels] = useState<Channel[]>([]);
  const [selectedChannel, setSelectedChannel] = useState<string>('all');
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingChannels, setIsLoadingChannels] = useState(true);
  const [result, setResult] = useState<{ type: 'success' | 'error', message: string } | null>(null);

  useEffect(() => {
    loadChannels();
  }, []);

  const loadChannels = async () => {
    try {
      const data = await channelAPI.getAll();
      setChannels(data);
    } catch (error) {
      console.error('Failed to load channels:', error);
    } finally {
      setIsLoadingChannels(false);
    }
  };

  const handleRunNewsletter = async () => {
    setIsLoading(true);
    setResult(null);

    try {
      const request = selectedChannel === 'all' ? {} : { channelId: selectedChannel };
      const response = await newsletterAPI.run(request);
      
      const successMessage = `Newsletter run completed successfully!\nProcessed ${response.channelsProcessed} channel(s), found ${response.newVideosFound} new video(s).\n${response.emailSent ? 'Email sent.' : 'No email sent (no new videos).'}`;
      
      setResult({ type: 'success', message: successMessage });
    } catch (error: any) {
      setResult({ 
        type: 'error', 
        message: error.message || 'Failed to run newsletter' 
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold mb-2">Notification Settings</h1>
      <p className="text-gray-600 dark:text-gray-400 mb-8">
        Configure email notifications and check intervals.
      </p>

      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm mb-6">
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-xl font-semibold mb-2">Manual Newsletter Trigger</h2>
          <p className="text-gray-600 dark:text-gray-400">
            Manually trigger a newsletter run to check for new videos and send notifications.
            This is useful for debugging and testing.
          </p>
        </div>
        <div className="p-6 space-y-4">
          <div className="flex gap-4 items-end">
            <div className="flex-1">
              <label htmlFor="channel-select" className="text-sm font-medium mb-2 block">
                Select Channel (optional)
              </label>
              <select
                id="channel-select"
                value={selectedChannel}
                onChange={(e) => setSelectedChannel(e.target.value)}
                disabled={isLoading || isLoadingChannels}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <option value="all">All Channels</option>
                {channels.map((channel) => (
                  <option key={channel.id} value={channel.id}>
                    {channel.title}
                  </option>
                ))}
              </select>
            </div>
            <button
              onClick={handleRunNewsletter}
              disabled={isLoading || isLoadingChannels}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              {isLoading ? (
                <>
                  <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Running...
                </>
              ) : (
                <>
                  <svg className="h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <line x1="22" y1="2" x2="11" y2="13"></line>
                    <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                  </svg>
                  Run Newsletter
                </>
              )}
            </button>
          </div>

          {result && (
            <div className={`p-4 rounded-md border ${
              result.type === 'success' 
                ? 'bg-green-50 dark:bg-green-900/20 border-green-300 dark:border-green-700' 
                : 'bg-red-50 dark:bg-red-900/20 border-red-300 dark:border-red-700'
            }`}>
              <div className="flex items-start gap-3">
                {result.type === 'success' ? (
                  <svg className="h-5 w-5 text-green-600 dark:text-green-400 mt-0.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                    <polyline points="22 4 12 14.01 9 11.01"></polyline>
                  </svg>
                ) : (
                  <svg className="h-5 w-5 text-red-600 dark:text-red-400 mt-0.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                  </svg>
                )}
                <div>
                  <h3 className={`font-medium ${
                    result.type === 'success' 
                      ? 'text-green-800 dark:text-green-200' 
                      : 'text-red-800 dark:text-red-200'
                  }`}>
                    {result.type === 'success' ? 'Success' : 'Error'}
                  </h3>
                  <p className={`whitespace-pre-line mt-1 ${
                    result.type === 'success' 
                      ? 'text-green-700 dark:text-green-300' 
                      : 'text-red-700 dark:text-red-300'
                  }`}>
                    {result.message}
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
      
      <div className="bg-yellow-100 dark:bg-yellow-900/30 border border-yellow-300 dark:border-yellow-700 rounded-lg p-4">
        <p className="text-yellow-800 dark:text-yellow-200">
          Additional features like SMTP configuration and interval settings will be available here soon.
        </p>
      </div>
    </div>
  );
} 