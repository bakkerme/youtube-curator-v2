import { rest } from 'msw'

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
    title: 'Introduction to React Testing',
    channelId: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
    channelTitle: 'Google Developers',
    thumbnailUrl: 'https://i.ytimg.com/vi/mock/maxresdefault.jpg',
    publishedAt: '2024-01-15T10:00:00Z',
    viewCount: 150000,
    likeCount: 5000,
    commentCount: 200,
  },
  {
    id: 'abc123xyz',
    title: 'Next.js 15 Performance Tips',
    channelId: 'UCsBjURrPoezykLs9EqgamOA',
    channelTitle: 'Fireship',
    thumbnailUrl: 'https://i.ytimg.com/vi/mock2/maxresdefault.jpg',
    publishedAt: '2024-01-14T15:30:00Z',
    viewCount: 250000,
    likeCount: 12000,
    commentCount: 450,
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
  rest.get('/api/channels', (req, res, ctx) => {
    return res(ctx.json(mockChannels))
  }),

  // Add a channel
  rest.post('/api/channels', async (req, res, ctx) => {
    const body = await req.json() as { channelId: string }
    const newChannel = {
      id: body.channelId,
      title: 'New Test Channel',
      customUrl: '@newtestchannel',
      thumbnailUrl: 'https://yt3.ggpht.com/new-mock',
      createdAt: new Date().toISOString(),
      lastVideoPublishedAt: new Date().toISOString(),
    }
    return res(ctx.status(201), ctx.json(newChannel))
  }),

  // Delete a channel
  rest.delete('/api/channels/:id', (req, res, ctx) => {
    return res(ctx.status(204))
  }),

  // Get videos
  rest.get('/api/videos', (req, res, ctx) => {
    const url = new URL(req.url)
    const refresh = url.searchParams.get('refresh')
    
    // Simulate refresh behavior
    if (refresh === 'true') {
      return res(ctx.json({
        ...mockVideos[0],
        title: 'Refreshed: ' + mockVideos[0].title,
      }))
    }
    
    return res(ctx.json(mockVideos))
  }),

  // Get channel by ID
  rest.get('/api/channels/search/:id', (req, res, ctx) => {
    const { id } = req.params
    const channel = mockChannels.find(c => c.id === id)
    if (channel) {
      return res(ctx.json(channel))
    }
    return res(ctx.status(404))
  }),

  // Import channels
  rest.post('/api/channels/import', async (req, res, ctx) => {
    const body = await req.json() as { channels: any[] }
    return res(ctx.json({
      imported: body.channels.length,
      skipped: 0,
      errors: [],
    }))
  }),

  // Newsletter actions
  rest.post('/api/newsletter/run', (req, res, ctx) => {
    return res(ctx.json({ message: 'Newsletter job started' }))
  }),

  rest.post('/api/newsletter/test', (req, res, ctx) => {
    return res(ctx.json({ message: 'Test email sent successfully' }))
  }),

  // Config endpoints
  rest.get('/api/config/smtp', (req, res, ctx) => {
    return res(ctx.json(mockConfig.smtp))
  }),

  rest.post('/api/config/smtp', async (req, res, ctx) => {
    const body = await req.json() as Record<string, any>
    return res(ctx.json({ ...mockConfig.smtp, ...body }))
  }),

  rest.post('/api/config/smtp/test', (req, res, ctx) => {
    return res(ctx.json({ message: 'Test email sent successfully' }))
  }),
] 