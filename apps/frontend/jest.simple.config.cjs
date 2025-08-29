/** @type {import('jest').Config} */
module.exports = {
  // Test environment
  testEnvironment: 'jsdom',

  // Setup files
  setupFiles: ['whatwg-fetch', '<rootDir>/src/test/polyfills.ts'],
  setupFilesAfterEnv: ['<rootDir>/src/test/setup-simple.ts'],

  // Module name mapping for TypeScript path aliases
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    // Handle CSS imports
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
    // Handle image imports
    '\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$': '<rootDir>/src/test/__mocks__/fileMock.ts',
  },

  // Test patterns - only run simple tests
  testMatch: [
    '<rootDir>/src/**/*simple.test.{js,jsx,ts,tsx}'
  ],

  // Transform configuration
  transform: {
    '^.+\\.(js|jsx|ts|tsx)$': 'babel-jest'
  },

  // Module file extensions
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx'],

  // Ignore patterns
  testPathIgnorePatterns: ['<rootDir>/.next/', '<rootDir>/node_modules/'],

  // Environment variables for tests
  testEnvironmentOptions: {
    url: 'http://localhost:3000',
  },
};
