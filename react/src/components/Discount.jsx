import React, { useState, useEffect } from 'react'

const Discount = () => {
  const [discountTypes, setDiscountTypes] = useState([]) // รายการประเภทส่วนลด
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingDiscountType, setEditingDiscountType] = useState(null)
  const [formData, setFormData] = useState({
    name: '',
    value: 0,
    type: 'percentage', // หรือ 'amount'
    isActive: true,
  })

  const token = localStorage.getItem('token') || ''
  if (!token) {
    window.location.href = '/login' 
  }
  
  // ดึงข้อมูลประเภทส่วนลดทั้งหมดจาก API
  useEffect(() => {
    const fetchDiscountTypes = async () => {
      try {
        const token = localStorage.getItem('token')
        if (!token) throw new Error('No token found')
  
        const response = await fetch(
          'http://localhost:8080/api/payment/discount-types',
          {
            method: 'GET',
            headers: {
              Accept: 'application/json',
              Authorization: `Bearer ${token}`,
            },
          }
        )
  
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`)
        }
  
        const data = await response.json()
        setDiscountTypes(data)
      } catch (error) {
        console.error('Error fetching discount types:', error)
      }
    }
  
    fetchDiscountTypes()
  }, [])

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target
    setFormData((prev) => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }))
  }

  const resetForm = () => {
    setFormData({
      name: '',
      value: 0,
      type: 'percentage',
      isActive: true,
    })
    setEditingDiscountType(null)
  }

  const handleAddNew = () => {
    setIsModalOpen(true)
    resetForm()
  }

  const handleEdit = (discountType) => {
    setEditingDiscountType(discountType)
    setFormData({
      name: discountType.Name,
      value: discountType.Value,
      type: discountType.Type,
      isActive: discountType.IsActive, // จัดการสถานะ isActive ให้ถูกต้องในฟอร์ม
    })
    setIsModalOpen(true)
  }

  // ฟังก์ชันสำหรับลบประเภทส่วนลด
  const handleDelete = async (id) => {
    if (window.confirm('คุณแน่ใจหรือไม่ที่จะลบประเภทส่วนลดนี้?')) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(
          `http://localhost:8080/api/payment/discount-types/${id}`,
          {
            method: 'DELETE',
            headers: { Accept: 'application/json',
              Authorization: `Bearer ${token}`,},
          }
        )

        if (response.ok) {
          const result = await response.json()
          alert(result.message) // แสดงข้อความจาก API ("Discount type deleted successfully")
          setDiscountTypes((prev) =>
            prev.filter((discountType) => discountType.ID !== id)
          )
        } else {
          alert('ไม่สามารถลบประเภทส่วนลดได้')
        }
      } catch (error) {
        console.error('Error deleting discount type:', error)
        alert('เกิดข้อผิดพลาดในการลบประเภทส่วนลด')
      }
    }
  }

  // ฟังก์ชันสำหรับ toggle การเปิดใช้งาน
  const toggleActiveStatus = async (id, currentStatus) => {
    const updatedStatus = !currentStatus

    try {
      const token = localStorage.getItem('token')
      const response = await fetch(
        `http://localhost:8080/api/payment/discount-types/${id}`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            is_active: updatedStatus, // ส่งข้อมูลสถานะ is_active ที่อัปเดต
            name: '', // ชื่อที่ไม่เปลี่ยนแปลง
            type: '', // ประเภทที่ไม่เปลี่ยนแปลง
            value: 0, // ค่า value ที่ไม่เปลี่ยนแปลง
          }),
        }
      )

      if (response.ok) {
        const updatedDiscountType = await response.json()
        setDiscountTypes((prev) =>
          prev.map((discountType) =>
            discountType.ID === id
              ? { ...discountType, IsActive: updatedStatus }
              : discountType
          )
        )
      } else {
        alert('เกิดข้อผิดพลาดในการอัปเดตสถานะการเปิดใช้งาน')
      }
    } catch (error) {
      console.error('Error updating active status:', error)
      alert('เกิดข้อผิดพลาดในการอัปเดตสถานะการเปิดใช้งาน')
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()

    const newDiscountType = {
      name: formData.name,
      type: formData.type,
      value: Number(formData.value),
      is_active: formData.isActive, // ส่งข้อมูล is_active
    }

    const method = editingDiscountType ? 'PUT' : 'POST'
    const url = editingDiscountType
      ? `http://localhost:8080/api/payment/discount-types/${editingDiscountType.ID}`
      : 'http://localhost:8080/api/payment/discount-types'

    try {
      const token = localStorage.getItem('token')
      const response = await fetch(url, {
        method: method,
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(newDiscountType),
      })

      if (response.ok) {
        const data = await response.json()

        setDiscountTypes((prev) => {
          if (editingDiscountType) {
            return prev.map((dt) =>
              dt.ID === editingDiscountType.ID ? data : dt
            )
          }
          return [...prev, data]
        })

        setIsModalOpen(false)
        resetForm()
      } else {
        alert('ไม่สามารถบันทึกข้อมูลประเภทส่วนลด')
      }
    } catch (error) {
      console.error('Error submitting discount type:', error)
    }
  }

  return (
    <div>
      {/* Header */}
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">
            จัดการประเภทส่วนลด
          </h1>
          <p className="text-gray-600 mt-2">
            จัดการประเภทส่วนลดสำหรับระบบการชำระเงิน
          </p>
        </div>
        <button
          onClick={handleAddNew}
          className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 flex items-center gap-2"
        >
          <span className="text-xl">+</span>
          เพิ่มประเภทส่วนลดใหม่
        </button>
      </div>

      {/* Discount Type Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {discountTypes.map((discountType) => (
          <div
            key={discountType.ID}
            className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow"
          >
            <div className="flex items-start justify-between">
              <div>
                <h3 className="text-xl font-bold text-gray-900">
                  {discountType.ID}
                  {discountType.Name}
                </h3>
                <p className="mt-2 text-gray-600">
                  ส่วนลด: {discountType.Value}{' '}
                  {discountType.Type === 'percentage' ? '%' : 'บาท'}
                </p>
              </div>
              <div
                className={`px-3 py-1 rounded-full text-sm font-medium ${
                  discountType.IsActive
                    ? 'bg-green-100 text-green-800'
                    : 'bg-gray-100 text-gray-800'
                }`}
              >
                {discountType.IsActive ? 'เปิดใช้งาน' : 'ปิดใช้งาน'}
              </div>
            </div>

            <div className="flex items-center justify-between mt-4">
              <div className="flex gap-2">
                <button
                  onClick={() => handleEdit(discountType)}
                  className="px-3 py-1 border border-gray-300 rounded-md hover:bg-gray-50"
                >
                  แก้ไข
                </button>
                <button
                  onClick={() => handleDelete(discountType.ID)}
                  className="px-3 py-1 border border-red-300 text-red-600 rounded-md hover:bg-red-50"
                >
                  ลบ
                </button>
                {/* ปุ่มสำหรับเปิด/ปิดใช้งาน */}
                <button
                  onClick={() =>
                    toggleActiveStatus(discountType.ID, discountType.IsActive)
                  }
                  className="px-3 py-1 border border-blue-300 text-blue-600 rounded-md hover:bg-blue-50"
                >
                  {discountType.IsActive ? 'ปิดใช้งาน' : 'เปิดใช้งาน'}
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-xl font-bold mb-4">
              {editingDiscountType
                ? 'แก้ไขประเภทส่วนลด'
                : 'เพิ่มประเภทส่วนลดใหม่'}
            </h2>

            <form onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    ชื่อประเภทส่วนลด
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
                    จำนวนส่วนลด
                  </label>
                  <input
                    type="number"
                    name="value"
                    value={formData.value}
                    onChange={handleInputChange}
                    className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
                    required
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    ประเภทส่วนลด
                  </label>
                  <select
                    name="type"
                    value={formData.type}
                    onChange={handleInputChange}
                    className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2"
                  >
                    <option value="percentage">เปอร์เซ็นต์</option>
                    <option value="amount">จำนวนเงิน</option>
                  </select>
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
                  {editingDiscountType ? 'บันทึกการแก้ไข' : 'เพิ่มประเภทส่วนลด'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

export default Discount
