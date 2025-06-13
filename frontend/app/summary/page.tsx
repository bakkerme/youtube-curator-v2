'use client';

import { useState } from 'react';
import { videoAPI, extractVideoId } from '@/lib/api';
import { Loader2, Video, AlertCircle, ChevronDown, ChevronRight, Brain } from 'lucide-react';
import ReactMarkdown from 'react-markdown';

export default function SummaryPage() {
  const [videoInput, setVideoInput] = useState('');
  const [summary, setSummary] = useState('');
  const [thinking, setThinking] = useState('');
  const [showThinking, setShowThinking] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!videoInput.trim()) {
      setError('Please enter a video ID or URL');
      return;
    }

    setLoading(true);
    setError('');
    setSummary('');
    setThinking('');
    setShowThinking(false);

    try {
      const videoId = extractVideoId(videoInput);
      const response = await videoAPI.getSummary(videoId);
      setSummary(response.summary);
      setThinking(response.thinking || '');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to get video summary');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          Video Summary
        </h1>
        <p className="text-gray-600 dark:text-gray-300">
          Generate AI-powered summaries for YouTube videos
        </p>
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="video-input" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              YouTube Video ID or URL
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <Video className="h-5 w-5 text-gray-400" />
              </div>
              <input
                id="video-input"
                type="text"
                value={videoInput}
                onChange={(e) => setVideoInput(e.target.value)}
                placeholder="e.g., dQw4w9WgXcQ or https://www.youtube.com/watch?v=dQw4w9WgXcQ or https://www.youtube.com/shorts/DM-pfKGioWw"
                className="block w-full pl-10 pr-3 py-3 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-red-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                disabled={loading}
              />
            </div>
            <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
              Enter either a YouTube video ID (11 characters) or a full YouTube URL (including Shorts)
            </p>
          </div>

          <button
            type="submit"
            disabled={loading || !videoInput.trim()}
            className="w-full flex items-center justify-center px-4 py-3 border border-transparent text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Generating Summary...
              </>
            ) : (
              'Generate Summary'
            )}
          </button>
        </form>

        {error && (
          <div className="mt-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
            <div className="flex items-center">
              <AlertCircle className="h-5 w-5 text-red-400 mr-2" />
              <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
            </div>
          </div>
        )}

        {summary && (
          <div className="mt-6 space-y-4">
            {/* Thinking Accordion */}
            {thinking && (
              <div className="border border-gray-200 dark:border-gray-600 rounded-md">
                <button
                  onClick={() => setShowThinking(!showThinking)}
                  className="w-full flex items-center justify-between p-4 text-left hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
                >
                  <div className="flex items-center space-x-2">
                    <Brain className="h-5 w-5 text-purple-500" />
                    <span className="font-medium text-gray-900 dark:text-white">
                      LLM Thinking Process
                    </span>
                  </div>
                  {showThinking ? (
                    <ChevronDown className="h-4 w-4 text-gray-500" />
                  ) : (
                    <ChevronRight className="h-4 w-4 text-gray-500" />
                  )}
                </button>
                
                {showThinking && (
                  <div className="px-4 pb-4 border-t border-gray-200 dark:border-gray-600">
                    <div className="bg-purple-50 dark:bg-purple-900/20 rounded-md p-4 mt-3">
                      <div className="text-sm text-purple-700 dark:text-purple-300 prose prose-sm dark:prose-invert max-w-none">
                        <ReactMarkdown>{thinking}</ReactMarkdown>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Summary */}
            <div>
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                Summary
              </h2>
              <div className="bg-gray-50 dark:bg-gray-900 rounded-md p-4 border border-gray-200 dark:border-gray-600">
                <div className="text-gray-700 dark:text-gray-300 prose prose-sm dark:prose-invert max-w-none">
                  <ReactMarkdown>{summary}</ReactMarkdown>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}