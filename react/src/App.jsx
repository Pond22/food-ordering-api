import React, { useState, useEffect } from 'react';
import { Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import Login from './components/Login';
import Section from './Section';
import Menu from './pages/Menu';

import Pos from './pages/POS';
import POSVerifyRoute from './pages/POSVerifyRoute';
import PaymentTables from './components/PaymentTables';
// import PosSection from './PosSection';
import TableManager from './components/TableManager';
const App = () => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    const token = localStorage.getItem('token');
    const userData = localStorage.getItem('user');
    if (token && userData) {
      setIsLoggedIn(true);
      setToken(token);
      setUser(JSON.parse(userData));
    } else {
      setIsLoggedIn(false);
      setUser(null);
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('posToken');
    localStorage.removeItem('posSessionId');
    setIsLoggedIn(false);
    setUser(null);
    setToken(null);
    navigate('/login');
  };

  return (
    <Routes>
      {/* POS Routes (ไม่ต้องการการล็อกอิน) */}
      {/* <Route path="/" element={<PosSection />} /> */}
      <Route path="/pos/verify" element={<POSVerifyRoute />} />
      <Route path="/pos" element={<Pos />} />
      {/* <Route path="/payment-tables" element={<PaymentTables user={user} />} /> */}

      {/* Public Routes */}
      <Route path="/menu" element={<Menu />} />
      <Route
        path="/login"
        element={isLoggedIn ? <Navigate to="/home" /> : <Login />}
      />

      {/* Main Application Routes */}
      <Route
        path="/*"
        element={
          isLoggedIn ? (
            <Section
              isLoggedIn={isLoggedIn}
              user={user}
              handleLogout={handleLogout}
              token={token}
            />
          ) : (
            <Navigate to="/login" />
          )
        }
      />
      {/* <Route path="/tables" element={<TableManager  />} /> */}
      <Route path="/payment-tables" element={<PaymentTables  />} />
    </Routes>
  )
};

export default App;