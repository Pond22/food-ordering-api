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

import axios from 'axios'
import POSVerifyRoute from './pages/POSVerifyRoute'
import Pos from './pages/Pos'
import PaymentTables from './components/PaymentTables'
import ErrorBoundary from './ErrorBoundary'

const PosSection = ({ user, token }) => {
  

  return (
    <ErrorBoundary>
      <Routes>
        <Route path="/pos/verify" element={<POSVerifyRoute />} />
        <Route path="/pos" element={<Pos />} />
        {/* <Route path="/payment-tables" element={<PaymentTables user={user} />} /> */}
        <Route path="/payment-tables" element={<PaymentTables user={user} />} />
      </Routes>
    </ErrorBoundary>
  )
}

export default PosSection
