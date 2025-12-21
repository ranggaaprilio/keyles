/**
 * Zustand store for OTP verification countdown state
 */

import { create } from 'zustand';

interface OTPStore {
  countdown: number;
  isActive: boolean;
  startCountdown: (seconds?: number) => void;
  resetCountdown: () => void;
  tick: () => void;
}

const DEFAULT_COUNTDOWN = 60; // 60 seconds

export const useOTPStore = create<OTPStore>((set) => ({
  countdown: 0,
  isActive: false,

  startCountdown: (seconds = DEFAULT_COUNTDOWN) => {
    set({ countdown: seconds, isActive: true });

    const interval = setInterval(() => {
      set((state) => {
        const newCountdown = state.countdown - 1;
        if (newCountdown <= 0) {
          clearInterval(interval);
          return { countdown: 0, isActive: false };
        }
        return { countdown: newCountdown };
      });
    }, 1000);
  },

  resetCountdown: () => {
    set({ countdown: 0, isActive: false });
  },

  tick: () => {
    set((state) => {
      const newCountdown = state.countdown - 1;
      if (newCountdown <= 0) {
        return { countdown: 0, isActive: false };
      }
      return { countdown: newCountdown };
    });
  },
}));
