import React, { useState, useEffect } from 'react'
import { Routes, Route, useNavigate } from 'react-router-dom'
import Login from './components/Login'
import Section from './Section'
import Menu from './pages/Menu'

const App = () => {
  const [isLoggedIn, setIsLoggedIn] = useState(false)
  const [user, setUser] = useState(null)
  const navigate = useNavigate()

  useEffect(() => {
    const token = localStorage.getItem('token')
    const userData = localStorage.getItem('user')
    if (token && userData) {
      setIsLoggedIn(true)
      setUser(JSON.parse(userData))
    } else {
      setIsLoggedIn(false)
      setUser(null)
    }
  }, [])

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    setIsLoggedIn(false)
    setUser(null)
    navigate('/login') // ไปที่หน้า login หลัง logout
  }

  return (
    <>
      {/* เส้นทางที่ไม่มีการล็อกอิน */}
      <Routes>
        {/* เส้นทางสำหรับเมนูของโต๊ะที่มี tableId และ uuid */}
        <Route path="/menu" element={<Menu />} />

        {/* เส้นทางอื่นๆ สำหรับ login */}
        {!isLoggedIn ? (
          <Route path="/login" element={<Login />} />
        ) : (
          // เมื่อผู้ใช้ล็อกอินสำเร็จ จะแสดง Section โดยไม่ใช้ Route
          <Route
            path="/home"
            element={
              <Section
                isLoggedIn={isLoggedIn}
                user={user}
                handleLogout={handleLogout}
              />
            }
          />
        )}
      </Routes>

      {/* แสดง Section โดยตรงหากล็อกอินแล้ว */}
      {isLoggedIn && (
        <Section
          isLoggedIn={isLoggedIn}
          user={user}
          handleLogout={handleLogout}
        />
      )}
    </>
  )
}

export default App
