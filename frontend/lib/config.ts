interface RuntimeConfig {
  apiUrl: string;
}

let cachedConfig: RuntimeConfig | null = null;

export async function getRuntimeConfig(): Promise<RuntimeConfig> {
  if (cachedConfig) {
    return cachedConfig;
  }

  try {
    const response = await fetch('/api/config');
    if (!response.ok) {
      throw new Error(`Failed to fetch config: ${response.status}`);
    }
    
    cachedConfig = await response.json();
    return cachedConfig!;
  } catch (error) {
    console.warn('Failed to fetch runtime config, using defaults:', error);
    
    // Fallback to defaults if config fetch fails
    cachedConfig = {
      apiUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api',
    };
    
    return cachedConfig;
  }
}

// Function to clear cache if needed (useful for testing or config changes)
export function clearConfigCache() {
  cachedConfig = null;
} 