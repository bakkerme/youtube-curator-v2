import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { videoAPI } from '../api';

// Mock the runtime config
jest.mock('../config', () => ({
  getRuntimeConfig: jest.fn().mockResolvedValue({
    apiUrl: 'http://localhost:8080/api'
  })
}));

// Setup MSW server for API mocking
const server = setupServer();

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('videoAPI', () => {
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