import React, { useState, useEffect } from 'react'

const Home = () => {
  const [currentTime, setCurrentTime] = useState(new Date())
  const [verificationCode, setVerificationCode] = useState(null) // State สำหรับเก็บ verification_code
  const token = localStorage.getItem('token')

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date())
    }, 1000)

    return () => clearInterval(timer)
  }, [])

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

  // ฟังก์ชันสำหรับการเรียก API
  const generateVerificationCode = async () => {
    try {
      const response = await fetch(
        'http://localhost:8080/api/pos/generate-code',
        {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
            Accept: 'application/json',
          },
        }
      )

      const data = await response.json()
      
      if (response.ok) {
        setVerificationCode(data.verification_code)
      } else {
        console.error('Error details:', data)
        alert(`Failed to generate code: ${data.error || data.details || 'Unknown error'}`)
      }
    } catch (error) {
      console.error('Error occurred:', error)
      alert('Failed to connect to server')
    }
  }

  return (
    <div className="min-h-screen bg-gray-100 flex justify-center px-4 py-2">
      <div className="w-full bg-white rounded-xl shadow-lg overflow-hidden">
        <div className="bg-gradient-to-r from-red-500 to-pink-500 p-6 text-white text-center">
          <h1 className="text-2xl text-left font-bold">
            Grand Kaze Yakiniku Chang Mai
          </h1>
        </div>

        <div className="p-6 border-b">
          <div className="text-2xl text-gold mb-2">ยินดีต้อนรับเข้าสู่ระบบ</div>
          {user && (
            <div className="text-black">
              <h2>ชื่อ {user.name}</h2>
              <h2>ตำแหน่ง: {user.role}</h2>
              <h2>Token: {token}</h2>
            </div>
          )}
        </div>

        <div className="p-6">
          <div className="bg-gray-100 rounded-xl p-4 mb-4 text-center">
            <p className="text-gray-600 mb-4">รหัสสำหรับเข้าระบบ POS</p>
            <div className="flex justify-center mb-4">
              <button
                className="px-2 py-1 bg-green-400 hover:bg-green-500 text-white rounded-full"
                onClick={generateVerificationCode} // เมื่อคลิกจะเรียกฟังก์ชัน
              >
                สร้างรหัส
              </button>
            </div>

            {/* แสดง verification_code หากมีการสร้างรหัสสำเร็จ */}
            {verificationCode && (
              <div className="text-2xl font-semibold text-red-600 mt-4">
                รหัสยืนยัน: {verificationCode}
              </div>
            )}

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
          </div>
        </div>
      </div>
    </div>
  )
}

export default Home
