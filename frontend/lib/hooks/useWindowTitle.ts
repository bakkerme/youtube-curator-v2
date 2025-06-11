import { useEffect, useRef } from 'react';

// Store the original title globally to handle multiple hook instances
let globalOriginalTitle: string | null = null;

/**
 * Reset the global original title - useful for testing
 */
export function resetOriginalTitle() {
  globalOriginalTitle = null;
}

/**
 * Custom hook for managing window title updates during long-running operations
 * @param status - The status message to display in the title
 * @param isActive - Whether the operation is currently active
 */
export function useWindowTitle(status: string, isActive: boolean) {
  const wasActiveRef = useRef(false);

  useEffect(() => {
    // Store the original title on first usage across all instances
    if (globalOriginalTitle === null) {
      globalOriginalTitle = document.title;
    }

    if (isActive) {
      // Update title with status message
      document.title = `${status} - ${globalOriginalTitle}`;
      wasActiveRef.current = true;
    } else if (wasActiveRef.current && globalOriginalTitle) {
      // Restore original title only if this hook was previously active
      document.title = globalOriginalTitle;
      wasActiveRef.current = false;
    }
  }, [status, isActive]);

  // Cleanup: restore original title on unmount if this hook was active
  useEffect(() => {
    return () => {
      if (wasActiveRef.current && globalOriginalTitle) {
        document.title = globalOriginalTitle;
      }
    };
  }, []);
}