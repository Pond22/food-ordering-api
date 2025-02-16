import React, { useState, useEffect } from 'react'
import axios from 'axios'

const DeletedMenuTable = () => {
  const [menus, setMenus] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      alert('Session หมดอายุ กรุณาเข้าสู่ระบบใหม่')
      window.location.href = '/login' // Redirect ไปที่หน้า Login
      return
    }
    fetchDeletedMenus()
  }, [])
  
  const fetchDeletedMenus = async () => {
    try {
      const token = localStorage.getItem('token')
      const response = await axios.get(`${import.meta.env.VITE_APP_API_URL}/api/menu/deleted`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
      )
      setMenus(response.data)
      setLoading(false)
    } catch (err) {
      setError('ไม่สามารถดึงข้อมูลเมนูที่ถูกลบได้')
      setLoading(false)
    }
  }

  const handleRestoreMenu = async (menuId) => {
    try {
      const token = localStorage.getItem('token')
      const response = await axios.post(
        `${import.meta.env.VITE_APP_API_URL}/api/menu/restore/${menuId}`, {
          headers: {
            Authorization: `Bearer ${token}`, // ✅ เพิ่ม JWT
          },
        })
      setMessage(`เมนู "${response.data.Name}" ได้รับการกู้คืนเรียบร้อยแล้ว`)
      fetchDeletedMenus() // Refresh the list after restoring
    } catch (err) {
      setError('เกิดข้อผิดพลาดในการกู้คืนเมนู')
    }
  }

  if (loading) return <div>กำลังโหลดข้อมูล...</div>
  if (error) return <div className="text-red-500">{error}</div>

  return (
    <div className="">
      {/* <h1 className="text-xl font-bold mb-4">เมนูที่ถูกลบ</h1> */}
      {message && <div className="text-green-500 mb-4">{message}</div>}
      <table className="w-full max-x-full bg-white rounded-3xl shadow-lg border border-gray-300">
        <thead>
          <tr className="bg-blue-500 text-left text-white">
            <th className="px-4 py-2">ID</th>
            <th className="px-4 py-2">ชื่อ</th>
            <th className="px-4 py-2">ชื่อภาษาอังกฤษ</th>
            <th className="px-4 py-2">ชื่อภาษาจีน</th>
            <th className="px-4 py-2">หมวดหมู่</th>
            <th className="px-4 py-2">ราคา</th>
            <th className="px-4 py-2">การจัดการ</th>
          </tr>
        </thead>
        <tbody>
          {menus.map((menu) => (
            <tr key={menu.ID}>
              <td className="px-4 py-2">{menu.ID}</td>
              <td className="px-4 py-2">{menu.Name}</td>
              <td className="px-4 py-2">{menu.NameEn}</td>
              <td className="px-4 py-2">{menu.NameCh}</td>
              <td className="px-4 py-2">{menu.Category.Name}</td>
              <td className="px-4 py-2">{menu.Price}</td>
              <td className="px-4 py-2">
                <button
                  onClick={() => handleRestoreMenu(menu.ID)}
                  className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-700"
                >
                  กู้คืน
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default DeletedMenuTable
