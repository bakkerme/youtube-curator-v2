import { renderHook } from '@testing-library/react';
import { useWindowTitle, resetOriginalTitle } from './useWindowTitle';

describe('useWindowTitle', () => {
  const originalTitle = 'Original Title';

  beforeEach(() => {
    // Set a consistent original title before each test
    document.title = originalTitle;
    // Reset the global original title for clean tests
    resetOriginalTitle();
  });

  afterEach(() => {
    // Restore original title after each test
    document.title = originalTitle;
    // Reset the global original title
    resetOriginalTitle();
  });

  it('should not change title when inactive', () => {
    renderHook(() => useWindowTitle('Loading...', false));
    expect(document.title).toBe(originalTitle);
  });

  it('should update title when active', () => {
    renderHook(() => useWindowTitle('Loading...', true));
    expect(document.title).toBe(`Loading... - ${originalTitle}`);
  });

  it('should restore title when becoming inactive', () => {
    const { rerender } = renderHook(
      ({ status, isActive }) => useWindowTitle(status, isActive),
      { initialProps: { status: 'Loading...', isActive: true } }
    );

    expect(document.title).toBe(`Loading... - ${originalTitle}`);

    rerender({ status: 'Loading...', isActive: false });
    expect(document.title).toBe(originalTitle);
  });

  it('should update status message while remaining active', () => {
    const { rerender } = renderHook(
      ({ status, isActive }) => useWindowTitle(status, isActive),
      { initialProps: { status: 'Loading...', isActive: true } }
    );

    expect(document.title).toBe(`Loading... - ${originalTitle}`);

    rerender({ status: 'Refreshing...', isActive: true });
    expect(document.title).toBe(`Refreshing... - ${originalTitle}`);
  });

  it('should restore title on unmount', () => {
    const { unmount } = renderHook(() => useWindowTitle('Loading...', true));
    expect(document.title).toBe(`Loading... - ${originalTitle}`);

    unmount();
    expect(document.title).toBe(originalTitle);
  });

  it('should handle multiple instances correctly', () => {
    const { unmount: unmount1 } = renderHook(() => useWindowTitle('Loading...', true));
    expect(document.title).toBe(`Loading... - ${originalTitle}`);

    const { unmount: unmount2 } = renderHook(() => useWindowTitle('Processing...', true));
    expect(document.title).toBe(`Processing... - ${originalTitle}`);

    unmount2();
    // Title should be restored to original after the second hook unmounts
    expect(document.title).toBe(originalTitle);

    unmount1();
    expect(document.title).toBe(originalTitle);
  });
});