/**
 * Test setup configuration for vitest
 * Configures testing-library and global test utilities
 */

import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

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
