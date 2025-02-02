import React, { useState, useEffect, useRef } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Link,
  Navigate,
} from 'react-router-dom'
import {
  ChevronDown,
  ChevronUp,
  Search,
  Filter,
  X,
  User,
  Edit,
  Trash2,
} from 'lucide-react'
import Home from './components/Home'
import Header from './components/Header'
import Login from './components/Login'
import AddCategory from './components/AddCategory'
import MenuList from './components/MenuList'
import TableManager from './components/TableManager'
import UserManagement from './components/UserManagement'
import DashBoard from './components/DashBoard'
import AddMenu from './components/AddMenu'
import Printer from './components/Printer'
import ErrorBoundary from './ErrorBoundary' // import Error Boundary
import ChargeManagement from './components/ChargeManagement'
import PaymentTables from './components/PaymentTables'
import Menu from './pages/Menu'

import styles from './styles/App.module.css'

const App = () => {
  const location = useLocation()
  const [isDropdownOpen, setIsDropdownOpen] = useState(false) // ควบคุมสถานะของ Dropdown

  const toggleDropdown = () => {
    setIsDropdownOpen((prev) => !prev)
  }

  // ตรวจสอบว่า token มีอยู่ใน localStorage หรือไม่
  const isLoggedIn = !!localStorage.getItem('token')

  const [isSidebarOpen, setIsSidebarOpen] = useState(false)
  const sidebarRef = useRef()

  // ดึงข้อมูล user จาก localStorage
  let user = null
  try {
    const storedUser = localStorage.getItem('user')
    if (storedUser) {
      user = JSON.parse(storedUser)
    }
  } catch (error) {
    console.error('Error parsing user data:', error)
  }

  const toggleSidebar = () => {
    setIsSidebarOpen((prevState) => !prevState)
  }

  const handleResize = () => {
    if (window.innerWidth > 900) {
      setIsSidebarOpen(true)
    } else {
      setIsSidebarOpen(false)
    }
  }

  useEffect(() => {
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

  const navigate = useNavigate() // ใช้ useNavigate สำหรับการเปลี่ยนเส้นทาง
  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    navigate('/login') // เปลี่ยนเส้นทางไปหน้า login
  }

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
              <p className="text-sm ml-2 text-gold">
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

              <Link
                to="/menu"
                className={`${styles.navLink} ${
                  location.pathname === '/menu' ? styles.activeLink : ''
                }`}
              >
                เมนูอาหาร
              </Link>

              {/* ตรวจสอบสิทธิ์สำหรับ "เดชบอร์ด" */}
              {user?.role === 'manager' && (
                <Link
                  to="/dashboard"
                  className={`${styles.navLink} ${
                    location.pathname === '/dashboard' ? styles.activeLink : ''
                  }`}
                >
                  เดชบอร์ด
                </Link>
              )}

              <Link
                to="/tables"
                className={`${styles.navLink} ${
                  location.pathname === '/tables' ? styles.activeLink : ''
                }`}
              >
                จัดการโต๊ะ
              </Link>

              {/* ตรวจสอบสิทธิ์สำหรับ "ส่วนลด" */}
              {user?.role === 'manager' && (
                <Link
                  to="/discount"
                  className={`${styles.navLink} ${
                    location.pathname === '/discount' ? styles.activeLink : ''
                  }`}
                >
                  ส่วนลด
                </Link>
              )}

              {/* ตรวจสอบสิทธิ์สำหรับ "เครื่องพิมพ์" */}
              {user?.role === 'manager' && (
                <Link
                  to="/printer"
                  className={`${styles.navLink} ${
                    location.pathname === '/printer' ? styles.activeLink : ''
                  }`}
                >
                  เครื่องพิมพ์
                </Link>
              )}

              {/* ตรวจสอบสิทธิ์สำหรับ "จัดการผู้ใช้" */}
              {user?.role === 'manager' && (
                <Link
                  to="/user"
                  className={`${styles.navLink} ${
                    location.pathname === '/user' ? styles.activeLink : ''
                  }`}
                >
                  จัดการผู้ใช้
                </Link>
              )}

              {/* เมนูที่มี dropdown */}
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
                    {/* ตรวจสอบสิทธิ์สำหรับ "จัดการเมนูอาหาร" */}
                    {user?.role === 'manager' && (
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
                    )}

                    {/* ตรวจสอบสิทธิ์สำหรับ "จัดการหมวดหมู่" */}
                    {user?.role === 'manager' && (
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
                    )}
                  </div>
                )}
              </div>
            </nav>

            <button className={styles.logoutButton} onClick={handleLogout}>
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
              path="/menu"
              element={isLoggedIn ? <Menu /> : <Navigate to="/login" />}
            />

            <Route
              path="/tables"
              element={isLoggedIn ? <TableManager /> : <Navigate to="/login" />}
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
            <Route path="/payment-tables" element={<PaymentTables />} />
            <Route path="/menu/:tableId/:uuid" element={<Menu />} />
          </Routes>
        </main>
      </div>
    </ErrorBoundary>
  )
}

const Root = () => {
  return (
    <Router>
      <App />
    </Router>
  )
}

export default Root