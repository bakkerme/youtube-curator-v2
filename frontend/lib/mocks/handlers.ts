import { http, HttpResponse } from 'msw'

// Mock data
export const mockChannels = [
  {
    id: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
    title: 'Google Developers',
    customUrl: '@GoogleDevelopers',
    thumbnailUrl: 'https://yt3.ggpht.com/mock-thumbnail',
    createdAt: '2023-01-01T00:00:00Z',
    lastVideoPublishedAt: '2024-01-15T10:00:00Z',
  },
  {
    id: 'UCsBjURrPoezykLs9EqgamOA',
    title: 'Fireship',
    customUrl: '@Fireship',
    thumbnailUrl: 'https://yt3.ggpht.com/mock-thumbnail-2',
    createdAt: '2023-01-02T00:00:00Z',
    lastVideoPublishedAt: '2024-01-14T15:30:00Z',
  },
]

export const mockVideos = [
  {
    id: 'dQw4w9WgXcQ',
    channelId: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
    cachedAt: '2024-01-15T10:00:00Z',
    watched: false,
    title: 'Introduction to React Testing',
    link: {
      Href: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
      Rel: 'alternate',
    },
    published: '2024-01-15T10:00:00Z',
    content: 'Learn how to test React components effectively with modern testing tools.',
    author: {
      name: 'Google Developers',
      uri: 'https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw',
    },
    mediaGroup: {
      mediaThumbnail: {
        URL: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg',
        Width: '1280',
        Height: '720',
      },
      mediaTitle: 'Introduction to React Testing',
      mediaContent: {
        URL: 'https://www.youtube.com/v/dQw4w9WgXcQ?version=3',
        Type: 'application/x-shockwave-flash',
        Width: '640',
        Height: '390',
      },
      mediaDescription: 'Learn how to test React components effectively with modern testing tools.',
    },
  },
  {
    id: 'abc123xyz',
    channelId: 'UCsBjURrPoezykLs9EqgamOA',
    cachedAt: '2024-01-14T15:30:00Z',
    watched: false,
    title: 'Next.js 15 Performance Tips',
    link: {
      Href: 'https://www.youtube.com/watch?v=abc123xyz',
      Rel: 'alternate',
    },
    published: '2024-01-14T15:30:00Z',
    content: 'Discover the latest performance optimizations in Next.js 15.',
    author: {
      name: 'Fireship',
      uri: 'https://www.youtube.com/channel/UCsBjURrPoezykLs9EqgamOA',
    },
    mediaGroup: {
      mediaThumbnail: {
        URL: 'https://i.ytimg.com/vi/abc123xyz/maxresdefault.jpg',
        Width: '1280',
        Height: '720',
      },
      mediaTitle: 'Next.js 15 Performance Tips',
      mediaContent: {
        URL: 'https://www.youtube.com/v/abc123xyz?version=3',
        Type: 'application/x-shockwave-flash',
        Width: '640',
        Height: '390',
      },
      mediaDescription: 'Discover the latest performance optimizations in Next.js 15.',
    },
  },
]

export const mockConfig = {
  smtp: {
    host: 'smtp.gmail.com',
    port: 587,
    username: 'test@example.com',
    password: '',
    fromAddress: 'test@example.com',
    recipientEmail: 'recipient@example.com',
    emailHour: 9,
    emailMinute: 0,
    emailTimezone: 'America/New_York',
  },
}

// Define handlers
export const handlers = [
  // Get all channels
  http.get('/api/channels', () => {
    return HttpResponse.json(mockChannels, {
      headers: {
        'Content-Type': 'application/json',
      },
    })
  }),

  // Add a channel
  http.post('/api/channels', async ({ request }) => {
    const body = await request.json() as { channelId: string }
    const newChannel = {
      id: body.channelId,
      title: 'New Test Channel',
      customUrl: '@newtestchannel',
      thumbnailUrl: 'https://yt3.ggpht.com/new-mock',
      createdAt: new Date().toISOString(),
      lastVideoPublishedAt: new Date().toISOString(),
    }
    return HttpResponse.json(newChannel, { status: 201 })
  }),

  // Delete a channel
  http.delete('/api/channels/:id', () => {
    return new HttpResponse(null, { status: 204 })
  }),

  // Get videos
  http.get('/api/videos', ({ request }) => {
    const url = new URL(request.url)
    const refresh = url.searchParams.get('refresh')
    
    // Simulate refresh behavior
    if (refresh === 'true') {
      const refreshedVideos = {
        videos: [
          {
            ...mockVideos[0],
            title: 'Refreshed: ' + mockVideos[0].title,
            mediaGroup: {
              ...mockVideos[0].mediaGroup,
              mediaTitle: 'Refreshed: ' + mockVideos[0].title,
            }
          },
          ...mockVideos.slice(1)
        ],
        lastRefresh: new Date().toISOString(),
        totalCount: mockVideos.length,
      }
      return HttpResponse.json(refreshedVideos)
    }
    
    return HttpResponse.json({
      videos: mockVideos,
      lastRefresh: '2024-01-15T10:00:00Z',
      totalCount: mockVideos.length,
    })
  }),

  // Get channel by ID
  http.get('/api/channels/search/:id', ({ params }) => {
    const { id } = params
    const channel = mockChannels.find(c => c.id === id)
    if (channel) {
      return HttpResponse.json(channel)
    }
    return new HttpResponse(null, { status: 404 })
  }),

  // Import channels
  http.post('/api/channels/import', async ({ request }) => {
    const body = await request.json() as { channels: Array<{ id: string; title?: string }> }
    return HttpResponse.json({
      imported: body.channels.length,
      skipped: 0,
      errors: [],
    })
  }),

  // Newsletter actions
  http.post('/api/newsletter/run', () => {
    return HttpResponse.json({ message: 'Newsletter job started' })
  }),

  http.post('/api/newsletter/test', () => {
    return HttpResponse.json({ message: 'Test email sent successfully' })
  }),

  // Config endpoints
  http.get('/api/config', () => {
    return HttpResponse.json({ apiUrl: 'http://localhost:8080/api' })
  }),

  http.get('/api/config/smtp', () => {
    return HttpResponse.json(mockConfig.smtp)
  }),

  http.post('/api/config/smtp', async ({ request }) => {
    const body = await request.json() as Record<string, unknown>
    return HttpResponse.json({ ...mockConfig.smtp, ...body })
  }),

  http.post('/api/config/smtp/test', () => {
    return HttpResponse.json({ message: 'Test email sent successfully' })
  }),
] 