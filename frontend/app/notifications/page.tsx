'use client';

import { useState, useEffect } from 'react';
import { newsletterAPI, channelAPI, configAPI } from '@/lib/api';
import { Channel, SMTPConfigRequest, LLMConfigRequest, RunNewsletterRequest, NewsletterConfigRequest } from '@/lib/types';

export default function NotificationsPage() {
  const [channels, setChannels] = useState<Channel[]>([]);
  const [selectedChannel, setSelectedChannel] = useState<string>('all');
  const [ignoreLastChecked, setIgnoreLastChecked] = useState<boolean>(false);
  const [maxItems, setMaxItems] = useState<number>(0);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingChannels, setIsLoadingChannels] = useState(true);
  const [result, setResult] = useState<{ type: 'success' | 'error', message: string } | null>(null);
  
  // SMTP Configuration State
  const [smtpConfig, setSMTPConfig] = useState<SMTPConfigRequest>({
    server: '',
    port: '',
    username: '',
    password: '',
    recipientEmail: ''
  });
  const [isLoadingSMTP, setIsLoadingSMTP] = useState(true);
  const [isSavingSMTP, setIsSavingSMTP] = useState(false);
  const [smtpResult, setSMTPResult] = useState<{ type: 'success' | 'error', message: string } | null>(null);
  const [smtpPasswordSet, setSMTPPasswordSet] = useState(false);

  // LLM Configuration State
  const [llmConfig, setLLMConfig] = useState<LLMConfigRequest>({
    endpoint: '',
    apiKey: '',
    model: ''
  });
  const [isLoadingLLM, setIsLoadingLLM] = useState(true);
  const [isSavingLLM, setIsSavingLLM] = useState(false);
  const [llmResult, setLLMResult] = useState<{ type: 'success' | 'error', message: string } | null>(null);
  const [llmApiKeySet, setLLMApiKeySet] = useState(false);

  // Newsletter Configuration State
  const [newsletterConfig, setNewsletterConfig] = useState<NewsletterConfigRequest>({
    enabled: true
  });
  const [isLoadingNewsletter, setIsLoadingNewsletter] = useState(true);
  const [isSavingNewsletter, setIsSavingNewsletter] = useState(false);
  const [newsletterResult, setNewsletterResult] = useState<{ type: 'success' | 'error', message: string } | null>(null);

  useEffect(() => {
    loadChannels();
    loadSMTPConfig();
    loadLLMConfig();
    loadNewsletterConfig();
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

  const loadSMTPConfig = async () => {
    try {
      const data = await configAPI.getSMTP();
      setSMTPConfig({
        server: data.server || '',
        port: data.port || '',
        username: data.username || '',
        password: '', // Password is never returned from API
        recipientEmail: data.recipientEmail || ''
      });
      setSMTPPasswordSet(data.passwordSet || false);
    } catch (error) {
      console.error('Failed to load SMTP configuration:', error);
    } finally {
      setIsLoadingSMTP(false);
    }
  };

  const loadLLMConfig = async () => {
    try {
      const data = await configAPI.getLLM();
      setLLMConfig({
        endpoint: data.endpointUrl || '',
        apiKey: '', // API key is never returned from API
        model: data.model || ''
      });
      setLLMApiKeySet(data.apiKeySet || false);
    } catch (error) {
      console.error('Failed to load LLM configuration:', error);
    } finally {
      setIsLoadingLLM(false);
    }
  };

  const loadNewsletterConfig = async () => {
    try {
      const data = await configAPI.getNewsletter();
      setNewsletterConfig({
        enabled: data.enabled
      });
    } catch (error) {
      console.error('Failed to load newsletter configuration:', error);
    } finally {
      setIsLoadingNewsletter(false);
    }
  };

  const handleRunNewsletter = async () => {
    setIsLoading(true);
    setResult(null);

    try {
      const request: RunNewsletterRequest = {};
      if (selectedChannel !== 'all') {
        request.channelId = selectedChannel;
      }
      if (ignoreLastChecked) {
        request.ignoreLastChecked = true;
      }
      if (maxItems > 0) {
        request.maxItems = maxItems;
      }
      
      const response = await newsletterAPI.run(request);
      
      const successMessage = `Newsletter run completed successfully!\nProcessed ${response.channelsProcessed} channel(s), found ${response.newVideosFound} new video(s).\n${response.emailSent ? 'Email sent.' : 'No email sent (no new videos).'}`;
      
      setResult({ type: 'success', message: successMessage });
    } catch (error) {
      setResult({ 
        type: 'error', 
        message: error instanceof Error ? error.message : 'Failed to run newsletter' 
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleSaveSMTP = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSavingSMTP(true);
    setSMTPResult(null);

    try {
      await configAPI.setSMTP(smtpConfig);
      setSMTPResult({ type: 'success', message: 'SMTP configuration saved successfully!' });
      setSMTPPasswordSet(true); // Password is now set
    } catch (error) {
      setSMTPResult({ 
        type: 'error', 
        message: error instanceof Error ? error.message : 'Failed to save SMTP configuration' 
      });
    } finally {
      setIsSavingSMTP(false);
    }
  };

  const handleSaveLLM = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSavingLLM(true);
    setLLMResult(null);

    try {
      await configAPI.setLLM(llmConfig);
      setLLMResult({ type: 'success', message: 'LLM configuration saved successfully!' });
      setLLMApiKeySet(true); // API key is now set
    } catch (error) {
      setLLMResult({ 
        type: 'error', 
        message: error instanceof Error ? error.message : 'Failed to save LLM configuration' 
      });
    } finally {
      setIsSavingLLM(false);
    }
  };

  const handleSaveNewsletter = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSavingNewsletter(true);
    setNewsletterResult(null);

    try {
      await configAPI.setNewsletter(newsletterConfig);
      setNewsletterResult({ 
        type: 'success', 
        message: `Newsletter ${newsletterConfig.enabled ? 'enabled' : 'disabled'} successfully!` 
      });
    } catch (error) {
      setNewsletterResult({ 
        type: 'error', 
        message: error instanceof Error ? error.message : 'Failed to save newsletter configuration' 
      });
    } finally {
      setIsSavingNewsletter(false);
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
            <div className="w-32">
              <label htmlFor="max-items" className="text-sm font-medium mb-2 block">
                Max Items
              </label>
              <input
                id="max-items"
                type="number"
                min="1"
                value={maxItems}
                onChange={(e) => setMaxItems(parseInt(e.target.value) || 0)}
                disabled={isLoading || isLoadingChannels}
                placeholder="0 (all)"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
              />
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
          <div className="flex items-center gap-2">
            <input
              id="ignore-last-checked"
              type="checkbox"
              checked={ignoreLastChecked}
              onChange={(e) => setIgnoreLastChecked(e.target.checked)}
              disabled={isLoading || isLoadingChannels}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded disabled:opacity-50 disabled:cursor-not-allowed"
            />
            <label htmlFor="ignore-last-checked" className="text-sm font-medium">
              Ignore last checked
            </label>
          </div>

          <div className="text-sm text-gray-600 dark:text-gray-400 space-y-2">
            <p><strong>Ignore last checked:</strong> When enabled, this will process all videos in the RSS feed regardless of when they were last checked. This is useful for debugging and testing to see all available videos.</p>
            <p><strong>Max Items:</strong> Limits the number of new videos to process per channel. Set to 0 to process all new videos. This can help reduce processing time for channels with many videos.</p>
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
      
      {/* Newsletter Configuration */}
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm mb-6">
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-xl font-semibold mb-2">Newsletter Schedule</h2>
          <p className="text-gray-600 dark:text-gray-400">
            Control whether the automatic newsletter scheduler runs. When disabled, the scheduled 
            checks for new videos are paused, but you can still use the manual trigger above.
          </p>
        </div>
        <div className="p-6">
          {isLoadingNewsletter ? (
            <div className="flex justify-center py-8">
              <svg className="animate-spin h-8 w-8 text-gray-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
            </div>
          ) : (
            <form onSubmit={handleSaveNewsletter} className="space-y-6">
              <div className="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                <div className="flex-1">
                  <div className="flex items-center gap-3">
                    <div className={`w-3 h-3 rounded-full ${newsletterConfig.enabled ? 'bg-green-500' : 'bg-red-500'}`}></div>
                    <h3 className="text-lg font-medium">
                      Newsletter Scheduler is {newsletterConfig.enabled ? 'Enabled' : 'Disabled'}
                    </h3>
                  </div>
                  <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                    {newsletterConfig.enabled 
                      ? 'The system will automatically check for new videos on schedule and send email notifications.'
                      : 'Automatic checks are paused. The manual newsletter trigger above will still work.'
                    }
                  </p>
                </div>
                <div className="flex items-center gap-4">
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={newsletterConfig.enabled}
                      onChange={(e) => setNewsletterConfig({ enabled: e.target.checked })}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                  </label>
                </div>
              </div>

              {!newsletterConfig.enabled && (
                <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-300 dark:border-yellow-700 rounded-lg">
                  <div className="flex items-start gap-3">
                    <svg className="h-5 w-5 text-yellow-600 dark:text-yellow-400 mt-0.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                      <line x1="12" y1="9" x2="12" y2="13"></line>
                      <line x1="12" y1="17" x2="12.01" y2="17"></line>
                    </svg>
                    <div>
                      <h4 className="font-medium text-yellow-800 dark:text-yellow-200">Newsletter Scheduler Disabled</h4>
                      <p className="text-yellow-700 dark:text-yellow-300 text-sm mt-1">
                        You won&apos;t receive automatic email notifications when new videos are published. 
                        You can still use the manual newsletter trigger above and configure SMTP settings below.
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {newsletterResult && (
                <div className={`p-4 rounded-md border ${
                  newsletterResult.type === 'success' 
                    ? 'bg-green-50 dark:bg-green-900/20 border-green-300 dark:border-green-700' 
                    : 'bg-red-50 dark:bg-red-900/20 border-red-300 dark:border-red-700'
                }`}>
                  <div className="flex items-start gap-3">
                    {newsletterResult.type === 'success' ? (
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
                      <p className={`${
                        newsletterResult.type === 'success' 
                          ? 'text-green-700 dark:text-green-300' 
                          : 'text-red-700 dark:text-red-300'
                      }`}>
                        {newsletterResult.message}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              <button
                type="submit"
                disabled={isSavingNewsletter}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
              >
                {isSavingNewsletter ? (
                  <>
                    <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Saving...
                  </>
                ) : (
                  'Save Configuration'
                )}
              </button>
            </form>
          )}
        </div>
      </div>
      
      {/* SMTP Configuration */}
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm mb-6">
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-xl font-semibold mb-2">SMTP Configuration</h2>
          <p className="text-gray-600 dark:text-gray-400">
            Configure your email server settings for sending newsletters.
          </p>
        </div>
        <div className="p-6">
          {isLoadingSMTP ? (
            <div className="flex justify-center py-8">
              <svg className="animate-spin h-8 w-8 text-gray-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
            </div>
          ) : (
            <form onSubmit={handleSaveSMTP} className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label htmlFor="smtp-server" className="block text-sm font-medium mb-2">
                    SMTP Server
                  </label>
                  <input
                    id="smtp-server"
                    type="text"
                    value={smtpConfig.server}
                    onChange={(e) => setSMTPConfig({ ...smtpConfig, server: e.target.value })}
                    placeholder="smtp.gmail.com"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    required
                  />
                </div>
                
                <div>
                  <label htmlFor="smtp-port" className="block text-sm font-medium mb-2">
                    SMTP Port
                  </label>
                  <input
                    id="smtp-port"
                    type="text"
                    value={smtpConfig.port}
                    onChange={(e) => setSMTPConfig({ ...smtpConfig, port: e.target.value })}
                    placeholder="587"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    required
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label htmlFor="smtp-username" className="block text-sm font-medium mb-2">
                    Username
                  </label>
                  <input
                    id="smtp-username"
                    type="text"
                    value={smtpConfig.username}
                    onChange={(e) => setSMTPConfig({ ...smtpConfig, username: e.target.value })}
                    placeholder="your-email@example.com"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    required
                  />
                </div>
                
                <div>
                  <label htmlFor="smtp-password" className="block text-sm font-medium mb-2">
                    Password
                  </label>
                  <input
                    id="smtp-password"
                    type="password"
                    value={smtpConfig.password}
                    onChange={(e) => setSMTPConfig({ ...smtpConfig, password: e.target.value })}
                    placeholder={smtpPasswordSet ? "••••••••" : "Enter password"}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    required
                  />
                  <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    {smtpPasswordSet ? "Password is already configured. Leave blank to keep current password." : "For Gmail, use an app-specific password"}
                  </p>
                </div>
              </div>

              <div>
                <label htmlFor="recipient-email" className="block text-sm font-medium mb-2">
                  Recipient Email
                </label>
                <input
                  id="recipient-email"
                  type="email"
                  value={smtpConfig.recipientEmail}
                  onChange={(e) => setSMTPConfig({ ...smtpConfig, recipientEmail: e.target.value })}
                  placeholder="recipient@example.com"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  required
                />
                <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  Email address where newsletters will be sent
                </p>
              </div>

              {smtpResult && (
                <div className={`p-4 rounded-md border ${
                  smtpResult.type === 'success' 
                    ? 'bg-green-50 dark:bg-green-900/20 border-green-300 dark:border-green-700' 
                    : 'bg-red-50 dark:bg-red-900/20 border-red-300 dark:border-red-700'
                }`}>
                  <div className="flex items-start gap-3">
                    {smtpResult.type === 'success' ? (
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
                      <p className={`${
                        smtpResult.type === 'success' 
                          ? 'text-green-700 dark:text-green-300' 
                          : 'text-red-700 dark:text-red-300'
                      }`}>
                        {smtpResult.message}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              <button
                type="submit"
                disabled={isSavingSMTP}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
              >
                {isSavingSMTP ? (
                  <>
                    <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Saving...
                  </>
                ) : (
                  'Save Configuration'
                )}
              </button>
            </form>
          )}
        </div>
      </div>

      {/* LLM Configuration */}
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm mb-6">
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-xl font-semibold mb-2">LLM Configuration</h2>
          <p className="text-gray-600 dark:text-gray-400">
            Configure your OpenAI-compatible LLM API settings for video summaries and analysis.
          </p>
        </div>
        <div className="p-6">
          {isLoadingLLM ? (
            <div className="flex justify-center py-8">
              <svg className="animate-spin h-8 w-8 text-gray-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
            </div>
          ) : (
            <form onSubmit={handleSaveLLM} className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label htmlFor="llm-endpoint" className="block text-sm font-medium mb-2">
                    API Endpoint
                  </label>
                  <input
                    id="llm-endpoint"
                    type="url"
                    value={llmConfig.endpoint}
                    onChange={(e) => setLLMConfig({ ...llmConfig, endpoint: e.target.value })}
                    placeholder="https://api.openai.com/v1"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    required
                  />
                  <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    OpenAI API or compatible endpoint (e.g., local LLM server)
                  </p>
                </div>
                
                <div>
                  <label htmlFor="llm-model" className="block text-sm font-medium mb-2">
                    Model
                  </label>
                  <input
                    id="llm-model"
                    type="text"
                    value={llmConfig.model}
                    onChange={(e) => setLLMConfig({ ...llmConfig, model: e.target.value })}
                    placeholder="gpt-3.5-turbo"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    required
                  />
                  <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    Model name (varies by provider)
                  </p>
                </div>
              </div>

              <div>
                <label htmlFor="llm-api-key" className="block text-sm font-medium mb-2">
                  API Key
                </label>
                <input
                  id="llm-api-key"
                  type="password"
                  value={llmConfig.apiKey}
                  onChange={(e) => setLLMConfig({ ...llmConfig, apiKey: e.target.value })}
                  placeholder={llmApiKeySet ? "••••••••" : "sk-..."}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  required
                />
                <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {llmApiKeySet ? "API key is already configured. Leave blank to keep current key." : "Your API key for the LLM service"}
                </p>
              </div>

              {llmResult && (
                <div className={`p-4 rounded-md border ${
                  llmResult.type === 'success' 
                    ? 'bg-green-50 dark:bg-green-900/20 border-green-300 dark:border-green-700' 
                    : 'bg-red-50 dark:bg-red-900/20 border-red-300 dark:border-red-700'
                }`}>
                  <div className="flex items-start gap-3">
                    {llmResult.type === 'success' ? (
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
                      <p className={`${
                        llmResult.type === 'success' 
                          ? 'text-green-700 dark:text-green-300' 
                          : 'text-red-700 dark:text-red-300'
                      }`}>
                        {llmResult.message}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              <button
                type="submit"
                disabled={isSavingLLM}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
              >
                {isSavingLLM ? (
                  <>
                    <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Saving...
                  </>
                ) : (
                  'Save Configuration'
                )}
              </button>
            </form>
          )}
        </div>
      </div>
      
      <div className="mt-6 bg-yellow-100 dark:bg-yellow-900/30 border border-yellow-300 dark:border-yellow-700 rounded-lg p-4">
        <p className="text-yellow-800 dark:text-yellow-200">
          Additional features like check interval settings will be available here soon.
        </p>
      </div>
    </div>
  );
}