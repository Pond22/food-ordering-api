import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Power,
  AlertTriangle,
  Loader2,
  Clock,
  UserCircle,
  MonitorSmartphone,
  Globe,
  Cpu,
  Network,
  Printer,
  X,
} from 'lucide-react'
import TableManager from '../components/TableManager'
import Reprint from '../components/Reprint'

const API_BASE_URL = 'http://127.0.0.1:8080/api/pos'

const POS = () => {
  const navigate = useNavigate()
  const [sessionStatus, setSessionStatus] = useState(null)
  const [isLoading, setIsLoading] = useState(true)
  const [showLogoutDialog, setShowLogoutDialog] = useState(false)
  const [isLoggingOut, setIsLoggingOut] = useState(false)
  const [showReprintDialog, setShowReprintDialog] = useState(false)

  // ตรวจสอบ session status ทุก 1 นาที
  useEffect(() => {
    const checkSession = async () => {
      try {
        const response = await fetch(
          `${API_BASE_URL}/session-status`,
          {
            headers: {
              Authorization: `Bearer ${localStorage.getItem('posToken')}`,
              'Content-Type': 'application/json',
              'Accept': 'application/json',
            },
          }
        )

        if (!response.ok) {
          throw new Error('Session check failed')
        }

        const data = await response.json()
        setSessionStatus(data)

        // ถ้า session ไม่ active ให้ redirect ไปหน้า login
        if (!data.is_active) {
          alert('POS session expired. Please verify with a new code.')
          // เคลียร์ token เก่าออก
          localStorage.removeItem('posToken')
          localStorage.removeItem('posSessionId')
          // ทำการ redirect ไปที่หน้าหลักของร้าน (ให้พนักงานขอ code ใหม่)
          navigate('/pos/verify')
        }
      } catch (error) {
        console.error('Session check error:', error)
        localStorage.removeItem('posToken')
        localStorage.removeItem('posSessionId')
        navigate('/pos/verify')
      } finally {
        setIsLoading(false)
      }
    }

    checkSession()
    const interval = setInterval(checkSession, 60000) // ทุก 1 นาที

    return () => clearInterval(interval)
  }, [navigate])

  // เพิ่มฟังก์ชันสำหรับเก็บข้อมูลอุปกรณ์
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

  const generateDeviceId = () => {
    const components = [
      navigator.userAgent,
      navigator.language,
      window.screen.colorDepth,
      window.screen.width,
      window.screen.height,
      new Date().getTimezoneOffset(),
      navigator.hardwareConcurrency,
      navigator.platform
    ].filter(Boolean).join('|');
    
    let hash = 0;
    for (let i = 0; i < components.length; i++) {
      const char = components.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
    
    const deviceId = Math.abs(hash).toString(36) + Date.now().toString(36);
    localStorage.setItem('device_id', deviceId);
    return deviceId;
  };

  const handleLogout = async () => {
    setIsLoggingOut(true);
    try {
      const deviceInfo = getDeviceInfo();
      console.log('Logging out with device info:', deviceInfo);

      const response = await fetch(`${API_BASE_URL}/logout`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('posToken')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ deviceInfo })
      });

      if (!response.ok) {
        throw new Error('Logout failed');
      }

      localStorage.removeItem('posToken');
      localStorage.removeItem('posSessionId');
      navigate('/pos/verify');
    } catch (error) {
      console.error('Logout error:', error);
      alert('Failed to logout. Please try again.');
    } finally {
      setIsLoggingOut(false);
      setShowLogoutDialog(false);
    }
  };

  const formatBytes = (bytes) => {
    if (!bytes) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center space-y-4">
          <Loader2 className="h-8 w-8 animate-spin mx-auto" />
          <p className="text-sm text-gray-500">Loading POS system...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header with session info */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-between items-center">
            <div className="grid grid-cols-4 gap-6 flex-grow mr-4">
              {/* Staff Info */}
              <div className="bg-gray-50 p-3 rounded-lg">
                <div className="flex items-center space-x-2 mb-2">
                  <UserCircle className="h-5 w-5 text-blue-500" />
                  <span className="font-medium">Staff Info</span>
                </div>
                <div className="text-sm space-y-1">
                  <p>{sessionStatus?.staff_name || 'Staff'}</p>
                  <p className="text-gray-500">ID: {sessionStatus?.staff_id}</p>
                </div>
              </div>

              {/* Session Info */}
              <div className="bg-gray-50 p-3 rounded-lg">
                <div className="flex items-center space-x-2 mb-2">
                  <Clock className="h-5 w-5 text-green-500" />
                  <span className="font-medium">Session</span>
                </div>
                <div className="text-sm space-y-1">
                  <p>
                    Started:{' '}
                    {new Date(sessionStatus?.start_time).toLocaleTimeString()}
                  </p>
                  <p className="text-gray-500">
                    Last Activity:{' '}
                    {new Date(
                      sessionStatus?.last_activity
                    ).toLocaleTimeString()}
                  </p>
                </div>
              </div>

              {/* Device Info */}
              <div className="bg-gray-50 p-3 rounded-lg">
                <div className="flex items-center space-x-2 mb-2">
                  <MonitorSmartphone className="h-5 w-5 text-purple-500" />
                  <span className="font-medium">Device</span>
                </div>
                <div className="text-sm space-y-1">
                  <p>{sessionStatus?.device_info?.platform}</p>
                  <p
                    className="text-gray-500 text-xs truncate"
                    title={sessionStatus?.device_info?.user_agent}
                  >
                    {sessionStatus?.device_info?.user_agent}
                  </p>
                  <p className="text-gray-500">
                    {sessionStatus?.device_info?.screen_width}x
                    {sessionStatus?.device_info?.screen_height}
                  </p>
                  <p className="text-gray-500">
                    {sessionStatus?.device_info?.vendor}
                  </p>
                </div>
              </div>

              {/* System Info */}
              <div className="bg-gray-50 p-3 rounded-lg">
                <div className="flex items-center space-x-2 mb-2">
                  <Network className="h-5 w-5 text-orange-500" />
                  <span className="font-medium">System</span>
                </div>
                <div className="text-sm space-y-1">
                  <div className="flex items-center">
                    <Globe className="h-4 w-4 mr-1 text-gray-400" />
                    <span
                      className="truncate"
                      title={sessionStatus?.ip_address}
                    >
                      {sessionStatus?.ip_address}
                    </span>
                  </div>
                  {sessionStatus?.device_info?.network_type && (
                    <div className="flex items-center">
                      <Cpu className="h-4 w-4 mr-1 text-gray-400" />
                      <span>
                        {sessionStatus?.device_info?.network_type.type}(
                        {sessionStatus?.device_info?.network_type.downlink}Mbps,
                        RTT: {sessionStatus?.device_info?.network_type.rtt}ms)
                      </span>
                    </div>
                  )}
                  <p className="text-gray-500">
                    {sessionStatus?.device_info?.language}
                  </p>
                  <p className="text-gray-500">
                    {sessionStatus?.device_info?.timezone}
                  </p>
                </div>
              </div>
            </div>

            {/* เพิ่มปุ่ม Reprint ก่อนปุ่ม Logout */}
            <div className="flex gap-2">
              <button
                onClick={() => setShowReprintDialog(true)}
                className="flex items-center px-4 py-2 text-blue-600 hover:text-blue-700 bg-white hover:bg-blue-50 border border-blue-200 rounded-lg transition-colors"
              >
                <Printer className="h-4 w-4 mr-2" />
                รีปริ้น
              </button>
              
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
        </div>
      </header>

      {/* Main POS content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Add your POS content here */}
        <TableManager
          user={sessionStatus?.staff_id}
          posToken={localStorage.getItem('posToken')}
        />
      </main>

      {/* เพิ่ม Reprint Dialog */}
      {showReprintDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-6xl h-[90vh] overflow-hidden">
            <div className="flex justify-between items-center p-4 border-b">
              <h2 className="text-lg font-semibold">ระบบจัดการรีปริ้น</h2>
              <button
                onClick={() => setShowReprintDialog(false)}
                className="p-2 hover:bg-gray-100 rounded-full"
              >
                <X className="h-5 w-5" />
              </button>
            </div>
            <div className="p-4 h-[calc(90vh-80px)] overflow-auto">
              <Reprint posToken={localStorage.getItem('posToken')} />
            </div>
          </div>
        </div>
      )}

      {/* Logout confirmation dialog */}
      {showLogoutDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
            <div className="flex items-center gap-2 mb-4">
              <AlertTriangle className="h-5 w-5 text-yellow-500" />
              <h2 className="text-lg font-semibold">Confirm Logout</h2>
            </div>
            <p className="text-gray-600 mb-6">
              Are you sure you want to end this POS session? This action cannot
              be undone.
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
  )
}

export default POS