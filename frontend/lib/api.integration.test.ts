import { server } from './mocks/server';
import { rest } from 'msw';
import axios from 'axios';

// Create a simple API client for testing that matches MSW routes
const testAPI = axios.create({
  baseURL: '', // MSW will intercept these requests
});

describe('Channel API Integration', () => {
  it('should successfully mock channel API calls', async () => {
    const response = await testAPI.get('/api/channels');

    expect(response.status).toBe(200);
    expect(response.data).toHaveLength(2);
    expect(response.data[0]).toEqual({
      id: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
      title: 'Google Developers',
      customUrl: '@GoogleDevelopers',
      thumbnailUrl: 'https://yt3.ggpht.com/mock-thumbnail',
      createdAt: '2023-01-01T00:00:00Z',
      lastVideoPublishedAt: '2024-01-15T10:00:00Z',
    });
  });

  it('should successfully mock adding a channel', async () => {
    const newChannel = { channelId: 'UCtest123' };
    const response = await testAPI.post('/api/channels', newChannel);

    expect(response.status).toBe(201);
    expect(response.data).toEqual({
      id: 'UCtest123',
      title: 'New Test Channel',
      customUrl: '@newtestchannel',
      thumbnailUrl: 'https://yt3.ggpht.com/new-mock',
      createdAt: expect.any(String),
      lastVideoPublishedAt: expect.any(String),
    });
  });

  it('should successfully mock deleting a channel', async () => {
    const response = await testAPI.delete('/api/channels/UC_x5XG1OV2P6uZZ5FSM9Ttw');

    expect(response.status).toBe(204);
  });

  it('should successfully mock channel search by ID', async () => {
    const channelId = 'UC_x5XG1OV2P6uZZ5FSM9Ttw';
    const response = await testAPI.get(`/api/channels/search/${channelId}`);

    expect(response.status).toBe(200);
    expect(response.data).toEqual({
      id: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
      title: 'Google Developers',
      customUrl: '@GoogleDevelopers',
      thumbnailUrl: 'https://yt3.ggpht.com/mock-thumbnail',
      createdAt: '2023-01-01T00:00:00Z',
      lastVideoPublishedAt: '2024-01-15T10:00:00Z',
    });
  });

  it('should return 404 for non-existent channel search', async () => {
    const nonExistentChannelId = 'UCnonexistent123';
    
    try {
      await testAPI.get(`/api/channels/search/${nonExistentChannelId}`);
      fail('Should have thrown an error for non-existent channel');
    } catch (error: any) {
      expect(error.response.status).toBe(404);
    }
  });

  it('should successfully mock channel import', async () => {
    const importRequest = {
      channels: [
        { id: 'UCimport1', title: 'Import Channel 1' },
        { id: 'UCimport2', title: 'Import Channel 2' },
        { id: 'UCimport3', title: 'Import Channel 3' }
      ]
    };

    const response = await testAPI.post('/api/channels/import', importRequest);

    expect(response.status).toBe(200);
    expect(response.data).toEqual({
      imported: 3,
      skipped: 0,
      errors: [],
    });
  });

  it('should handle empty channel import', async () => {
    const importRequest = {
      channels: []
    };

    const response = await testAPI.post('/api/channels/import', importRequest);

    expect(response.status).toBe(200);
    expect(response.data).toEqual({
      imported: 0,
      skipped: 0,
      errors: [],
    });
  });
});

describe('Video API Integration', () => {
  it('should successfully mock video API calls', async () => {
    const response = await testAPI.get('/api/videos');

    expect(response.status).toBe(200);
    expect(response.data).toHaveLength(2);
    expect(response.data[0]).toEqual({
      id: 'dQw4w9WgXcQ',
      title: 'Introduction to React Testing',
      channelId: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
      channelTitle: 'Google Developers',
      thumbnailUrl: 'https://i.ytimg.com/vi/mock/maxresdefault.jpg',
      publishedAt: '2024-01-15T10:00:00Z',
      viewCount: 150000,
      likeCount: 5000,
      commentCount: 200,
    });
  });

  it('should mock refresh behavior for videos', async () => {
    const response = await testAPI.get('/api/videos?refresh=true');

    expect(response.status).toBe(200);
    expect(response.data).toEqual({
      id: 'dQw4w9WgXcQ',
      title: 'Refreshed: Introduction to React Testing',
      channelId: 'UC_x5XG1OV2P6uZZ5FSM9Ttw',
      channelTitle: 'Google Developers',
      thumbnailUrl: 'https://i.ytimg.com/vi/mock/maxresdefault.jpg',
      publishedAt: '2024-01-15T10:00:00Z',
      viewCount: 150000,
      likeCount: 5000,
      commentCount: 200,
    });
  });
});

