'use client';

import { useState } from 'react';
import { authApi } from '@/lib/api/auth';

export default function TestAuthPage() {
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const testLogin = async () => {
    setLoading(true);
    try {
      console.log('🧪 Test: Starting login test...');
      const response = await authApi.login('admin', 'Admin@123');
      console.log('🧪 Test: Login response:', response);
      setResult({ success: true, data: response });
    } catch (error: any) {
      console.error('🧪 Test: Login error:', error);
      setResult({ 
        success: false, 
        error: error.message,
        response: error.response?.data,
        status: error.response?.status
      });
    } finally {
      setLoading(false);
    }
  };

  const testDirectFetch = async () => {
    setLoading(true);
    try {
      console.log('🧪 Test: Starting direct fetch...');
      const response = await fetch('http://localhost:8080/api/v1/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: 'admin',
          password: 'Admin@123'
        })
      });
      
      const data = await response.json();
      console.log('🧪 Test: Direct fetch response:', data);
      setResult({ 
        success: response.ok, 
        status: response.status,
        data 
      });
    } catch (error: any) {
      console.error('🧪 Test: Direct fetch error:', error);
      setResult({ success: false, error: error.message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">🔧 Auth Diagnostic Tool</h1>
      
      <div className="space-y-4">
        <div className="flex gap-4">
          <button
            onClick={testLogin}
            disabled={loading}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            Test authApi.login()
          </button>
          
          <button
            onClick={testDirectFetch}
            disabled={loading}
            className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
          >
            Test Direct Fetch
          </button>
          
          <button
            onClick={() => {
              localStorage.clear();
              sessionStorage.clear();
              console.log('🧪 Storage cleared');
              setResult({ success: true, message: 'Storage cleared' });
            }}
            className="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700"
          >
            Clear Storage
          </button>
        </div>

        {loading && (
          <div className="p-4 bg-blue-50 text-blue-700 rounded-lg">
            Testing... Check the browser console (F12) for logs
          </div>
        )}

        {result && (
          <div className="mt-6">
            <h2 className="text-lg font-semibold mb-2">Result:</h2>
            <pre className="p-4 bg-gray-100 dark:bg-gray-800 rounded-lg overflow-auto text-sm">
              {JSON.stringify(result, null, 2)}
            </pre>
          </div>
        )}

        <div className="mt-8 p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
          <h2 className="font-semibold mb-2">📋 Instructions:</h2>
          <ol className="list-decimal list-inside space-y-1 text-sm">
            <li>Open Browser DevTools (F12)</li>
            <li>Go to Console tab</li>
            <li>Click "Test authApi.login()"</li>
            <li>Check the console logs - they're color-coded:
              <ul className="list-disc list-inside ml-4 mt-1">
                <li className="text-blue-600">🔵 [authApi] - API calls</li>
                <li className="text-purple-600">📤 [apiClient] - Request</li>
                <li className="text-green-600">📥 [apiClient] - Response</li>
                <li className="text-red-600">🔴 Error logs</li>
              </ul>
            </li>
          </ol>
        </div>
      </div>
    </div>
  );
}