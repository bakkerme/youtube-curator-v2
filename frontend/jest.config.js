const nextJest = require('next/jest')

// Providing the path to your Next.js app to load next.config.js and .env files in your test environment
const createJestConfig = nextJest({
  dir: './',
})

// Add any custom config to be passed to Jest
const customJestConfig = {
  // Add more setup options before each test is run
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  
  // if using TypeScript with a baseUrl set to the root directory then you need the below for alias' to work
  moduleDirectories: ['node_modules', '<rootDir>/'],
  
  // Handle module aliases (this will be automatically configured for you soon)
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/$1',
  },
  
  testEnvironment: 'jest-environment-jsdom',
  
  // Ignore .next and node_modules
  testPathIgnorePatterns: ['<rootDir>/.next/', '<rootDir>/node_modules/'],
  
  // Module paths to ignore for Haste
  modulePathIgnorePatterns: ['<rootDir>/.next/'],
  
  // Transform files
  transform: {
    '^.+\\.(js|jsx|ts|tsx)$': ['@swc/jest'],
  },
  
  // Handle ESM modules
  transformIgnorePatterns: [
    'node_modules/(?!(msw|@bundled-es-modules)/)',
  ],
  
  // Coverage configuration
  collectCoverageFrom: [
    'components/**/*.{js,jsx,ts,tsx}',
    'app/**/*.{js,jsx,ts,tsx}',
    'lib/**/*.{js,jsx,ts,tsx}',
    '!**/*.d.ts',
    '!**/node_modules/**',
    '!**/.next/**',
  ],
  
  // Match test files that are co-located with source files
  testMatch: [
    '**/__tests__/**/*.(ts|tsx|js|jsx)',
    '**/*.(test|spec).(ts|tsx|js|jsx)'
  ],
  
  // Module file extensions
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  
  // Treat TypeScript files as ESM
  extensionsToTreatAsEsm: ['.ts', '.tsx'],
}

// createJestConfig is exported this way to ensure that next/jest can load the Next.js config which is async
module.exports = createJestConfig(customJestConfig) 