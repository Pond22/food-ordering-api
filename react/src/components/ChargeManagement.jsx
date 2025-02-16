import React, { useState, useEffect } from 'react'
import Discount from './Discount'

const API_BASE_URL = 'http://127.0.0.1:8080/api/payment'

const ChargeTypeManagement = () => {
  const [chargeTypes, setChargeTypes] = useState([]) // รายการประเภทค่าใช้จ่าย
  const [filteredChargeTypes, setFilteredChargeTypes] = useState([]) // ข้อมูลที่กรองตามคำค้นหา
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingChargeType, setEditingChargeType] = useState(null)
  const [activeTab, setActiveTab] = useState('Discount') // เปลี่ยนค่าเป็น 'Discount'
  const [searchQuery, setSearchQuery] = useState('') // ค่าค้นหาจากผู้ใช้
  const [refreshTrigger, setRefreshTrigger] = useState(0) // เพิ่ม state สำหรับ trigger การรีเฟรช
  const [formData, setFormData] = useState({
    name: '',
    default_amount: 0,
    isActive: true,
  })

  // ดึงข้อมูลประเภทค่าใช้จ่ายทั้งหมดจาก API
  useEffect(() => {
    const token = localStorage.getItem('token')
    fetch(`${API_BASE_URL}/charge-types`, {
      method: 'GET',
      headers: {
        Accept: 'application/json',
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })
      .then((response) => response.json())
      .then((data) => {
        // เรียงข้อมูลตาม ID จากน้อยไปมาก
        const sortedData = data.sort((a, b) => a.ID - b.ID)
        setChargeTypes(sortedData)
        setFilteredChargeTypes(sortedData)
      })
      .catch((error) => console.error('Error fetching charge types:', error))
  }, [refreshTrigger])

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target

    // อัปเดตข้อมูลใน formData เมื่อมีการเปลี่ยนแปลง
    setFormData((prev) => ({
      ...prev,
      [name]:
        type === 'checkbox'
          ? checked
          : name === 'defaultAmount'
          ? isNaN(parseFloat(value)) // ตรวจสอบว่าค่าเป็นตัวเลขหรือไม่
            ? 0 // ถ้าไม่ใช่ตัวเลข จะตั้งค่าเป็น 0
            : parseFloat(value) // แปลงเป็นตัวเลข
          : value,
    }))
  }

  const resetForm = () => {
    setFormData({
      name: '',
      defaultAmount: 0,
      isActive: true,
    })
    setEditingChargeType(null)
  }

  const handleAddNew = () => {
    setIsModalOpen(true)
  }

  const handleEdit = (chargeType) => {
    setEditingChargeType(chargeType)
    setFormData({
      name: chargeType.Name, // ตั้งชื่อที่แก้ไข
      defaultAmount: chargeType.DefaultAmount, // ตั้งจำนวนเริ่มต้นที่แก้ไข
      isActive: chargeType.IsActive, // ตั้งสถานะเปิดใช้งานที่แก้ไข
    })
    setIsModalOpen(true) // เปิด Modal สำหรับแก้ไข
  }

  // ฟังก์ชันสำหรับลบประเภทค่าใช้จ่าย
  // แก้ไขฟังก์ชัน handleDelete
  const handleDelete = async (id) => {
    if (window.confirm('คุณแน่ใจหรือไม่ที่จะลบประเภทค่าใช้จ่ายนี้?')) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(`${API_BASE_URL}/charge-types/${id}`, {
          method: 'DELETE',
          headers: {
            Accept: 'application/json',
            Authorization: `Bearer ${token}`,
          },
        })

        if (response.ok) {
          alert('ลบประเภทค่าใช้จ่ายสำเร็จ')
          setRefreshTrigger((prev) => prev + 1) // trigger การรีเฟรช
        } else {
          alert('ไม่สามารถลบประเภทค่าใช้จ่ายได้')
        }
      } catch (error) {
        console.error('Error deleting charge type:', error)
        alert('เกิดข้อผิดพลาดในการลบประเภทค่าใช้จ่าย')
      }
    }
  }

  // ฟังก์ชันสำหรับ toggle การเปิดใช้งาน
  const toggleActiveStatus = (id, currentStatus) => {
    const updatedStatus = !currentStatus // เปลี่ยนสถานะเปิด/ปิดใช้งาน
    const token = localStorage.getItem('token')
    fetch(`${API_BASE_URL}/charge-types/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
        Accept: 'application/json',
      },
      body: JSON.stringify({
        is_active: updatedStatus, // เปลี่ยนเป็น is_active ตามที่ API ต้องการ
      }),
    })
      .then((response) => response.json())
      .then((data) => {
        // อัปเดตสถานะ isActive ในรายการ chargeTypes
        setChargeTypes((prev) =>
          prev.map((chargeType) =>
            chargeType.ID === id
              ? { ...chargeType, IsActive: updatedStatus }
              : chargeType
          )
        )
      })
      .catch((error) => {
        console.error('Error updating active status:', error)
        alert('เกิดข้อผิดพลาดในการอัปเดตสถานะการเปิดใช้งาน')
      })
  }

  const handleSubmit = async (e) => {
    e.preventDefault()

    // ข้อมูลที่ต้องการบันทึกหรือแก้ไข
    const updatedChargeType = {
      name: formData.name, // ชื่อประเภทค่าใช้จ่าย
      defaultAmount: formData.defaultAmount, // ใช้ defaultAmount สำหรับ POST หรือ PUT API
      isActive: formData.isActive, // ใช้ isActive
    }

    const method = editingChargeType ? 'PUT' : 'POST' // กำหนดวิธีการ (POST สำหรับเพิ่ม, PUT สำหรับแก้ไข)
    const url = editingChargeType
      ? `${API_BASE_URL}/charge-types/${editingChargeType.ID}` // ถ้าเป็นการแก้ไขข้อมูล
      : `${API_BASE_URL}/charge-types` // ถ้าเป็นการเพิ่มข้อมูลใหม่

    try {
      const token = localStorage.getItem('token')
      const response = await fetch(url, {
        method: method,
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(updatedChargeType),
      })

      if (response.ok) {
        const data = await response.json()

        // อัปเดตข้อมูลใน state
        setChargeTypes((prev) => {
          if (editingChargeType) {
            // ถ้าเป็นการแก้ไขข้อมูล ให้แทนที่ข้อมูลเดิม
            return prev.map((ct) =>
              ct.ID === editingChargeType.ID ? data : ct
            )
          }
          // ถ้าเป็นการเพิ่มข้อมูลใหม่ ให้เพิ่มเข้าไป
          return [...prev, data]
        })

        // อัปเดตข้อมูลใน filteredChargeTypes
        setFilteredChargeTypes((prev) => {
          if (editingChargeType) {
            return prev.map((ct) =>
              ct.ID === editingChargeType.ID ? data : ct
            )
          }
          return [...prev, data]
        })

        setIsModalOpen(false) // ปิด Modal
        resetForm() // รีเซ็ตฟอร์ม
      } else {
        alert('ไม่สามารถบันทึกข้อมูลประเภทค่าใช้จ่าย')
      }
    } catch (error) {
      console.error('Error submitting charge type:', error)
      alert('เกิดข้อผิดพลาดในการบันทึกข้อมูลประเภทค่าใช้จ่าย')
    }
  }

  // ฟังก์ชันสำหรับการค้นหาข้อมูล
  const handleSearch = (e) => {
    const query = e.target.value.toLowerCase()
    setSearchQuery(query)

    if (query === '') {
      setFilteredChargeTypes(chargeTypes) // ถ้าไม่มีคำค้นหา ให้แสดงข้อมูลทั้งหมด
    } else {
      const filtered = chargeTypes.filter(
        (chargeType) =>
          chargeType.Name.toLowerCase().includes(query) ||
          chargeType.ID.toString().includes(query)
      )
      setFilteredChargeTypes(filtered)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8 ">
      <div className="max-w-6xl mx-auto">
        <div>
          {/* ปุ่มเปลี่ยนตาราง */}
          <button
            className={`px-4 py-2 rounded-t-lg ${
              activeTab === 'Discount'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
            onClick={() => setActiveTab('Discount')}
          >
            ส่วนลด
          </button>
          <button
            className={`px-4 py-2 rounded-t-lg ${
              activeTab === 'ChargeType'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
            onClick={() => setActiveTab('ChargeType')}
          >
            ค่าใช้จ่ายเพิ่มเติม
          </button>
        </div>
        {activeTab === 'Discount' && (
          <div>
            <Discount /> {/* เรียกใช้งานคอมโพเนนต์ Discount */}
          </div>
        )}

        {/* Header */}

        {activeTab === 'ChargeType' && (
          <div className="flex justify-between items-center mb-8">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                จัดการประเภทค่าใช้จ่าย
              </h1>
              <p className="text-gray-600 mt-2">
                จัดการประเภทค่าใช้จ่ายสำหรับระบบการชำระเงิน
              </p>
            </div>
            <button
              onClick={handleAddNew}
              className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 flex items-center gap-2"
            >
              <span className="text-xl">+</span>
              เพิ่มประเภทค่าใช้จ่ายใหม่
            </button>
          </div>
        )}

        {/* ช่องค้นหาข้อมูล */}
        {activeTab === 'ChargeType' && (
          <div className="mb-6">
            <input
              type="text"
              value={searchQuery}
              onChange={handleSearch}
              placeholder="ค้นหาตาม ID หรือ ชื่อ"
              className="p-2 border border-gray-300 rounded-md w-full"
            />
          </div>
        )}

        {/* แสดงข้อมูลตามแท็บ */}
        {activeTab === 'ChargeType' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {filteredChargeTypes.map((chargeType) => (
              <div
                key={chargeType.ID}
                className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow"
              >
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="text-xl font-bold text-gray-900">
                      {/* {chargeType.ID} */}
                      {chargeType.Name}
                    </h3>
                    <p className="mt-2 text-gray-600">
                      จำนวนเริ่มต้น: {chargeType.DefaultAmount}
                    </p>
                  </div>
                  <div
                    className={`px-3 py-1 rounded-full text-sm font-medium ${
                      chargeType.IsActive
                        ? 'bg-green-100 text-green-800'
                        : 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {chargeType.IsActive ? 'เปิดใช้งาน' : 'ปิดใช้งาน'}
                  </div>
                </div>

                <div className="flex items-center justify-between mt-4">
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleEdit(chargeType)}
                      className="px-3 py-1 border border-gray-300 rounded-md hover:bg-gray-50"
                    >
                      แก้ไข
                    </button>
                    <button
                      onClick={() => handleDelete(chargeType.ID)}
                      className="px-3 py-1 border border-red-300 text-red-600 rounded-md hover:bg-red-50"
                    >
                      ลบ
                    </button>
                    {/* ปุ่มสำหรับเปิด/ปิดใช้งาน */}
                    <button
                      onClick={() =>
                        toggleActiveStatus(chargeType.ID, chargeType.IsActive)
                      }
                      className="px-3 py-1 border border-blue-300 text-blue-600 rounded-md hover:bg-blue-50"
                    >
                      {chargeType.IsActive ? 'ปิดใช้งาน' : 'เปิดใช้งาน'}
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Modal */}
        {isModalOpen && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
            <div className="bg-white rounded-lg p-6 max-w-md w-full">
              <h2 className="text-xl font-bold mb-4">
                {editingChargeType
                  ? 'แก้ไขประเภทค่าใช้จ่าย'
                  : 'เพิ่มประเภทค่าใช้จ่ายใหม่'}
              </h2>

              <form onSubmit={handleSubmit}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      ชื่อประเภทค่าใช้จ่าย
                    </label>
                    <input
                      type="text"
                      name="name"
                      value={formData.name}
                      onChange={handleInputChange}
                      className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
                      required
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      จำนวนเริ่มต้น
                    </label>
                    <input
                      type="number"
                      name="defaultAmount"
                      value={formData.defaultAmount}
                      onChange={handleInputChange}
                      className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
                      required
                      min="0" // กำหนดให้ไม่สามารถป้อนค่าต่ำกว่า 0
                    />
                  </div>

                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      name="isActive"
                      checked={formData.isActive}
                      onChange={handleInputChange}
                      className="h-4 w-4 text-indigo-600 border-gray-300 rounded"
                    />
                    <label className="ml-2 text-sm text-gray-700">
                      เปิดใช้งาน
                    </label>
                  </div>
                </div>

                <div className="mt-6 flex justify-end space-x-3">
                  <button
                    type="button"
                    onClick={() => setIsModalOpen(false)}
                    className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
                  >
                    ยกเลิก
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
                  >
                    {editingChargeType
                      ? 'บันทึกการแก้ไข'
                      : 'เพิ่มประเภทค่าใช้จ่าย'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default ChargeTypeManagement
