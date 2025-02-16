import React, { useState, useEffect, useRef } from 'react'
import { Routes, Route, useLocation, Link, Navigate } from 'react-router-dom' // นำเข้า Routes และ Route
import {
  ChevronDown,
  ChevronUp,
  Bell,
  X,
  UserRound,
  CreditCard,
  Check,
} from 'lucide-react'
import styles from './styles/App.module.css'
import Home from './components/Home'
import DashBoard from './components/DashBoard'
import TableManager from './components/TableManager'
import Printer from './components/Printer'
import UserManagement from './components/UserManagement'
import AddMenu from './components/AddMenu'
import AddCategory from './components/AddCategory'
import ChargeManagement from './components/ChargeManagement'
import PaymentTables from './components/PaymentTables'
import ErrorBoundary from './ErrorBoundary'
import OrderConfirmation from './components/OrderConfirmation'
import Reprint from './components/Reprint'
import ReservationConfig from './components/ReservationConfig';
import MenuImportPage from './components/MenuImportPage'
import axios from 'axios'

const Section = ({ isLoggedIn, user, handleLogout, token }) => {
  const location = useLocation()
  const [isDropdownOpen, setIsDropdownOpen] = useState(false)
  const [isSidebarOpen, setIsSidebarOpen] = useState(false)
  const sidebarRef = useRef()
  const [showNotifications, setShowNotifications] = useState(false)
  const [notifications, setNotifications] = useState([])
  const [ws, setWs] = useState(null)

  const handleDismiss = (id) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      try {
        const message = {
          notification_id: id.toString(),
          staff_id: user.id  // เพิ่ม staff_id
        };
        ws.send(JSON.stringify(message));
        
        setNotifications(prev => {
          const notification = prev.find(n => n.id === id);
          if (!notification) return prev;

          const confirmMessage = notification.type === 'payment' 
            ? `กำลังไปเก็บเงินที่โต๊ะ ${notification.table}`
            : `กำลังไปให้บริการที่โต๊ะ ${notification.table}`;
          alert(confirmMessage);

          return prev.filter(n => n.id !== id);
        });
      } catch (error) {
        console.error('Error sending acknowledgment:', error);
        alert('เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง');
      }
    }
  };

  const markAllAsRead = () => {
    setNotifications(notifications.map((note) => ({ ...note, read: true })))
  }

  const toggleDropdown = () => {
    setIsDropdownOpen((prev) => !prev)
  }

  const toggleSidebar = () => {
    setIsSidebarOpen((prevState) => !prevState)
  }

  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth > 900) {
        setIsSidebarOpen(true)
      } else {
        setIsSidebarOpen(false)
      }
    }

    window.addEventListener('resize', handleResize)
    handleResize()

    return () => {
      window.removeEventListener('resize', handleResize)
    }
  }, [])

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (
        window.innerWidth < 1024 &&
        sidebarRef.current &&
        !sidebarRef.current.contains(event.target)
      ) {
        setIsSidebarOpen(false)
      }
    }

    if (window.innerWidth < 1024) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isSidebarOpen])

  useEffect(() => {
    const ws = new WebSocket(`${import.meta.env.VITE_APP_WS_URL}/ws/staff?api_key=${import.meta.env.VITE_APP_WS_STAFF_KEY}`);
    ws.onmessage = (event) => {
      const notification = JSON.parse(event.data);
      
      if (notification.status === 'read') {
        // ลบการแจ้งเตือนที่ถูกรับทราบแล้ว
        setNotifications(prev => prev.filter(n => n.id !== notification.id));
      } else {
        // เพิ่มการแจ้งเตือนใหม่
        setNotifications(prev => {
          if (prev.some(n => n.id === notification.id)) return prev;
          
          return [...prev, {
            id: notification.id,
            type: notification.type === 'payment' ? 'payment' : 'service',
            table: notification.table_id,
            message: notification.message,
            time: new Date(notification.created_at).toLocaleTimeString(),
            read: false
          }];
        });
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    setWs(ws);

    return () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
    };
  }, []);

  return (
    <ErrorBoundary>
      <div className={styles.appLayout}>
        {isLoggedIn && (
          <aside
            ref={sidebarRef}
            className={`${styles.sidebar} ${isSidebarOpen ? styles.open : ''}`}
            style={{
              transition: 'transform 0.3s ease, opacity 0.3s ease',
              transform: isSidebarOpen ? 'translateX(0)' : 'translateX(-100%)',
              opacity: isSidebarOpen ? 1 : 0,
            }}
          >
            <div className={styles.logo}>
              <img
                className="w-14 h-14 rounded-full"
                src="src\image\images (1).jpg"
                alt="Logo"
              />
              <p className="text-sm ml-2 bg-gradient-to-t from-yellow-400 to-yellow-600 text-transparent bg-clip-text">
                Grand Kaze Yakiniku Chang mai
              </p>
            </div>
            <nav className={styles.navMenu}>
              <Link
                to="/home"
                className={`${styles.navLink} ${
                  location.pathname === '/home' ? styles.activeLink : ''
                }`}
              >
                หน้าหลัก
              </Link>

              {/* เมนูสำหรับผู้ใช้ที่เป็น manager */}
              {user?.role === 'manager' && (
                <>
                  <Link
                    to="/dashboard"
                    className={`${styles.navLink} ${
                      location.pathname === '/dashboard'
                        ? styles.activeLink
                        : ''
                    }`}
                  >
                    เดชบอร์ด
                  </Link>

                  <Link
                    to="/discount"
                    className={`${styles.navLink} ${
                      location.pathname === '/discount' ? styles.activeLink : ''
                    }`}
                  >
                    ส่วนลด
                  </Link>

                  <Link
                    to="/printer"
                    className={`${styles.navLink} ${
                      location.pathname === '/printer' ? styles.activeLink : ''
                    }`}
                  >
                    เครื่องพิมพ์
                  </Link>

                  <Link
                    to="/reservation-config"
                    className={`${styles.navLink} ${
                      location.pathname === '/reservation-config' ? styles.activeLink : ''
                    }`}
                  >
                    ตั้งค่า Rules การจอง
                  </Link>

                  <Link
                    to="/user"
                    className={`${styles.navLink} ${
                      location.pathname === '/user' ? styles.activeLink : ''
                    }`}
                  >
                    จัดการผู้ใช้
                  </Link>

                  {/* เมนู dropdown สำหรับจัดการเมนูอาหาร */}
                  <div className="dropdown">
                    <div className="flex">
                      <button
                        className="dropdown-toggle nav-link"
                        onClick={toggleDropdown}
                      >
                        จัดการอาหาร
                      </button>
                      {isDropdownOpen ? <ChevronUp /> : <ChevronDown />}
                    </div>
                    {isDropdownOpen && (
                      <div className="dropdown-menu">
                        <Link
                          to="/addMenu"
                          className={`${styles.navLink} ${
                            location.pathname === '/addMenu'
                              ? styles.activeLink
                              : ''
                          }`}
                        >
                          จัดการเมนูอาหาร
                        </Link>

                        <Link
                          to="/MenuImportPage"
                          className={`${styles.navLink} ${
                            location.pathname === '/MenuImportPage'
                              ? styles.activeLink
                              : ''
                          }`}
                        >
                          เพิ่มรายการอาหาร via excel
                        </Link>

                        <Link
                          to="/addCategory"
                          className={`${styles.navLink} ${
                            location.pathname === '/addCategory'
                              ? styles.activeLink
                              : ''
                          }`}
                        >
                          จัดการหมวดหมู่
                        </Link>
                      </div>
                    )}
                  </div>
                </>
              )}

              {/* เมนูทั่วไป */}
              {/* <Link
                to="/tables"
                className={`${styles.navLink} ${
                  location.pathname === '/tables' ? styles.activeLink : ''
                }`}
              >
                จัดการโต๊ะ
              </Link> */}

              {/* <Link
                to="/reprint"
                className={`${styles.navLink} ${
                  location.pathname === '/reprint' ? styles.activeLink : ''
                }`}
              >
                rePrint
              </Link> */}

              <Link
                to="/comfirmation"
                className={`${styles.navLink} ${
                  location.pathname === '/comfirmation' ? styles.activeLink : ''
                }`}
              >
                รายการออเดอร์
              </Link>
            </nav>

            <button className={styles.logoutButton} onClick={() => {
              // ลบ token ออกจาก localStorage
              localStorage.removeItem('token');
              localStorage.removeItem('user');
              handleLogout();
            }}>
              LOGOUT
            </button>
          </aside>
        )}
        <button className={styles.menuIcon} onClick={toggleSidebar}>
          <span className={styles.bar}></span>
          <span className={styles.bar}></span>
          <span className={styles.bar}></span>
        </button>

        <main className={styles.mainContent}>
          <Routes>
            <Route
              path="/login"
              element={!isLoggedIn ? <Login /> : <Navigate to="/home" />}
            />

            <Route
              path="/home"
              element={isLoggedIn ? <Home /> : <Navigate to="/login" />}
            />

            <Route
              path="/reprint"
              element={isLoggedIn ? <Reprint /> : <Navigate to="/login" />}
            />

            <Route path="/reservation-config" element={<ReservationConfig />} />
          
            <Route
              path="/comfirmation"
              element={
                isLoggedIn ? <OrderConfirmation /> : <Navigate to="/login" />
              }
            />

            <Route
              path="/dashboard"
              element={
                isLoggedIn && user?.role === 'manager' ? (
                  <DashBoard />
                ) : (
                  <Navigate to="/login" />
                )
              }
            />

            

            <Route
              path="/printer"
              element={
                isLoggedIn && user?.role === 'manager' ? (
                  <Printer />
                ) : (
                  <Navigate to="/login" />
                )
              }
            />

            <Route
              path="/user"
              element={
                isLoggedIn && user?.role === 'manager' ? (
                  <UserManagement />
                ) : (
                  <Navigate to="/login" />
                )
              }
            />

            <Route
              path="/addMenu"
              element={
                isLoggedIn && user?.role === 'manager' ? (
                  <AddMenu />
                ) : (
                  <Navigate to="/login" />
                )
              }
            />

            <Route
              path="/MenuImportPage"
              element={
                isLoggedIn && user?.role === 'manager' ? (
                  <MenuImportPage />
                ) : (
                  <Navigate to="/login" />
                )
              }
            />

            <Route
              path="/discount"
              element={
                isLoggedIn && user?.role === 'manager' ? (
                  <ChargeManagement />
                ) : (
                  <Navigate to="/login" />
                )
              }
            />

            <Route
              path="/addCategory"
              element={
                isLoggedIn && user?.role === 'manager' ? (
                  <AddCategory />
                ) : (
                  <Navigate to="/login" />
                )
              }
            />

            
          </Routes>

          <div className="fixed bottom-4 right-4 z-50">
            {/* Notification Icon */}
            <button
              onClick={() => setShowNotifications(!showNotifications)}
              className="relative bg-white p-3 rounded-full shadow-lg hover:bg-gray-50"
            >
              <Bell className="w-6 h-6 text-gray-700" />
              {notifications.some((note) => !note.read) && (
                <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs w-5 h-5 flex items-center justify-center rounded-full">
                  {notifications.filter((note) => !note.read).length}
                </span>
              )}
            </button>

            {/* Notification Popup */}
            {showNotifications && (
              <div className="absolute bottom-16 right-0 w-80 bg-white rounded-lg shadow-xl border overflow-hidden">
                <div className="p-4 border-b bg-gray-50 flex justify-between items-center">
                  <h3 className="font-semibold text-gray-700">การแจ้งเตือน</h3>
                  <button
                    onClick={markAllAsRead}
                    className="text-sm text-blue-600 hover:text-blue-800"
                  >
                    อ่านทั้งหมด
                  </button>
                </div>

                <div className="max-h-96 overflow-y-auto">
                  {notifications.length === 0 ? (
                    <div className="p-4 text-center text-gray-500">
                      ไม่มีการแจ้งเตือน
                    </div>
                  ) : (
                    <div className="divide-y">
                      {notifications.map((notification) => (
                        <div
                          key={notification.id}
                          className={`p-4 hover:bg-gray-50 flex items-start justify-between ${
                            !notification.read ? 'bg-blue-50' : ''
                          }`}
                        >
                          <div className="flex items-start space-x-3">
                            <div
                              className={`rounded-full p-2 ${
                                notification.type === 'service'
                                  ? 'bg-blue-100 text-blue-600'
                                  : 'bg-green-100 text-green-600'
                              }`}
                            >
                              {notification.type === 'service' ? (
                                <UserRound className="w-5 h-5" />
                              ) : (
                                <CreditCard className="w-5 h-5" />
                              )}
                            </div>
                            <div>
                              <p className="font-medium text-gray-800">
                                โต๊ะ {notification.table}
                              </p>
                              <p className="text-sm text-gray-600">
                                {notification.message}
                              </p>
                              <p className="text-xs text-gray-400 mt-1">
                                {notification.time}
                              </p>
                            </div>
                          </div>
                          <div className="flex gap-2">
                            {/* ปุ่มรับทราบ */}
                            <button
                              onClick={() => handleDismiss(notification.id)}
                              className="px-2 py-1 bg-green-500 text-white rounded-md hover:bg-green-600 text-sm flex items-center gap-1"
                            >
                              <Check className="w-4 h-4" />
                              รับทราบ
                            </button>
                            {/* ปุ่มปิด */}
                            <button
                              onClick={() => setNotifications(prev => prev.filter(n => n.id !== notification.id))}
                              className="text-gray-400 hover:text-gray-600"
                            >
                              <X className="w-5 h-5" />
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                <div className="p-3 bg-gray-50 border-t">
                  <button
                    onClick={() => setShowNotifications(false)}
                    className="w-full py-2 text-sm text-center text-gray-600 hover:text-gray-800"
                  >
                    ปิด
                  </button>
                </div>
              </div>
            )}
          </div>
        </main>
      </div>
    </ErrorBoundary>
  )
}

export default Section
