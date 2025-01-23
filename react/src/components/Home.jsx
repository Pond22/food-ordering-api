import React, { useState, useEffect } from 'react'

import styles from '../styles/Header.module.css'

const Home = () => {
  const [currentTime, setCurrentTime] = useState(new Date())

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date())
    }, 1000)

    return () => clearInterval(timer)
  }, [])

  const employeeData = {
    name: 'อาคิระ คิมุระ',
    position: 'พนักงานเสิร์ฟ',
  }

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

  const generateQRCodeSVG = () => {
    const qrMatrix = generateQRMatrix('pos-login')

    return (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="160"
        height="160"
        viewBox="0 0 160 160"
      >
        {qrMatrix.map((row, y) =>
          row.map((cell, x) =>
            cell ? (
              <rect
                key={`${x}-${y}`}
                x={x * 10}
                y={y * 10}
                width="10"
                height="10"
                fill="#D32F2F"
              />
            ) : null
          )
        )}
      </svg>
    )
  }

  const generateQRMatrix = () => {
    const matrix = Array(16)
      .fill()
      .map(() => Array(16).fill(false))

    // Simple QR code pattern
    for (let i = 0; i < 16; i++) {
      if (i < 6 || i > 9) {
        matrix[i][0] = true
        matrix[i][5] = true
        matrix[0][i] = true
        matrix[5][i] = true
      }
    }

    return matrix
  }

  return (
    <div className="min-h-screen bg-gray-100 flex justify-center px-4 py-2">
      <div className="w-full bg-white rounded-xl shadow-lg overflow-hidden">
        <div className="bg-gradient-to-r from-red-500 to-pink-500 p-6 text-white text-center">
          <h1 className="text-2xl text-left font-bold">Grand Kaze Yakiniku Chang Mai</h1>
        </div>

        <div className="p-6 border-b">
          <div className="text-2xl text-gold mb-2">ยินดีต้อนรับเข้าสู่ระบบ</div>
          {/* <div className="w-24 h-24 rounded-full bg-gray-300 mx-auto mb-4 flex items-center justify-center">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              className="w-16 h-16 text-gray-500"
            >
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
              <circle cx="12" cy="7" r="4" />
            </svg>
          </div> */}
          {user && (
            <div className="text-black">
              <h2>ชื่อ {user.name}</h2>
              <h2>ตำแหน่ง: {user.role}</h2>
            </div>
          )}
        </div>

        <div className="p-6">
          <div className="bg-gray-100 rounded-xl p-4 mb-4 text-center">
            <div className="mb-4">
              <p className="text-gray-700 text-sm mb-2">เวลาปัจจุบัน</p>
              <p className="text-2xl font-semibold text-red-600">
                {currentTime.toLocaleTimeString('th-TH', {
                  hour: '2-digit',
                  minute: '2-digit',
                  second: '2-digit',
                  hour12: false,
                })}
              </p>
            </div>

            <div className="flex justify-center mb-4">
              {generateQRCodeSVG()}
            </div>
            <p className="text-gray-600 mb-4">QR Code สำหรับเข้าระบบ POS</p>
          </div>

          {/* <button className="w-full bg-red-500 text-white py-3 rounded-lg hover:bg-red-600 transition-colors">
            เข้าสู่ระบบ POS
          </button> */}
        </div>

        {/* <div className="bg-gray-200 p-4 text-center text-sm text-gray-600">
          © 2024 Sushi Master Restaurant
        </div> */}
      </div>
    </div>
  )
}

export default Home
