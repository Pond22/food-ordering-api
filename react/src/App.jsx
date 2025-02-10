import React, { useState, useEffect } from 'react';
import { Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import Login from './components/Login';
import Section from './Section';
import Menu from './pages/Menu';
import Pos from './pages/POS';
import POSVerifyRoute from './pages/POSVerifyRoute';

// Protected Route Component
const ProtectedRoute = ({ children, isLoggedIn }) => {
  if (!isLoggedIn) {
    return <Navigate to="/login" replace />;
  }
  return children;
};

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
    localStorage.removeItem('posToken'); // ลบ POS token ด้วย
    localStorage.removeItem('posSessionId');
    setIsLoggedIn(false);
    setUser(null);
    setToken(null);
    navigate('/login');
  };

  return (
    <Routes>
      {/* Public Routes */}
      <Route path="/login" element={
        isLoggedIn ? <Navigate to="/home" replace /> : <Login />
      } />
      <Route path="/menu" element={<Menu />} />

      {/* POS Routes */}
      <Route path="/pos/verify" element={<POSVerifyRoute />} />
      <Route path="/pos" element={
        <ProtectedRoute isLoggedIn={isLoggedIn}>
          <Pos />
        </ProtectedRoute>
      } />

      {/* Protected Routes */}
      <Route path="/home" element={
        <ProtectedRoute isLoggedIn={isLoggedIn}>
          <Section
            isLoggedIn={isLoggedIn}
            user={user}
            handleLogout={handleLogout}
            token={token}
          />
        </ProtectedRoute>
      } />

      {/* Root Route */}
      <Route path="/" element={
        isLoggedIn ? <Navigate to="/home" replace /> : <Navigate to="/login" replace />
      } />

      {/* Catch all route */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
};

export default App;