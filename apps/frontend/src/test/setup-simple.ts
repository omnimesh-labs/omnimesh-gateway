import 'whatwg-fetch';
import '@testing-library/jest-dom';

// Mock localStorage and sessionStorage are set up in the main setup file
// Mock window.localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};

  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value.toString();
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
    key: (index: number) => {
      const keys = Object.keys(store);
      return keys[index] || null;
    },
    get length() {
      return Object.keys(store).length;
    }
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

// Mock window.sessionStorage
const sessionStorageMock = (() => {
  let store: Record<string, string> = {};

  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value.toString();
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
    key: (index: number) => {
      const keys = Object.keys(store);
      return keys[index] || null;
    },
    get length() {
      return Object.keys(store).length;
    }
  };
})();

Object.defineProperty(window, 'sessionStorage', {
  value: sessionStorageMock,
});

// Mock window.location - check if it exists and is configurable
if (!window.location || Object.getOwnPropertyDescriptor(window, 'location')?.configurable !== false) {
  // Only mock if it doesn't exist or is configurable
  try {
    Object.defineProperty(window, 'location', {
      value: {
        href: 'http://localhost:3000',
        pathname: '/',
        search: '',
        hash: '',
        assign: jest.fn(),
        replace: jest.fn(),
        reload: jest.fn(),
      },
      writable: true,
      configurable: true,
    });
  } catch (error) {
    // If we can't redefine location, just extend the existing one
    Object.assign(window.location, {
      assign: jest.fn(),
      replace: jest.fn(),
      reload: jest.fn(),
    });
  }
}
