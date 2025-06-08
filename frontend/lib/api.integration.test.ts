import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { videoAPI } from './api';
import { handlers } from './mocks/handlers';

// Helper function to make API calls using fetch (which works correctly with MSW v2)
const apiClient = {
  async get(url: string) {
    const response = await fetch(url);

    let data;
    try {
      data = await response.json();
    } catch {
      data = null; // For responses without JSON body (404, 500, etc.)
    }

    return { status: response.status, data };
  },

  async post(url: string, body?: Record<string, unknown>) {
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: body ? JSON.stringify(body) : undefined,
    });

    let data;
    try {
      data = await response.json();
    } catch {
      data = null; // For 204 responses or empty responses
    }

    return { status: response.status, data };
  },

  async delete(url: string) {
    const response = await fetch(url, { method: 'DELETE' });

    let data;
    try {
      data = await response.json();
    } catch {
      data = null; // For 204 responses or empty responses
    }

    return { status: response.status, data };
  },
};

// Setup MSW server for API mocking
const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());


describe('API Integration Tests', () => {
  describe('Channels API', () => {
    test('GET /api/channels returns channel list', async () => {
      // Use fetch since it works correctly with MSW v2
      const response = await fetch('/api/channels');
      const data = await response.json();

      expect(response.status).toBe(200);
      expect(data).toHaveProperty('channels');
      expect(data).toHaveProperty('totalCount');
      expect(data).toHaveProperty('lastUpdated');
      expect(data.channels).toHaveLength(2);
      expect(data.channels[0]).toHaveProperty('id', 'UC_x5XG1OV2P6uZZ5FSM9Ttw');
      expect(data.channels[0]).toHaveProperty('title', 'Google Developers');
    });

    test('POST /api/channels creates new channel', async () => {
      const newChannelData = { channelId: 'UC123newchannel' };

      const response = await apiClient.post('/api/channels', newChannelData);

      expect(response.status).toBe(201);
      expect(response.data).toHaveProperty('id', 'UC123newchannel');
      expect(response.data).toHaveProperty('title', 'New Test Channel');
    });

    test('DELETE /api/channels/:id deletes channel', async () => {
      const response = await apiClient.delete('/api/channels/UC_x5XG1OV2P6uZZ5FSM9Ttw');

      expect(response.status).toBe(204);
    });

    test('GET /api/channels/search/:id returns specific channel', async () => {
      const response = await apiClient.get('/api/channels/search/UC_x5XG1OV2P6uZZ5FSM9Ttw');

      expect(response.status).toBe(200);
      expect(response.data).toHaveProperty('id', 'UC_x5XG1OV2P6uZZ5FSM9Ttw');
      expect(response.data).toHaveProperty('title', 'Google Developers');
    });

    test('GET /api/channels/search/:id returns 404 for unknown channel', async () => {
      const response = await apiClient.get('/api/channels/search/unknown-id');
      expect(response.status).toBe(404);
    });

    test('POST /api/channels/import imports channels', async () => {
      const importData = { channels: [{ id: '1' }, { id: '2' }] };

      const response = await apiClient.post('/api/channels/import', importData);

      expect(response.status).toBe(200);
      expect(response.data).toEqual({
        imported: 2,
        skipped: 0,
        errors: [],
      });
    });
  });

  describe('Videos API', () => {
    test('GET /api/videos returns video list', async () => {
      const response = await apiClient.get('/api/videos');

      expect(response.status).toBe(200);
      expect(response.data).toHaveProperty('videos');
      expect(response.data).toHaveProperty('lastRefresh');
      expect(response.data).toHaveProperty('totalCount');
      expect(response.data.videos).toHaveLength(2);
      expect(response.data.videos[0]).toHaveProperty('id', 'dQw4w9WgXcQ');
      expect(response.data.videos[0]).toHaveProperty('title', 'Introduction to React Testing');
    });

    test('GET /api/videos with refresh param returns refreshed data', async () => {
      const response = await apiClient.get('/api/videos?refresh=true');

      expect(response.status).toBe(200);
      expect(response.data).toHaveProperty('videos');
      expect(response.data.videos[0]).toHaveProperty('title');
      expect(response.data.videos[0].title).toContain('Refreshed: Introduction to React Testing');
    });

    describe('markAsWatched', () => {
      it('should successfully mark a video as watched', async () => {
        // Arrange
        const videoId = 'test-video-id';

        server.use(
          http.post(`http://localhost:8080/api/videos/${videoId}/watch`, () => {
            return new HttpResponse(null, { status: 200 });
          })
        );

        // Act & Assert
        await expect(videoAPI.markAsWatched(videoId)).resolves.not.toThrow();
      });

      it('should handle API errors when marking video as watched', async () => {
        // Arrange
        const videoId = 'test-video-id';
        const errorMessage = 'Video not found';

        server.use(
          http.post(`http://localhost:8080/api/videos/${videoId}/watch`, () => {
            return HttpResponse.json(
              { message: errorMessage },
              { status: 404 }
            );
          })
        );

        // Act & Assert
        await expect(videoAPI.markAsWatched(videoId)).rejects.toThrow();
      });

      it('should handle network errors', async () => {
        // Arrange
        const videoId = 'test-video-id';

        server.use(
          http.post(`http://localhost:8080/api/videos/${videoId}/watch`, () => {
            return HttpResponse.error();
          })
        );

        // Act & Assert
        await expect(videoAPI.markAsWatched(videoId)).rejects.toThrow('Unable to connect to the server');
      });
    });
  });

  describe('Newsletter API', () => {
    test('POST /api/newsletter/run starts newsletter job', async () => {
      const response = await apiClient.post('/api/newsletter/run');

      expect(response.status).toBe(200);
      expect(response.data).toEqual({ message: 'Newsletter job started' });
    });

    test('POST /api/newsletter/test sends test email', async () => {
      const response = await apiClient.post('/api/newsletter/test');

      expect(response.status).toBe(200);
      expect(response.data).toEqual({ message: 'Test email sent successfully' });
    });
  });

  describe('Config API', () => {
    test('GET /api/config/smtp returns SMTP config', async () => {
      const response = await apiClient.get('/api/config/smtp');

      expect(response.status).toBe(200);
      expect(response.data).toHaveProperty('host', 'smtp.gmail.com');
      expect(response.data).toHaveProperty('port', 587);
    });

    test('POST /api/config/smtp updates SMTP config', async () => {
      const updateData = { host: 'smtp.new.com', port: 465 };

      const response = await apiClient.post('/api/config/smtp', updateData);

      expect(response.status).toBe(200);
      expect(response.data).toHaveProperty('host', 'smtp.new.com');
      expect(response.data).toHaveProperty('port', 465);
    });

    test('POST /api/config/smtp/test sends test SMTP email', async () => {
      const response = await apiClient.post('/api/config/smtp/test');

      expect(response.status).toBe(200);
      expect(response.data).toEqual({ message: 'Test email sent successfully' });
    });
  });
});

