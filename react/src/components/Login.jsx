import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Lock, User, AlertTriangle, Eye, EyeOff } from 'lucide-react';

const API_BASE_URL = 'http://127.0.0.1:8080/api/auth'

const Login = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  
  // เพิ่ม useEffect เพื่อตรวจสอบ token เมื่อโหลดหน้า
  useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem('token');
      if (token) {
        try {
          const userResponse = await axios.get(`${API_BASE_URL}/verify-token`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });

          if (userResponse.status === 200) {
            const { role } = userResponse.data;
            if (role === "manager") {
              window.location.href = "/dashboard";
            } else if (role === "staff") {
              window.location.href = "/home";
            }
          }
        } catch (err) {
          localStorage.removeItem('token');
          localStorage.removeItem('user');
        }
      }
    };
    
    checkAuth();
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);
  
    try {
      if (!username || !password) {
        throw new Error('กรุณากรอกชื่อผู้ใช้และรหัสผ่าน');
      }
  
      // เรียก API Login
      const response = await axios.post(`${API_BASE_URL}/login`, {
        username: username,
        password: password,
      });
  
      if (response.status === 200) {
        const { token } = response.data;
        localStorage.setItem('token', token);
  
        // เรียก API เพื่อดึงข้อมูลผู้ใช้
        const userResponse = await axios.get(`${API_BASE_URL}/verify-token`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
  
        if (userResponse.status === 200) {
          const { id, name, role, username: userUsername } = userResponse.data;
          const userData = { id, name, role, username: userUsername };
          localStorage.setItem('user', JSON.stringify(userData));
  
          // นำทางไปยังหน้าหลักตาม Role
          if (role === "manager") {
            window.location.href = "/dashboard"; // Manager ไปที่ Dashboard
          } else if (role === "staff") {
            window.location.href = "/home"; // Staff ไปที่ Tables
          } else {
            window.location.href = "/login"; // กรณีไม่มี Role ที่ระบุ
          }
        }
      }
    } catch (err) {
      if (err.response) {
        if (err.response.status === 400) {
          setError('ข้อมูลไม่ถูกต้อง');
        } else if (err.response.status === 401) {
          setError('ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง หรือ Token หมดอายุ');
        } else {
          setError('เกิดข้อผิดพลาดที่ไม่คาดคิด');
        }
      } else {
        setError('ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center relative overflow-hidden p-4 sm:p-6 md:p-8">
      {/* Luxurious Background Layer */}
      <div className="absolute inset-0 z-0">
        <div className="absolute inset-0 bg-gradient-to-br from-black via-[#2c1810] to-[#4a1d12]"></div>

        {/* Gold Decorative Elements */}
        <div className="absolute top-0 left-0 w-48 h-48 sm:w-64 sm:h-64 bg-gradient-to-br from-yellow-600/30 to-yellow-900/30 rounded-full -translate-x-1/2 -translate-y-1/2"></div>
        <div className="absolute bottom-0 right-0 w-72 h-72 sm:w-96 sm:h-96 bg-gradient-to-tl from-yellow-600/30 to-yellow-900/30 rounded-full translate-x-1/2 translate-y-1/2"></div>

        {/* Subtle Texture Overlay */}
        <div className="absolute inset-0 opacity-10 bg-[radial-gradient(#ffffff_1px,transparent_1px)] [background-size:16px_16px]"></div>
      </div>

      {/* Login Container */}
      <div className="w-full max-w-md relative z-10">
        <div className="bg-white/95 backdrop-blur-sm shadow-2xl rounded-2xl overflow-hidden border border-yellow-900/20">
          {/* Elegant Header */}
          <div className="bg-gradient-to-r from-black to-gray-900 text-white p-4 sm:p-6 text-center">
            <h1 className="text-2xl sm:text-3xl font-semibold tracking-wide">
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-yellow-500 to-yellow-700">
                和
              </span>{' '}
              Grand Kaze Yakiniku
            </h1>
            <p className="text-gray-300 mt-1 sm:mt-2 text-xs sm:text-sm">
              ระบบสำหรับพนักงาน
            </p>
          </div>

          {/* Login Form */}
          <form
            onSubmit={handleSubmit}
            className="p-6 sm:p-8 space-y-4 sm:space-y-6"
          >
            {error && (
              <div className="bg-red-50 border border-red-200 text-red-800 p-3 sm:p-4 rounded-lg flex items-center">
                <AlertTriangle className="mr-2 sm:mr-3 text-red-600 w-4 h-4 sm:w-5 sm:h-5" />
                <span className="text-sm sm:text-base">{error}</span>
              </div>
            )}

            <div className="relative">
              <label
                className="block text-gray-700 text-xs sm:text-sm font-bold mb-1 sm:mb-2"
                htmlFor="username"
              >
                ชื่อผู้ใช้
              </label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4 sm:w-5 sm:h-5" />
                <input
                  id="username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="กรอกชื่อผู้ใช้"
                  className="w-full pl-8 sm:pl-10 pr-4 py-2 sm:py-3 text-xs sm:text-base border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-yellow-500 transition duration-300"
                  disabled={isLoading}
                  required
                />
              </div>
            </div>

            <div className="relative">
              <label
                className="block text-gray-700 text-xs sm:text-sm font-bold mb-1 sm:mb-2"
                htmlFor="password"
              >
                รหัสผ่าน
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4 sm:w-5 sm:h-5" />
                <input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="กรอกรหัสผ่าน"
                  className="w-full pl-8 sm:pl-10 pr-10 py-2 sm:py-3 text-xs sm:text-base border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-yellow-500 transition duration-300"
                  disabled={isLoading}
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-2 sm:right-3 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-yellow-600 transition duration-300"
                >
                  {showPassword ? (
                    <Eye className="w-4 h-4 sm:w-5 sm:h-5" />
                  ) : (
                    <EyeOff className="w-4 h-4 sm:w-5 sm:h-5" />
                  )}
                </button>
              </div>
            </div>

            <button
              type="submit"
              className={`w-full py-2 sm:py-3 rounded-lg text-white font-bold text-xs sm:text-base transition duration-300 ${
                isLoading
                  ? 'bg-gray-400 cursor-not-allowed'
                  : 'bg-gradient-to-r from-red-600 to-red-800 hover:from-red-700 hover:to-red-900'
              }`}
              disabled={isLoading}
            >
              {isLoading ? 'กำลังเข้าสู่ระบบ...' : 'เข้าสู่ระบบ'}
            </button>

            <div className="text-center">
              <a
                href="#"
                className="text-xs sm:text-sm text-gray-600 hover:text-red-600 transition duration-300"
              >
                ลืมรหัสผ่าน?
              </a>
            </div>
          </form>
        </div>

        <div className="text-center text-gray-300 mt-3 sm:mt-4 text-xs sm:text-sm relative z-10">
          © 2024 Restaurant Management System
        </div>
      </div>
    </div>
  )
};

export default Login;
