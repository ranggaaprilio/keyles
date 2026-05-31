/**
 * Test setup configuration for vitest
 * Configures testing-library and global test utilities
 */

import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

// Node exposes unavailable storage globals without --localstorage-file.
const createStorage = (): Storage => {
  const values = new Map<string, string>();
  return {
    get length() {
      return values.size;
    },
    clear: () => values.clear(),
    getItem: key => values.get(key) ?? null,
    key: index => Array.from(values.keys())[index] ?? null,
    removeItem: key => values.delete(key),
    setItem: (key, value) => values.set(key, value),
  };
};

Object.defineProperty(globalThis, 'localStorage', { configurable: true, value: createStorage() });
Object.defineProperty(globalThis, 'sessionStorage', { configurable: true, value: createStorage() });

// Mock ResizeObserver (not available in jsdom, required by Radix UI)
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
} as unknown as typeof globalThis.ResizeObserver;

// Cleanup after each test
afterEach(() => {
  cleanup();
});
