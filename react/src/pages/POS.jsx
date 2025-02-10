import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Power,
  AlertTriangle,
  Loader2,
  Clock,
  UserCircle,
  MonitorSmartphone
} from 'lucide-react';

const POS = () => {
  const navigate = useNavigate();
  const [sessionStatus, setSessionStatus] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showLogoutDialog, setShowLogoutDialog] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  // ตรวจสอบ session status ทุก 1 นาที
  useEffect(() => {
    const checkSession = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/pos/session-status', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('posToken')}`,
          },
        });

        if (!response.ok) {
          throw new Error('Session check failed');
        }

        const data = await response.json();
        setSessionStatus(data);

        // ถ้า session ไม่ active ให้ redirect ไปหน้า login
        if (!data.is_active) {
          alert('POS session expired. Please verify with a new code.');
          // เคลียร์ token เก่าออก
          localStorage.removeItem('posToken');
          localStorage.removeItem('posSessionId');
          // ทำการ redirect ไปที่หน้าหลักของร้าน (ให้พนักงานขอ code ใหม่)
          navigate('/');
        }
      } catch (error) {
        console.error('Session check error:', error);
        localStorage.removeItem('posToken');
        localStorage.removeItem('posSessionId');
        navigate('/');
      } finally {
        setIsLoading(false);
      }
    };

    checkSession();
    const interval = setInterval(checkSession, 60000); // ทุก 1 นาที

    return () => clearInterval(interval);
  }, [navigate]);

  const handleLogout = async () => {
    setIsLoggingOut(true);
    try {
      const response = await fetch('http://localhost:8080/api/pos/logout', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('posToken')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Logout failed');
      }

      // ลบข้อมูล session จาก localStorage
      localStorage.removeItem('posToken');
      localStorage.removeItem('posSessionId');
      navigate('/pos/verify');
    } catch (error) {
      alert('Failed to logout. Please try again.');
    } finally {
      setIsLoggingOut(false);
      setShowLogoutDialog(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center space-y-4">
          <Loader2 className="h-8 w-8 animate-spin mx-auto" />
          <p className="text-sm text-gray-500">Loading POS system...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header with session info */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <UserCircle className="h-5 w-5 text-gray-500" />
                <span className="text-sm font-medium">
                  {sessionStatus?.staff_name || 'Staff'}
                </span>
              </div>
              <div className="flex items-center space-x-2">
                <Clock className="h-5 w-5 text-gray-500" />
                <span className="text-sm">
                  {new Date(sessionStatus?.start_time).toLocaleTimeString()}
                </span>
              </div>
              <div className="flex items-center space-x-2">
                <MonitorSmartphone className="h-5 w-5 text-gray-500" />
                <span className="text-sm text-gray-600">
                  {sessionStatus?.device_info || 'Unknown Device'}
                </span>
              </div>
            </div>
            <button
              onClick={() => setShowLogoutDialog(true)}
              disabled={isLoggingOut}
              className="flex items-center px-4 py-2 text-red-600 hover:text-red-700 bg-white hover:bg-red-50 border border-red-200 rounded-lg transition-colors disabled:opacity-50"
            >
              <Power className="h-4 w-4 mr-2" />
              {isLoggingOut ? 'Logging out...' : 'End Session'}
            </button>
          </div>
        </div>
      </header>

      {/* Main POS content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Add your POS content here */}
      </main>

      {/* Logout confirmation dialog */}
      {showLogoutDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
            <div className="flex items-center gap-2 mb-4">
              <AlertTriangle className="h-5 w-5 text-yellow-500" />
              <h2 className="text-lg font-semibold">Confirm Logout</h2>
            </div>
            <p className="text-gray-600 mb-6">
              Are you sure you want to end this POS session? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setShowLogoutDialog(false)}
                disabled={isLoggingOut}
                className="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg"
              >
                Cancel
              </button>
              <button
                onClick={handleLogout}
                disabled={isLoggingOut}
                className="flex items-center px-4 py-2 text-white bg-red-600 hover:bg-red-700 rounded-lg disabled:opacity-50"
              >
                {isLoggingOut ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Logging out...
                  </>
                ) : (
                  'End Session'
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default POS;