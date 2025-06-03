import { useState, useEffect } from 'react';
import { getRuntimeConfig } from '../config';

interface UseConfigReturn {
  config: { apiUrl: string } | null;
  loading: boolean;
  error: string | null;
}

export function useConfig(): UseConfigReturn {
  const [config, setConfig] = useState<{ apiUrl: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    const loadConfig = async () => {
      try {
        setLoading(true);
        setError(null);
        const runtimeConfig = await getRuntimeConfig();
        
        if (isMounted) {
          setConfig(runtimeConfig);
        }
      } catch (err) {
        if (isMounted) {
          setError(err instanceof Error ? err.message : 'Failed to load configuration');
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    };

    loadConfig();

    return () => {
      isMounted = false;
    };
  }, []);

  return { config, loading, error };
} 