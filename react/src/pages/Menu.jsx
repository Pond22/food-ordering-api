import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import MenuBar from '../components/MenuBar'
import MenuList2 from '../components/MenuList2'
import axios from 'axios'

export default function Menu() {
  const { tableId, uuid } = useParams() // ดึง tableId และ uuid จาก URL
  const [menuData, setMenuData] = useState([])

  useEffect(() => {
    // เมื่อมี tableId และ uuid, คุณสามารถใช้ข้อมูลนี้ในการดึงเมนูจาก API หรือการทำงานอื่นๆ
    if (tableId && uuid) {
      fetchMenuData(tableId, uuid)
    }
  }, [tableId, uuid])

  const fetchMenuData = async (tableId, uuid) => {
    try {
      // ดึงข้อมูลเมนูจาก API ตาม tableId และ uuid
      const response = await axios.get(`/api/menu/${tableId}?uuid=${uuid}`)
      setMenuData(response.data)
    } catch (error) {
      console.error('Error fetching menu:', error)
      alert('ไม่สามารถโหลดเมนูได้')
    }
  }

  return (
    <div className="bg-gray-100 min-h-screen">
      {/* ส่ง tableId และ uuid ไปยัง MenuBar และ MenuList2 */}
      <MenuBar tableId={tableId} uuid={uuid} />
      <MenuList2 tableId={tableId} uuid={uuid} menuData={menuData} />
    </div>
  )
}