describe('Configuration API Integration', () => {
  it('should successfully mock SMTP config retrieval', async () => {
    const response = await testAPI.get('/api/config/smtp');

    expect(response.status).toBe(200);
    expect(response.data).toEqual({
      host: 'smtp.gmail.com',
      port: 587,
      username: 'test@example.com',
      password: '',
      fromAddress: 'test@example.com',
      recipientEmail: 'recipient@example.com',
      emailHour: 9,
      emailMinute: 0,
      emailTimezone: 'America/New_York',
    });
  });

  it('should successfully mock SMTP config update', async () => {
    const configUpdate = {
      server: 'smtp.outlook.com',
      port: '587',
      username: 'newuser@outlook.com',
      password: 'newpassword',
      recipientEmail: 'newrecipient@example.com'
    };

    const response = await testAPI.post('/api/config/smtp', configUpdate);

    expect(response.status).toBe(200);
    expect(response.data).toMatchObject({
      host: 'smtp.gmail.com',
      fromAddress: 'test@example.com',
      emailHour: 9,
      emailMinute: 0,
      emailTimezone: 'America/New_York',
      // The mock merges the request with existing config
      server: 'smtp.outlook.com',
      port: '587',
      username: 'newuser@outlook.com',
      password: 'newpassword',
      recipientEmail: 'newrecipient@example.com'
    });
  });

  it('should successfully mock SMTP test endpoint', async () => {
    const response = await testAPI.post('/api/config/smtp/test');

    expect(response.status).toBe(200);
    expect(response.data).toEqual({
      message: 'Test email sent successfully'
    });
  });

  it('should handle SMTP test with request body', async () => {
    const testRequest = {
      recipientEmail: 'test-recipient@example.com'
    };

    const response = await testAPI.post('/api/config/smtp/test', testRequest);

    expect(response.status).toBe(200);
    expect(response.data).toEqual({
      message: 'Test email sent successfully'
    });
  });
});

describe('Newsletter API Integration', () => {
  it('should successfully mock newsletter run', async () => {
    const response = await testAPI.post('/api/newsletter/run');

    expect(response.status).toBe(200);
    expect(response.data).toEqual({
      message: 'Newsletter job started'
    });
  });

  it('should successfully mock newsletter test', async () => {
    const response = await testAPI.post('/api/newsletter/test');

    expect(response.status).toBe(200);
    expect(response.data).toEqual({
      message: 'Test email sent successfully'
    });
  });
});

describe('Dynamic MSW Handler Override', () => {
  it('should allow runtime handler modification', async () => {
    // Add a custom handler at runtime
    server.use(
      rest.get('/api/channels', (req, res, ctx) => {
        return res(
          ctx.json([
            {
              id: 'CUSTOM123',
              title: 'Custom Test Channel',
              customUrl: '@customtest',
              thumbnailUrl: 'https://custom.example.com/thumb.jpg',
              createdAt: '2024-01-01T00:00:00Z',
              lastVideoPublishedAt: '2024-01-01T12:00:00Z',
            }
          ])
        );
      })
    );

    const response = await testAPI.get('/api/channels');

    expect(response.status).toBe(200);
    expect(response.data).toHaveLength(1);
    expect(response.data[0].title).toBe('Custom Test Channel');
  });
});

describe('Error Handling', () => {
  it('should handle 404 errors appropriately', async () => {
    // Override with error handler
    server.use(
      rest.get('/api/channels/nonexistent', (req, res, ctx) => {
        return res(ctx.status(404), ctx.json({ message: 'Channel not found' }));
      })
    );

    try {
      await testAPI.get('/api/channels/nonexistent');
      fail('Should have thrown an error');
    } catch (error: any) {
      expect(error.response.status).toBe(404);
      expect(error.response.data.message).toBe('Channel not found');
    }
  });
});