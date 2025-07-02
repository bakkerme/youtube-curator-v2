import { createServer } from 'http';

/**
 * Check if a port is available (not in use)
 */
async function isPortAvailable(port: number): Promise<boolean> {
  return new Promise((resolve) => {
    const server = createServer();
    
    server.once('error', (err: any) => {
      if (err.code === 'EADDRINUSE') {
        resolve(false); // Port is in use
      } else {
        resolve(false); // Other error means we can't use it
      }
    });
    
    server.once('listening', () => {
      server.close();
      resolve(true); // Port is available
    });
    
    server.listen(port);
  });
}

/**
 * Global setup function that runs before all tests
 * Checks if localhost:3000 is already in use and errors if so
 */
async function globalSetup() {
  const port = 3000;
  const available = await isPortAvailable(port);
  
  if (!available) {
    console.error(`\n❌ Error: Port ${port} is already in use!`);
    console.error(`Please stop any existing dev servers before running screenshot tests.`);
    console.error(`The running server may cause incorrect test outputs.\n`);
    process.exit(1);
  }
  
  console.log(`✅ Port ${port} is available for testing`);
}

export default globalSetup;