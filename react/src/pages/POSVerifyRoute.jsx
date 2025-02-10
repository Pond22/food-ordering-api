import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Loader2, ShieldCheck, AlertCircle } from 'lucide-react';

const POSVerifyRoute = () => {
  const [verificationCode, setVerificationCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleVerification = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const response = await fetch('http://localhost:8080/api/pos/verify-code', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          code: verificationCode,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Verification failed');
      }

      // เก็บ token ลงใน localStorage
      localStorage.setItem('posToken', data.token);
      localStorage.setItem('posSessionId', data.session_id);
      localStorage.setItem('staffId', data.staff_id);

      // นำทางไปยังหน้า POS
      navigate('/pos');
    } catch (err) {
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