describe('Error Handling', () => {
  test.skip('handles network errors gracefully', async () => {
    // This test demonstrates error handling when no MSW handler matches
    // Skip due to MSW + JSDOM compatibility issues
    try {
      await apiClient.get('/api/nonexistent-endpoint');
      fail('Should have thrown an error');
    } catch (error: unknown) {
      // With fetch, network errors will be different than axios
      expect(error).toBeDefined();
    }
  });
});

describe('Advanced API Patterns', () => {
  test('handles concurrent requests', async () => {
    const promises = [
      apiClient.get('/api/channels'),
      apiClient.get('/api/videos'),
      apiClient.get('/api/config/smtp'),
    ];

    const responses = await Promise.all(promises);

    expect(responses).toHaveLength(3);
    responses.forEach(response => {
      expect(response.status).toBe(200);
    });
  });

  test('handles request with complex query parameters', async () => {
    const params = new URLSearchParams({
      refresh: 'true',
      limit: '10',
      sort: 'date',
    });

    const response = await apiClient.get(`/api/videos?${params.toString()}`);

    expect(response.status).toBe(200);
    expect(response.data).toHaveProperty('videos');
    expect(response.data.videos[0].title).toContain('Refreshed:');
  });

  test('handles request with complex payload', async () => {
    const complexPayload = {
      channels: [
        { id: 'UC1', title: 'Channel 1', meta: { verified: true } },
        { id: 'UC2', title: 'Channel 2', meta: { verified: false } },
      ],
      options: {
        overwrite: true,
        validateUrls: false,
      }
    };

    const response = await apiClient.post('/api/channels/import', complexPayload);

    expect(response.status).toBe(200);
    expect(response.data.imported).toBe(2);
  });
});

describe('Response Validation', () => {
  test('validates channel response structure', async () => {
    const response = await apiClient.get('/api/channels');

    expect(response.data).toEqual(
      expect.objectContaining({
        channels: expect.arrayContaining([
          expect.objectContaining({
            id: expect.any(String),
            title: expect.any(String),
            customUrl: expect.any(String),
            thumbnailUrl: expect.any(String),
            createdAt: expect.any(String),
            lastVideoPublishedAt: expect.any(String),
          })
        ]),
        totalCount: expect.any(Number),
        lastUpdated: expect.any(String),
      })
    );
  });

  test('validates video response structure', async () => {
    const response = await apiClient.get('/api/videos');

    expect(response.data).toEqual(
      expect.objectContaining({
        videos: expect.arrayContaining([
          expect.objectContaining({
            id: expect.any(String),
            title: expect.any(String),
            channelId: expect.any(String),
            cachedAt: expect.any(String),
            watched: expect.any(Boolean),
            published: expect.any(String),
            content: expect.any(String),
            link: expect.objectContaining({
              Href: expect.any(String),
              Rel: expect.any(String),
            }),
            author: expect.objectContaining({
              name: expect.any(String),
              uri: expect.any(String),
            }),
            mediaGroup: expect.objectContaining({
              mediaThumbnail: expect.objectContaining({
                URL: expect.any(String),
                Width: expect.any(String),
                Height: expect.any(String),
              }),
              mediaTitle: expect.any(String),
              mediaContent: expect.objectContaining({
                URL: expect.any(String),
                Type: expect.any(String),
                Width: expect.any(String),
                Height: expect.any(String),
              }),
              mediaDescription: expect.any(String),
            }),
          })
        ]),
        lastRefresh: expect.any(String),
        totalCount: expect.any(Number),
      })
    );
  });

  test('validates config response structure', async () => {
    const response = await apiClient.get('/api/config/smtp');

    expect(response.data).toEqual(
      expect.objectContaining({
        host: expect.any(String),
        port: expect.any(Number),
        username: expect.any(String),
        password: expect.any(String),
        fromAddress: expect.any(String),
        recipientEmail: expect.any(String),
        emailHour: expect.any(Number),
        emailMinute: expect.any(Number),
        emailTimezone: expect.any(String),
      })
    );
  });
});