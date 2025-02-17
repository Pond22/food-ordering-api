import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Loader2, ShieldCheck, AlertCircle } from 'lucide-react';

const API_BASE_URL = `${import.meta.env.VITE_APP_API_URL}/api/pos`

const POSVerifyRoute = () => {
  const [verificationCode, setVerificationCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  // ใช้ getDeviceInfo แบบเดียวกับที่ใช้ใน POS.jsx
  const getDeviceInfo = () => {
    return {
      user_agent: navigator.userAgent,
      platform: navigator.platform || "Unknown",
      screen_width: window.screen?.width || window.innerWidth || 0,
      screen_height: window.screen?.height || window.innerHeight || 0,
      language: navigator.languages ? navigator.languages.join(',') : navigator.language,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      vendor: navigator.vendor || "Unknown",
      network_type: JSON.stringify((() => {
        const connection = navigator.connection || 
                          navigator.mozConnection || 
                          navigator.webkitConnection;
        if (connection) {
          return {
            type: connection.effectiveType || connection.type || 'unknown',
            downlink: connection.downlink || null,
            rtt: connection.rtt || null
          };
        }
        return {};
      })())
    };
  };

  const handleVerification = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const deviceInfo = getDeviceInfo();
      console.log('Sending verification with device info:', deviceInfo);

      const response = await fetch(`${API_BASE_URL}/verify-code`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          code: verificationCode,
          deviceInfo
        }),
      });

      console.log('Request body sent:', deviceInfo);
      console.log('Response status:', response.status);

      const data = await response.json();
      console.log('Verification response:', data);

      if (!response.ok) {
        throw new Error(data.error || 'Verification failed');
      }

      localStorage.setItem('posToken', data.token);
      localStorage.setItem('posSessionId', data.session_id);
      localStorage.setItem('staffId', data.staff_id);

      navigate('/pos');
    } catch (err) {
      console.error('Verification error:', err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-50">
      <div className="bg-white rounded-lg shadow-lg p-8 w-96">
        <div className="mb-6">
          <div className="flex items-center gap-2 mb-2">
            <ShieldCheck className="h-6 w-6 text-blue-500" />
            <h1 className="text-xl font-semibold">POS Verification</h1>
          </div>
          <p className="text-gray-500 text-sm">
            Enter the verification code to access POS system
          </p>
        </div>

        <form onSubmit={handleVerification} className="space-y-4">
          <div className="space-y-2">
            <input
              type="text"
              placeholder="Enter 6-digit code"
              value={verificationCode}
              onChange={(e) => setVerificationCode(e.target.value)}
              maxLength={6}
              className="w-full text-center text-2xl tracking-wider p-3 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition"
              disabled={isLoading}
            />
            {error && (
              <div className="bg-red-50 text-red-700 p-3 rounded-lg flex items-center">
                <AlertCircle className="h-4 w-4 mr-2" />
                {error}
              </div>
            )}
          </div>
          <button
            type="submit"
            className={`w-full bg-blue-600 hover:bg-blue-700 text-white p-3 rounded-lg font-medium flex items-center justify-center ${
              isLoading || verificationCode.length !== 6 ? 'opacity-50 cursor-not-allowed' : ''
            }`}
            disabled={verificationCode.length !== 6 || isLoading}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Verifying...
              </>
            ) : (
              'Verify & Access POS'
            )}
          </button>
        </form>
      </div>
    </div>
  );
};

export default POSVerifyRoute;