import { extractVideoId } from './api';

describe('extractVideoId', () => {
  it('should extract video ID from standard YouTube URL', () => {
    const url = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ';
    expect(extractVideoId(url)).toBe('dQw4w9WgXcQ');
  });

  it('should extract video ID from youtu.be URL', () => {
    const url = 'https://youtu.be/dQw4w9WgXcQ';
    expect(extractVideoId(url)).toBe('dQw4w9WgXcQ');
  });

  it('should extract video ID from embed URL', () => {
    const url = 'https://www.youtube.com/embed/dQw4w9WgXcQ';
    expect(extractVideoId(url)).toBe('dQw4w9WgXcQ');
  });

  it('should extract video ID from YouTube Shorts URL', () => {
    const url = 'https://www.youtube.com/shorts/DM-pfKGioWw';
    expect(extractVideoId(url)).toBe('DM-pfKGioWw');
  });

  it('should extract video ID from URL with additional parameters', () => {
    const url = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=30s';
    expect(extractVideoId(url)).toBe('dQw4w9WgXcQ');
  });

  it('should return video ID as-is when already a valid ID', () => {
    const videoId = 'dQw4w9WgXcQ';
    expect(extractVideoId(videoId)).toBe('dQw4w9WgXcQ');
  });

  it('should handle URLs with additional query parameters', () => {
    const url = 'https://www.youtube.com/watch?list=PLxxx&v=dQw4w9WgXcQ&index=1';
    expect(extractVideoId(url)).toBe('dQw4w9WgXcQ');
  });

  it('should trim whitespace from input', () => {
    const videoId = '  dQw4w9WgXcQ  ';
    expect(extractVideoId(videoId)).toBe('dQw4w9WgXcQ');
  });

  it('should return input as-is for invalid format', () => {
    const invalidInput = 'not-a-valid-id';
    expect(extractVideoId(invalidInput)).toBe('not-a-valid-id');
  });
});