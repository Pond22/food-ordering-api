import React, { useState, useEffect, useRef } from "react";
import { BrowserRouter as Router, Routes, Route, Link, Navigate } from "react-router-dom";
import Header from "./components/Header";
import Login from "./components/Login";
import AddCategory from "./components/AddCategory";
import MenuList from "./components/MenuList";
import TableManager from "./components/TableManager";
import UserManagement from "./components/UserManagement";
import DashBoard from "./components/DashBoard";
import AddMenu from "./components/AddMenu";
import ErrorBoundary from "./ErrorBoundary";  // import Error Boundary
import styles from "./styles/App.module.css";

const App = () => {
  // ตรวจสอบว่า token มีอยู่ใน localStorage หรือไม่
  const isLoggedIn = !!localStorage.getItem("token");

  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const sidebarRef = useRef();

  // ดึงข้อมูล user จาก localStorage
  let user = null;
  try {
    const storedUser = localStorage.getItem("user");
    if (storedUser) {
      user = JSON.parse(storedUser);
    }
  } catch (error) {
    console.error("Error parsing user data:", error);
  }

  const toggleSidebar = () => {
    setIsSidebarOpen((prevState) => !prevState);
  };

  const handleResize = () => {
    if (window.innerWidth > 900) {
      setIsSidebarOpen(true);
    } else {
      setIsSidebarOpen(false);
    }
  };

  useEffect(() => {
    window.addEventListener("resize", handleResize);
    handleResize();

    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, []);

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (window.innerWidth < 903 && sidebarRef.current && !sidebarRef.current.contains(event.target)) {
        setIsSidebarOpen(false);
      }
    };

    if (window.innerWidth < 903) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isSidebarOpen]);

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
     // ใช้ Navigate เพื่อเปลี่ยนเส้นทางไปยังหน้าล็อกอิน
  window.location.href = '/login';
  };

  return (
    <ErrorBoundary>
      <Router>
        <div className={styles.appLayout}>
          {isLoggedIn && (
            <aside
              ref={sidebarRef}
              className={`${styles.sidebar} ${isSidebarOpen ? styles.open : ""}`}
              style={{
                transition: "transform 0.3s ease, opacity 0.3s ease",
                transform: isSidebarOpen ? "translateX(0)" : "translateX(-100%)",
                opacity: isSidebarOpen ? 1 : 0,
              }}
            >
              <div className={styles.logo}>LOGO</div>
              <nav className={styles.navMenu}>
                <Link to="/menu" className={styles.navLink}>เมนูอาหาร</Link>
                <Link to="/dashboard" className={styles.navLink}>เดชบอร์ด</Link>
                <Link to="/tables" className={styles.navLink}>จัดการโต๊ะ</Link>
                <Link to="/user" className={styles.navLink}>จัดการผู้ใช้</Link>
                <Link to="/addMenu" className={styles.navLink}>จัดการเมนูอาหาร</Link>
                <Link to="/addCategory" className={styles.navLink}>เพิ่มหมวดหมู่</Link>
              </nav>

              {user && (
                <div className={styles.userInfo}>
                  <p>{user.name}</p>
                  <p>{user.role}</p>
                </div>
              )}

              <button className={styles.logoutButton} onClick={handleLogout}>LOGOUT</button>
            </aside>
          )}

          <button className={styles.menuIcon} onClick={toggleSidebar}>
            <span className={styles.bar}></span>
            <span className={styles.bar}></span>
            <span className={styles.bar}></span>
          </button>

          <main className={styles.mainContent}>
            <Routes>
              {/* เส้นทางเข้าสู่ระบบ */}
              <Route
                path="/login"
                element={!isLoggedIn ? <Login /> : <Navigate to={user?.role === 'manager' ? '/dashboard' : '/tables'} />}
              />

              {/* Dashboard - อนุญาตเฉพาะ Manager */}
              <Route
                path="/dashboard"
                element={isLoggedIn && user?.role === 'manager' ? <DashBoard /> : <Navigate to="/login" />}
              />

              {/* Menu - อนุญาตทุกคนที่ล็อกอิน */}
              <Route
                path="/menu"
                element={isLoggedIn ? <MenuList /> : <Navigate to="/login" />}
              />

              {/* Tables - อนุญาต */}
              <Route
                path="/tables"
                element={isLoggedIn ? <TableManager /> : <Navigate to="/login" />}
              />

              {/* User Management - อนุญาตเฉพาะ Manager */}
              <Route
                path="/user"
                element={isLoggedIn && user?.role === 'manager' ? <UserManagement /> : <Navigate to="/login" />}
              />

              {/* Add Menu - อนุญาตเฉพาะ Manager */}
              <Route
                path="/addMenu"
                element={isLoggedIn && user?.role === 'manager' ? <AddMenu /> : <Navigate to="/login" />}
              />

              {/* Add Category - อนุญาตเฉพาะ Manager */}
              <Route
                path="/addCategory"
                element={isLoggedIn && user?.role === 'manager' ? <AddCategory /> : <Navigate to="/login" />}
              />
            </Routes>
          </main>
        </div>
      </Router>
    </ErrorBoundary>
  );
};

export default App;

