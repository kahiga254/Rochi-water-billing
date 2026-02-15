"use client";

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function RootPage() {
  const router = useRouter();

  useEffect(() => {
    // Check if user is already logged in
    const token = localStorage.getItem('token');
    
    if (token) {
      // If logged in, redirect to appropriate dashboard based on role
      // You'd need to get the user role from localStorage or make an API call
      // For now, just redirect to admin
      router.replace('/admin');
    } else {
      // If not logged in, redirect to login page
      router.replace('/login');
    }
  }, [router]);

  // Show nothing while redirecting
  return null;
}