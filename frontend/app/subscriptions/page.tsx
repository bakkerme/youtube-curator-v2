'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { channelAPI } from '@/lib/api';
import { ChannelRequest, ImportChannelsRequest, ChannelImport, Channel } from '@/lib/types';
import { X, Plus, Loader2, Upload, FileText } from 'lucide-react';

export default function SubscriptionsPage() {
  const [channelInput, setChannelInput] = useState('');
  const [channelTitle, setChannelTitle] = useState('');
  const [showImport, setShowImport] = useState(false);
  const [importText, setImportText] = useState('');
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

  // Import channels mutation
  const importChannelsMutation = useMutation({
    mutationFn: (request: ImportChannelsRequest) => channelAPI.import(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['channels'] });
      setImportText('');
      setShowImport(false);
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

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target?.result as string;
        setImportText(content);
      };
      reader.readAsText(file);
    }
  };

  // Parse JSON format: [{ "id": "UC...", "title": "..." }]
  const parseImportText = (text: string): ChannelImport[] => {
    if (!text || typeof text !== 'string') {
      return [];
    }
    
    const trimmedText = text.trim();
    
    const parsed = JSON.parse(trimmedText);
    if (Array.isArray(parsed)) {
      return parsed.map((item: Channel): ChannelImport | null => {
        if (typeof item === 'object' && item.id) {
          return {
            url: item.id,
            title: item.title || undefined
          };
        }
        return null;
      }).filter((item): item is ChannelImport => item !== null);
    }

    return [];
  };

  const handleImport = (e: React.FormEvent) => {
    e.preventDefault();
    if (importText.trim()) {
      const channels = parseImportText(importText);
      if (channels.length > 0) {
        importChannelsMutation.mutate({ channels });
      }
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold mb-2">Manage Subscriptions</h1>
      <p className="text-gray-600 dark:text-gray-400 mb-8">
        Add or remove YouTube channels to track their latest uploads.
      </p>

      {/* Add Single Channel Form */}
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
      </form>

      {/* Import Section */}
      <div className="mb-8">
        <div className="flex items-center justify-between mb-4">
          <p>Or import channels</p>
          <button
            onClick={() => setShowImport(!showImport)}
            className="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 
                     transition-colors flex items-center gap-2"
          >
            <Upload className="w-4 h-4" />
            {showImport ? 'Hide Import' : 'Import Channels'}
          </button>
        </div>

        {showImport && (
          <div className="bg-gray-50 dark:bg-gray-800 p-6 rounded-lg border border-gray-200 dark:border-gray-700">
            <form onSubmit={handleImport} className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-2">
                  Upload File or Paste Channel Data
                </label>
                <div className="space-y-4">
                  {/* File Upload */}
                  <div className="flex items-center gap-4">
                    <input
                      type="file"
                      accept=".json,.txt"
                      onChange={handleFileUpload}
                      className="text-sm text-gray-500 file:mr-4 file:py-2 file:px-4
                               file:rounded-lg file:border-0 file:text-sm file:font-semibold
                               file:bg-red-50 file:text-red-700 hover:file:bg-red-100
                               dark:file:bg-red-900/30 dark:file:text-red-300"
                      disabled={importChannelsMutation.isPending}
                    />
                  </div>

                  {/* Text Area */}
                  <textarea
                    value={importText}
                    onChange={(e) => setImportText(e.target.value)}
                    placeholder="Paste JSON array or channel URLs here.&#10;&#10;JSON format:&#10;[&#10;  { &quot;id&quot;: &quot;UCAYF6ZY9gWBR1GW3R7PX7yw&quot;, &quot;title&quot;: &quot;Channel Name&quot; },&#10;  { &quot;id&quot;: &quot;UC456...&quot;, &quot;title&quot;: &quot;Another Channel&quot; }&#10;]&#10;&#10;Or simple format (one per line):&#10;https://www.youtube.com/channel/UC123...&#10;UC456...,Channel Name&#10;@username"
                    className="w-full h-40 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                             bg-white dark:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-red-500
                             font-mono text-sm"
                    disabled={importChannelsMutation.isPending}
                  />
                </div>
              </div>

              <div className="flex items-center gap-4">
                <button
                  type="submit"
                  disabled={importChannelsMutation.isPending || !importText.trim()}
                  className="px-6 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 
                           disabled:opacity-50 disabled:cursor-not-allowed transition-colors
                           flex items-center gap-2"
                >
                  {importChannelsMutation.isPending ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Importing...
                    </>
                  ) : (
                    <>
                      <FileText className="w-4 h-4" />
                      Import Channels
                    </>
                  )}
                </button>
                
                {importText && (
                  <span className="text-sm text-gray-500">
                    {parseImportText(importText.trim()).length} channels detected
                  </span>
                )}
              </div>
            </form>

            <div className="mt-4 p-4 bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-700 rounded-lg">
              <h4 className="font-medium text-blue-900 dark:text-blue-100 mb-2">Import Formats:</h4>
              <ul className="text-sm text-blue-800 dark:text-blue-200 space-y-1">
                <li><strong>JSON format:</strong> <code>[{`{&quot;id&quot;: &quot;UC...&quot;, &quot;title&quot;: &quot;Channel Name&quot;}`}]</code></li>
              </ul>
            </div>
          </div>
        )}
      </div>

      {/* Success Messages */}
      {addChannelMutation.isSuccess && addChannelMutation.data && (
        <div role="alert" className="mb-4 p-4 bg-green-100 dark:bg-green-900/30 border border-green-300 dark:border-green-700 rounded-lg text-green-700 dark:text-green-300">
          Successfully added channel: {addChannelMutation.data.title}
        </div>
      )}

      {/* Error Messages */}
      {addChannelMutation.isError && (
        <div className="mb-4 p-4 bg-red-100 dark:bg-red-900/30 border border-red-300 dark:border-red-700 rounded-lg text-red-700 dark:text-red-300">
          {addChannelMutation.error?.message || 'Failed to add channel'}
        </div>
      )}

      {importChannelsMutation.isError && (
        <div className="mb-4 p-4 bg-red-100 dark:bg-red-900/30 border border-red-300 dark:border-red-700 rounded-lg text-red-700 dark:text-red-300">
          {importChannelsMutation.error?.message || 'Failed to import channels'}
        </div>
      )}

      {/* Import Results */}
      {importChannelsMutation.isSuccess && importChannelsMutation.data && (
        <div className="mb-4 space-y-2">
          {importChannelsMutation.data.imported?.length > 0 && (
            <div className="p-4 bg-green-100 dark:bg-green-900/30 border border-green-300 dark:border-green-700 rounded-lg text-green-700 dark:text-green-300">
              Successfully imported {importChannelsMutation.data.imported.length} channels
            </div>
          )}
          {importChannelsMutation.data.failed?.length > 0 && (
            <div className="p-4 bg-yellow-100 dark:bg-yellow-900/30 border border-yellow-300 dark:border-yellow-700 rounded-lg">
              <p className="text-yellow-800 dark:text-yellow-200 font-medium mb-2">
                {importChannelsMutation.data.failed.length} channels failed to import:
              </p>
              <ul className="text-sm text-yellow-700 dark:text-yellow-300 space-y-1">
                {importChannelsMutation.data.failed.map((failure, index) => (
                  <li key={index}>
                    <span className="font-mono">{failure.channel.url}</span>: {failure.error}
                  </li>
                ))}
              </ul>
            </div>
          )}
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