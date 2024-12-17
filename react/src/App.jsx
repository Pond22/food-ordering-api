import React, { useState, useRef, useEffect } from "react";
import { BrowserRouter as Router, Routes, Route, Link } from "react-router-dom";
import Header from "./components/Header";
import Login from "./components/Login";
import AddCategory from "./components/AddCategory";
import MenuList from "./components/MenuList";
import TableManager from "./components/TableManager";
import UserManagement from "./components/UserManagement";
import DashBoard from "./components/DashBoard";
import AddMenu from "./components/AddMenu";

import styles from "./styles/App.module.css";

const App = () => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const sidebarRef = useRef(); // ใช้ useRef เพื่อตรวจจับ sidebar

  // ฟังก์ชันสำหรับสลับการเปิด/ปิด sidebar
  const toggleSidebar = () => {
    setIsSidebarOpen((prevState) => !prevState); // ใช้ prevState เพื่อให้แน่ใจว่าเป็นการสลับสถานะอย่างถูกต้อง
  };

  // ฟังก์ชันเพื่อเช็คขนาดหน้าจอและจัดการการปิด sidebar
  const handleResize = () => {
    if (window.innerWidth > 900) {
      setIsSidebarOpen(true); // ถ้าขนาดหน้าจอเกิน 900px, เปิด sidebar
    } else {
      setIsSidebarOpen(false); // ถ้าขนาดหน้าจอไม่เกิน 900px, ปิด sidebar
    }
  };

  // ใช้ useEffect เพื่อตรวจจับการเปลี่ยนแปลงขนาดหน้าจอ
  useEffect(() => {
    // เรียก handleResize ทุกครั้งที่ขนาดหน้าจอเปลี่ยน
    window.addEventListener("resize", handleResize);
    
    // เรียก handleResize ครั้งแรกเมื่อ component ถูก mount
    handleResize();

    // ทำความสะอาด (clean-up) โดยการลบ event listener เมื่อ component ถูก unmount
    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, []);

  // ปิด sidebar เมื่อคลิกที่นอก sidebar
  useEffect(() => {
    const handleClickOutside = (event) => {
      // ตรวจสอบว่าคลิกที่ข้างนอก sidebar หรือไม่
      // เงื่อนไขนี้จะทำงานเฉพาะเมื่อหน้าจอเล็กกว่า 900px
      if (window.innerWidth < 903 && sidebarRef.current && !sidebarRef.current.contains(event.target)) {
        setIsSidebarOpen(false); // ปิด sidebar
      }
    };

    // เฉพาะเมื่อหน้าจอเล็กกว่า 900px จะเพิ่ม event listener สำหรับคลิกข้างนอก
    if (window.innerWidth < 903) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside); // ลบ event listener เมื่อ component ถูกทำลาย
    };
  }, [isSidebarOpen]); // เพิ่ม isSidebarOpen เป็น dependency เพื่อให้ตรวจจับการเปลี่ยนแปลง

  return (
    <Router>
      <div className={styles.appLayout}>
        {/* <header className={styles.header}>

        </header> */}
        {/* Sidebar */}
        <aside
          ref={sidebarRef}
          className={`${styles.sidebar} ${isSidebarOpen ? styles.open : ""}`}
          style={{
            transition: "transform 0.3s ease, opacity 0.3s ease",
            transform: isSidebarOpen ? "translateX(0)" : "translateX(-100%)", // เปิด/ปิด sidebar
            opacity: isSidebarOpen ? 1 : 0,
          }}
        >
          <div className={styles.logo}>LOGO</div>
          <nav className={styles.navMenu}>
            <Link to="/login" className={styles.navLink}>Login</Link>
            <Link to="/menu" className={styles.navLink}>เมนูอาหาร</Link>
            <Link to="/dashboard" className={styles.navLink}>เดชบอร์ด</Link>
            <Link to="/tables" className={styles.navLink}>จัดการโต๊ะ</Link>
            <Link to="/user" className={styles.navLink}>จัดการผู้ใช้</Link>
            <Link to="/addMenu" className={styles.navLink}>จัดการเมนูอาหาร</Link>
            <Link to="/addCategory" className={styles.navLink}>เพิ่มหมวดหมู่</Link>
          </nav>
          <button className={styles.logoutButton}>LOGOUT</button>
        </aside>

        {/* Hamburger Menu */}
        <button
          className={styles.menuIcon}
          onClick={toggleSidebar} // เมื่อคลิก hamburger จะเรียก toggleSidebar
        >
          <span className={styles.bar}></span>
          <span className={styles.bar}></span>
          <span className={styles.bar}></span>
        </button>
        <body>
        {/* Main Content */}
        <main className={styles.mainContent}>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/dashboard" element={<DashBoard />} />
            <Route path="/menu" element={<MenuList />} />
            <Route path="/tables" element={<TableManager />} />
            <Route path="/user" element={<UserManagement />} />
            <Route path="/addMenu" element={<AddMenu />} />
            <Route path="/addCategory" element={<AddCategory />} />
          </Routes>
        </main>
        
          
        </body>
      </div>
    </Router>
  );
};

export default App;
