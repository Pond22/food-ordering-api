import React, { useState, useEffect } from 'react'
import axios from 'axios'
import {
  ShieldCheck,
  Plus,
  Search,
  Filter,
  X,
  User,
  Edit,
  Trash2,
} from 'lucide-react'

const Promotions = () => {
  // States for form fields
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [isActive, setIsActive] = useState(true)
  const [items, setItems] = useState([{ menu_item_id: '', quantity: 1 }])
  const [menuItems, setMenuItems] = useState([]) // State to store menu items

  // Function to fetch menu items
  const fetchMenuItems = async () => {
    try {
      const response = await axios.get('http://127.0.0.1:8080/api/menu', {
        params: {
          action: 'getAll', // Get all menu items
        },
      })
      setMenuItems(response.data) // Set menu items into state
    } catch (error) {
      console.error('Error fetching menu items:', error)
    }
  }

  // Fetch menu items when component mounts
  useEffect(() => {
    fetchMenuItems()
  }, [])

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault()

    const promotionData = {
      name,
      description,
      start_date: startDate,
      end_date: endDate,
      is_active: isActive,
      items,
    }

    try {
      const response = await axios.post(
        'http://127.0.0.1:8080/api/promotions',
        promotionData,
        {
          headers: {
            'Content-Type': 'application/json',
          },
        }
      )

      if (response.status === 200) {
        alert('โปรโมชั่นถูกสร้างสำเร็จ')
        console.log(response.data)
      } else {
        alert('เกิดข้อผิดพลาด: ' + response.data.error)
      }
    } catch (error) {
      console.error('Error submitting form:', error)
      alert('เกิดข้อผิดพลาดในการสร้างโปรโมชั่น')
    }
  }

  // Handle input changes for item list
  const handleItemChange = (index, e) => {
    const newItems = [...items]
    newItems[index][e.target.name] = e.target.value
    setItems(newItems)
  }

  // Handle adding/removing items
  const handleAddItem = () => {
    setItems([...items, { menu_item_id: '', quantity: 1 }])
  }

  const handleRemoveItem = (index) => {
    const newItems = items.filter((_, i) => i !== index)
    setItems(newItems)
  }

  return (
    <div className="max-w-3xl mx-auto p-6 bg-white rounded shadow-lg">
      <h2 className="text-2xl font-semibold mb-4">เพิ่มโปรโมชั่น</h2>
      <form onSubmit={handleSubmit}>
        <div className="mb-4">
          <label className="block text-gray-700">ชื่อโปรโมชั่น</label>
          <input
            type="text"
            className="w-full px-4 py-2 border border-gray-300 rounded"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>

        <div className="mb-4">
          <label className="block text-gray-700">คำอธิบาย</label>
          <textarea
            className="w-full px-4 py-2 border border-gray-300 rounded"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            required
          />
        </div>

        <div className="mb-4 grid grid-cols-2 gap-4">
          <div>
            <label className="block text-gray-700">วันที่เริ่มต้น</label>
            <input
              type="date"
              className="w-full px-4 py-2 border border-gray-300 rounded"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              required
            />
          </div>

          <div>
            <label className="block text-gray-700">วันที่สิ้นสุด</label>
            <input
              type="date"
              className="w-full px-4 py-2 border border-gray-300 rounded"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              required
            />
          </div>
        </div>

        <div className="mb-4">
          <label className="block text-gray-700">สถานะโปรโมชั่น</label>
          <input
            type="checkbox"
            checked={isActive}
            onChange={() => setIsActive(!isActive)}
          />{' '}
          เปิดใช้งาน
        </div>

        <div className="mb-4">
          <h3 className="text-lg font-semibold">
            รายการสินค้าที่ร่วมโปรโมชั่น
          </h3>
          {items.map((item, index) => (
            <div key={index} className="mb-4 flex space-x-4">
              <div>
                <label className="block text-gray-700">เมนู</label>
                <select
                  name="menu_item_id"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={item.menu_item_id}
                  onChange={(e) => handleItemChange(index, e)}
                  required
                >
                  <option value="">เลือกเมนู</option>
                  {menuItems.map((menu) => (
                    <option key={menu.ID} value={menu.ID}>
                      {menu.Name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-gray-700">จำนวน</label>
                <input
                  type="number"
                  name="quantity"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={item.quantity}
                  onChange={(e) => handleItemChange(index, e)}
                  required
                />
              </div>
              <button
                type="button"
                className="text-red-500"
                onClick={() => handleRemoveItem(index)}
              >
                ลบ
              </button>
            </div>
          ))}
          <button
            type="button"
            className="text-blue-500"
            onClick={handleAddItem}
          >
            เพิ่มรายการสินค้า
          </button>
        </div>

        <div className="mt-6">
          <button
            type="submit"
            className="w-full py-2 px-4 bg-blue-600 text-white rounded"
          >
            สร้างโปรโมชั่น
          </button>
        </div>
      </form>
    </div>
  )
}

export default Promotions
