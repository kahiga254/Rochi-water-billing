"use client";

import { useEffect } from 'react';
import { useAuthStore } from '@/lib/store/authStore';

export function AuthInitializer() {
  useEffect(() => {
    // Check auth status when the app loads
    useAuthStore.getState().checkAuth();
  }, []);

  return null; // This component doesn't render anything
}