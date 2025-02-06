import React, { useState, useEffect } from 'react'
import axios from 'axios'
import { Plus, X, Ellipsis, Trash2, Edit, Image } from 'lucide-react'

const Promotions = () => {
  const [promotionData, setPromotionData] = useState({
    name: '', // ชื่อโปรโมชั่น (ภาษาไทย)
    nameEn: '', // ชื่อโปรโมชั่น (ภาษาอังกฤษ)
    nameCh: '', // ชื่อโปรโมชั่น (ภาษาจีน)
    description: '', // คำอธิบายโปรโมชั่น (ภาษาไทย)
    descriptionEn: '', // คำอธิบายโปรโมชั่น (ภาษาอังกฤษ)
    descriptionCh: '', // คำอธิบายโปรโมชั่น (ภาษาจีน)
    startDate: '', // วันที่เริ่มต้น
    endDate: '', // วันที่สิ้นสุด
    price: '', // ราคาของโปรโมชั่น
    isActive: true, // สถานะโปรโมชั่น (เปิด/ปิด)
    items: [{ menu_item_id: '', quantity: 1 }], // รายการสินค้าที่ร่วมโปรโมชั่น
  })

  const [menuItems, setMenuItems] = useState([]) // State สำหรับเก็บข้อมูลเมนู
  const [promotions, setPromotions] = useState([]) // State สำหรับเก็บข้อมูลโปรโมชั่น
  const [isEditPopupOpen, setIsEditPopupOpen] = useState(false) // สถานะสำหรับ Popup แก้ไข
  const [isPopupOpen, setIsPopupOpen] = useState(false) // สถานะของ popup
  const [activeTab, setActiveTab] = useState('Promo_used') // state สลับการแสดงตาราง
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false) // State สำหรับแสดง Popup ยืนยันการลบ
  const [selectedPromotionId, setSelectedPromotionId] = useState(null) // State สำหรับเก็บ ID โปรโมชั่นที่เลือก
  const [showConfirm, setShowConfirm] = useState(false)
  const [selectedPromotion, setSelectedPromotion] = useState(null)
  const [isImageUploadPopupOpen, setIsImageUploadPopupOpen] = useState(false) // สถานะของ popup อัปเดตรูป
  const [selectedImage, setSelectedImage] = useState(null) // ตัวแปรเก็บรูปภาพที่เลือก
  const [promotionsActive, setPromotionsActive] = useState([])
  const [error, setError] = useState(null)
  const [searchTerm, setSearchTerm] = useState('') // ฟิลด์ค้นหา
  const [isDropdownOpen, setIsDropdownOpen] = useState(false)

  // ใช้ useEffect เพื่อดึงข้อมูลเมนูเมื่อโหลด component
  useEffect(() => {
    fetchMenuItems()
    fetchPromotions() // ดึงข้อมูลโปรโมชั่น
    fetchPromotionsActive()
  }, [])

  const fetchPromotions = async () => {
    try {
      const response = await axios.get('http://localhost:8080/api/promotions', {
        headers: { accept: 'application/json' },
      })

      if (response.data) {
        const promotions = Array.isArray(response.data)
          ? response.data
          : [response.data]

        // เรียงข้อมูลตาม ID
        promotions.sort((a, b) => a.ID - b.ID)

        setPromotions(promotions)
      } else {
        console.error('โครงสร้างข้อมูลไม่ถูกต้อง', response.data)
      }
    } catch (error) {
      console.error('ไม่สามารถดึงข้อมูลโปรโมชั่นได้:', error)
    }
  }

  const fetchPromotionsActive = async () => {
    try {
      const response = await fetch(
        'http://localhost:8080/api/promotions/Active',
        {
          method: 'GET',
          headers: {
            Accept: 'application/json',
          },
        }
      )

      if (!response.ok) {
        throw new Error('Failed to fetch promotions')
      }

      const data = await response.json()
      setPromotionsActive(data)
    } catch (error) {
      setError(error.message)
    }
  }

  // ฟังก์ชันดึงข้อมูลเมนู
  const fetchMenuItems = async (action = 'getAll', params = {}) => {
    try {
      const response = await axios.get('http://localhost:8080/api/menu', {
        params: {
          action, // ใช้ action ตามที่ต้องการ เช่น getByID, getByCategory, getAll
          ...params, // พารามิเตอร์เพิ่มเติม เช่น category_id หรือ id
        },
      })
      setMenuItems(response.data) // เซ็ตข้อมูลเมนูที่ดึงมา
    } catch (error) {
      console.error('Error fetching menu items:', error)
    }
  }

  // ฟังก์ชันจัดการการเปลี่ยนแปลงของ input ทั่วไป
  const handleInputChange = (e) => {
    const { name, value } = e.target // ดึงชื่อและค่าจาก input
    setPromotionData((prevData) => ({ ...prevData, [name]: value }))
  }

  // ฟังก์ชันจัดการการเปลี่ยนแปลงในรายการสินค้า
  const handleItemChange = (index, e) => {
    const { name, value } = e.target
    const updatedItems = [...promotionData.items]

    updatedItems[index][name] = value // อัพเดตค่าใน state ของ items
    setPromotionData((prevData) => ({ ...prevData, items: updatedItems }))
  }

  // ฟังก์ชันอัพเดทสถานะการใช้งานโปรโมชั่น
  const toggleStatus = async (id, currentStatus) => {
    try {
      const newStatus = !currentStatus // สลับสถานะของโปรโมชั่น
      const response = await axios.patch(
        `http://localhost:8080/api/promotions/status/${id}`,
        {
          is_active: newStatus,
        }
      )

      if (response.status === 200) {
        // อัพเดทสถานะใน state
        setPromotions(
          promotions.map((promo) =>
            promo.ID === id ? { ...promo, IsActive: newStatus } : promo
          )
        )
        alert(`โปรโมชั่น ${newStatus ? 'เปิดใช้งาน' : 'ปิดใช้งาน'} สำเร็จ`)
      }
    } catch (error) {
      console.error('Error updating promotion status:', error)
      alert('เกิดข้อผิดพลาดในการอัพเดทสถานะโปรโมชั่น')
    }
  }

  // ฟังก์ชันเพิ่มรายการสินค้า
  const handleAddItem = () => {
    setPromotionData((prevData) => ({
      ...prevData,
      items: [...prevData.items, { menu_item_id: '', quantity: 1 }],
    }))
  }

  // ฟังก์ชันลบรายการสินค้า
  const handleRemoveItem = (index) => {
    const updatedItems = promotionData.items.filter((_, i) => i !== index)
    setPromotionData((prevData) => ({ ...prevData, items: updatedItems }))
  }

  const handleSubmit = async (e) => {
    e.preventDefault()

    const {
      name,
      nameEn,
      nameCh,
      description,
      descriptionEn,
      descriptionCh,
      startDate,
      endDate,
      price,
      isActive,
      items,
    } = promotionData

    const start_date = new Date(startDate).toISOString()
    const end_date = new Date(endDate).toISOString()

    // ตรวจสอบและแปลงค่าของ `items.quantity` เป็น number
    const formattedItems = items.map((item) => ({
      menu_item_id: parseInt(item.menu_item_id, 10),
      quantity: parseInt(item.quantity, 10), // แปลงค่า quantity เป็น number
    }))

    try {
      const promotionPayload = {
        name,
        nameEn,
        nameCh,
        description,
        descriptionEn,
        descriptionCh,
        start_date,
        end_date,
        price: parseFloat(price), // แปลง price เป็น number
        is_active: isActive,
        items: formattedItems,
      }

      // ส่งข้อมูลไปยัง API
      const response = await axios.post(
        'http://localhost:8080/api/promotions',
        promotionPayload,
        {
          headers: { 'Content-Type': 'application/json' },
        }
      )

      if (response.status === 200) {
        alert('โปรโมชั่นถูกสร้างสำเร็จ')
        console.log(response.data)
      } else {
        alert(`เกิดข้อผิดพลาด: ${response.data.error}`)
      }
    } catch (error) {
      console.error('Error submitting form:', error)

      if (error.response && error.response.data) {
        console.error('API Error:', error.response.data)
        alert(
          `เกิดข้อผิดพลาดในการสร้างโปรโมชั่น: ${
            error.response.data.error || error.message
          }`
        )
      } else {
        alert('เกิดข้อผิดพลาดในการสร้างโปรโมชั่น')
      }
    }
  }

  // ฟังก์ชันเปิด/ปิด popup
  const togglePopup = () => {
    setIsPopupOpen(!isPopupOpen)
  }
  const resetPromotionData = () => {
    setPromotionData({
      name: '',
      nameEn: '',
      nameCh: '',
      description: '',
      descriptionEn: '',
      descriptionCh: '',
      startDate: '',
      endDate: '',
      price: '',
      isActive: false,
      items: [],
    })
  }

  const handleEdit = (promotion) => {
    setSelectedPromotionId(promotion.ID)
    setPromotionData({
      name: promotion.Name,
      nameEn: promotion.NameEn,
      nameCh: promotion.NameCh,
      description: promotion.Description,
      descriptionEn: promotion.DescriptionEn,
      descriptionCh: promotion.DescriptionCh,
      startDate: new Date(promotion.StartDate).toISOString().slice(0, 16), // datetime-local
      endDate: new Date(promotion.EndDate).toISOString().slice(0, 16), // datetime-local
      price: promotion.Price,
      isActive: promotion.IsActive,
      items: promotion.Items.map((item) => ({
        menu_item_id: item.MenuItemID,
        quantity: item.Quantity,
      })),
    })
    setIsEditPopupOpen(true)
  }

  // ฟังก์ชันเปิด/ปิด popup อัปเดตรูป
  const toggleImageUploadPopup = () => {
    setIsImageUploadPopupOpen(!isImageUploadPopupOpen)
  }

  // ฟังก์ชันเลือกไฟล์ภาพ
  const handleImageChange = (e) => {
    const file = e.target.files[0]

    // ตรวจสอบประเภทไฟล์
    if (file && (file.type === 'image/jpeg' || file.type === 'image/png')) {
      // ตรวจสอบขนาดไฟล์ไม่เกิน 5MB
      if (file.size <= 5 * 1024 * 1024) {
        setSelectedImage(file)
      } else {
        alert('ขนาดไฟล์เกิน 5MB')
      }
    } else {
      alert('กรุณาเลือกไฟล์ JPG หรือ PNG เท่านั้น')
    }
  }

  // ฟังก์ชันอัปโหลดรูปภาพ
  const handleImageUpload = async () => {
    if (!selectedImage) {
      alert('กรุณาเลือกไฟล์รูปภาพ')
      return
    }

    const formData = new FormData()
    formData.append('image', selectedImage)

    try {
      const response = await axios.put(
        `http://localhost:8080/api/promotions/image/${selectedPromotionId}`,
        formData,
        {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        }
      )

      if (response.status === 200) {
        alert('อัปเดตรูปภาพสำเร็จ')
        setIsImageUploadPopupOpen(false) // ปิด popup หลังอัปโหลดเสร็จ
      } else {
        alert('เกิดข้อผิดพลาดในการอัปเดตรูปภาพ')
      }
    } catch (error) {
      console.error('Error uploading image:', error)
      alert('เกิดข้อผิดพลาดในการอัปเดตรูปภาพ')
    }
  }

  const confirmDelete = (id) => {
    if (window.confirm('คุณแน่ใจหรือไม่ว่าต้องการลบโปรโมชั่นนี้?')) {
      handleDelete(id)
    }
  }

  const handleDelete = async (id) => {
    try {
      const response = await axios.delete(
        `http://localhost:8080/api/promotions/${id}`
      )
      if (response.status === 200) {
        alert('ลบโปรโมชั่นสำเร็จ')

        // ลบข้อมูลจาก State ทันที
        setPromotions((prevPromotions) =>
          prevPromotions.filter((promotion) => promotion.ID !== id)
        )
      } else {
        alert('เกิดข้อผิดพลาด')
      }
    } catch (error) {
      if (error.response?.status === 404) {
        alert('ไม่พบโปรโมชั่นที่ระบุ')
      } else {
        alert('เกิดข้อผิดพลาดในการลบ')
      }
    }
  }

  // ฟังก์ชันการค้นหา
  const handleSearchChange = (e) => {
    setSearchTerm(e.target.value)
  }

  // ฟังก์ชันกรองข้อมูลตามคำค้นหา
  const filteredPromotions = promotions.filter((promotion) => {
    return (
      promotion.Name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      promotion.ID.toString().includes(searchTerm)
    )
  })

  const toggleDropdown = (ID) => {
    setIsDropdownOpen((prev) => (prev === ID ? null : ID))
  }

  return (
    <div className="w-full mx-auto py-2 bg-white rounded shadow-lg">
      {/* ช่องค้นหา */}
      <div className="flex justify-between my-4 px-4">
        <input
          type="text"
          placeholder="ค้นหาข้อมูลโปรโมชั่น..."
          value={searchTerm}
          onChange={handleSearchChange}
          className="p-2 border border-gray-300 rounded w-2/3"
        />
        <div>
          <button
            type="button"
            onClick={togglePopup}
            className=" w-auto py-2 px-4 bg-green-600 text-white rounded "
          >
            เพิ่มโปรโมชั่นใหม่
          </button>
        </div>
      </div>
      {/* ตารางแสดงโปรโมชั่น */}
      <div className="flex justify-between">
        <div>
          {/* ปุ่มเปลี่ยนตาราง */}
          <button
            className={`px-4 py-2 rounded-t-lg ${
              activeTab === 'Promo_used'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
            onClick={() => setActiveTab('Promo_used')}
          >
            ข้อมูลโปรโมชันทั้งหมด
          </button>
          <button
            className={`px-4 py-2 rounded-t-lg ${
              activeTab === 'Promo_All'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
            onClick={() => setActiveTab('Promo_All')}
          >
            ข้อมูลโปรโมชันที่ใช้งาน
          </button>
        </div>
      </div>
      <div className="">
        {activeTab === 'Promo_used' && (
          <table className="min-w-full w-auto table-auto border-collapse">
            <thead>
              <tr className="text-white bg-blue-500">
                <th className=" px-4 py-2">สถานะ</th>
                <th className=" px-4 py-2 w-1/12">รหัสโปรโมชั่น</th>
                <th className=" px-4 py-2 w-3/12">รูปโปรโมชั่น</th>
                <th className=" px-4 py-2">ชื่อโปรโมชั่น</th>
                <th className=" px-4 py-2 w-2/12">ชื่อโปรโมชั่น</th>
                <th className=" px-4 py-2 w-3/12">สินค้าที่ร่วมรายการ</th>
                <th className="px-4 py-2">ราคา(THB)</th>
                {/* <th className="px-4 py-2">สถานะ</th> */}
                <th className="px-4 py-2">วันที่เริ่มต้น</th>
                <th className="px-4 py-2">วันที่สิ้นสุด</th>
                <th className="px-4 py-2">ดำเนินการ</th>
              </tr>
            </thead>
            <tbody>
              {filteredPromotions.map((promotion) => (
                <tr key={promotion.ID}>
                  <td>
                    <label
                      htmlFor={`toggle-${promotion.ID}`}
                      className="flex items-center cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        id={`toggle-${promotion.ID}`}
                        checked={promotion.IsActive}
                        onChange={() =>
                          toggleStatus(promotion.ID, promotion.IsActive)
                        }
                        className="hidden"
                      />
                      <div
                        className={`w-12 h-6 rounded-full ${
                          promotion.IsActive ? 'bg-green-500' : 'bg-gray-400'
                        } relative`}
                      >
                        <div
                          className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform ${
                            promotion.IsActive ? 'transform translate-x-6' : ''
                          }`}
                        ></div>
                      </div>
                      <span className="ml-2 text-sm">
                        {promotion.IsActive ? 'เปิด' : 'ปิด'}
                      </span>
                    </label>
                  </td>
                  <td className="text-center border p-2">{promotion.ID}</td>
                  <td className="text-center border p-2">
                    {promotion.Image && promotion.Image.length > 0 ? (
                      <img
                        src={`data:image/png;base64,${promotion.Image}`}
                        alt={promotion.Name}
                        className="w-50 h-25 "
                      />
                    ) : (
                      'ไม่มีรูป'
                    )}
                  </td>
                  <td className="pl-2 border">
                    <div>TH: {promotion.Name}</div>
                    <div>EN: {promotion.NameEn}</div>
                    <div>CH: {promotion.NameCh}</div>
                  </td>
                  <td className="pl-2 border w-3/12">
                    <div>TH: {promotion.Description}</div>
                    <div>EN: {promotion.DescriptionEn}</div>
                    <div>CH: {promotion.DescriptionCh}</div>
                  </td>
                  <td className="border px-2">
                    {promotion.Items && promotion.Items.length > 0 ? (
                      promotion.Items.map((item, idx) => (
                        <div key={idx}>
                          •{item.MenuItem?.Name || 'ไม่มีชื่อสินค้า'} (ราคา:{' '}
                          {item.MenuItem?.Price || 0}, จำนวน:{' '}
                          {item.Quantity || 0})
                        </div>
                      ))
                    ) : (
                      <span>ไม่มีสินค้าที่ร่วมรายการ</span>
                    )}
                  </td>
                  <td className="text-center border">{promotion.Price || 0}</td>
                  <td className="text-center border">
                    {promotion.StartDate
                      ? new Date(promotion.StartDate).toLocaleString()
                      : 'ไม่ระบุ'}
                  </td>
                  <td className="text-center border">
                    {promotion.EndDate
                      ? new Date(promotion.EndDate).toLocaleString()
                      : 'ไม่ระบุ'}
                  </td>
                  <td className="flex justify-center border-t">
                    <div className="relative inline-block ">
                      <div className="relative group">
                        <button
                          className="hover:bg-gray-100 p-2 rounded-full"
                          onClick={() => toggleDropdown(promotion.ID)}
                        >
                          <Ellipsis />
                          <div className="absolute left-1/2 -translate-x-1/2 bottom-full mb-2 hidden group-hover:block bg-gray-400 text-white text-sm py-1 px-2 rounded-md shadow-lg">
                            จัดการข้อมูล
                          </div>
                        </button>
                        {isDropdownOpen === promotion.ID && (
                          <div className="absolute right-0 z-10 mt-2 w-48 origin-top-right rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
                            <div className="py-2 text-left px-4">
                              <div className="hover:bg-gray-100 ">
                                <button
                                  onClick={() => handleEdit(promotion)}
                                  className=" "
                                >
                                  <Edit className="inline w-4 h-4 mr-2" />
                                  แก้ไขโปรโมชัน
                                </button>
                              </div>
                              <div className="hover:bg-gray-100 ">
                                <button
                                  onClick={() => {
                                    setSelectedPromotionId(promotion.ID) // กำหนด ID ของโปรโมชั่นที่ต้องการอัปเดตรูป
                                    toggleImageUploadPopup() // เปิด popup อัปเดตรูป
                                  }}
                                >
                                  <Image className="inline w-4 h-4 mr-2" />
                                  อัปโหลดรูปภาพ
                                </button>
                              </div>
                              <div className="hover:bg-gray-100">
                                <button
                                  onClick={() => confirmDelete(promotion.ID)}
                                  className="text-red-600 "
                                >
                                  <Trash2 className="inline w-4 h-4 mr-2" />
                                  ลบ
                                </button>
                              </div>
                            </div>
                          </div>
                        )}
                      </div>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {/* // Popup อัปเดตรูป */}
        {isImageUploadPopupOpen && (
          <div className="fixed inset-0 flex justify-center items-center bg-gray-500 bg-opacity-50">
            <div className="bg-white p-6 rounded shadow-lg">
              <h2 className="text-xl mb-4">อัปเดตรูปภาพโปรโมชั่น</h2>
              <input
                type="file"
                onChange={handleImageChange}
                className="mb-4"
              />
              <div className="flex justify-end">
                <button
                  onClick={toggleImageUploadPopup}
                  className="px-4 py-2 bg-gray-400 text-white rounded mr-2"
                >
                  ยกเลิก
                </button>
                <button
                  onClick={handleImageUpload}
                  className="px-4 py-2 bg-blue-500 text-white rounded"
                >
                  อัปโหลด
                </button>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'Promo_All' && (
          <div className="p-6">
            {error && <p className="text-red-500">Error: {error}</p>}
            <h2 className="text-2xl font-semibold mb-4">Active Promotions</h2>

            {promotionsActive.length > 0 ? (
              <table className="min-w-full table-auto border-collapse border border-gray-300">
                <thead className="bg-gray-100">
                  <tr>
                    <th className="px-4 py-2 border border-gray-300">
                      Promotion Name
                    </th>
                    <th className="px-4 py-2 border border-gray-300">
                      Description
                    </th>
                    <th className="px-4 py-2 border border-gray-300">
                      Start Date
                    </th>
                    <th className="px-4 py-2 border border-gray-300">
                      End Date
                    </th>
                    <th className="px-4 py-2 border border-gray-300">
                      Price (THB)
                    </th>
                    <th className="px-4 py-2 border border-gray-300">Items</th>
                  </tr>
                </thead>
                <tbody>
                  {promotionsActive.map((promotion) => (
                    <tr key={promotion.ID} className="border-b">
                      <td className="px-4 py-2 border border-gray-300">
                        {promotion.Name}
                      </td>
                      <td className="px-4 py-2 border border-gray-300">
                        {promotion.Description || 'No description available'}
                      </td>
                      <td className="px-4 py-2 border border-gray-300">
                        {new Date(promotion.StartDate).toLocaleString()}
                      </td>
                      <td className="px-4 py-2 border border-gray-300">
                        {new Date(promotion.EndDate).toLocaleString()}
                      </td>
                      <td className="px-4 py-2 border border-gray-300">
                        {promotion.Price} THB
                      </td>
                      <td className="px-4 py-2 border border-gray-300">
                        <ul>
                          {promotion.Items.map((item, index) => (
                            <li key={index}>
                              <span>
                                •{item.MenuItem.Name} (Quantity: {item.Quantity}
                                )
                              </span>
                            </li>
                          ))}
                        </ul>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <p>No active promotions available</p>
            )}
          </div>
        )}
      </div>

      {isEditPopupOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center">
          <div className="bg-white p-6 rounded shadow-lg w-auto h-4/5 overflow-y-auto">
            <h2 className="text-2xl font-semibold mb-4">แก้ไขโปรโมชั่น</h2>
            <button
              onClick={() => {
                setIsEditPopupOpen(false)
                resetPromotionData() // เคลียร์ข้อมูลเมื่อปิด Popup
              }}
              className="absolute top-2 right-2 text-gray-500 hover:text-gray-700"
            >
              <X className="w-6 h-6" />
            </button>
            <form
              onSubmit={async (e) => {
                e.preventDefault()
                const updatedPromotion = {
                  name: promotionData.name,
                  nameEn: promotionData.nameEn,
                  nameCh: promotionData.nameCh,
                  description: promotionData.description,
                  descriptionEn: promotionData.descriptionEn,
                  descriptionCh: promotionData.descriptionCh,
                  start_date: new Date(promotionData.startDate).toISOString(),
                  end_date: new Date(promotionData.endDate).toISOString(),
                  price: parseFloat(promotionData.price),
                  is_active: promotionData.isActive,
                  items: promotionData.items.map((item) => ({
                    menu_item_id: parseInt(item.menu_item_id, 10),
                    quantity: parseInt(item.quantity, 10),
                  })),
                }

                try {
                  const response = await axios.put(
                    `http://localhost:8080/api/promotions/${selectedPromotionId}`,
                    updatedPromotion,
                    {
                      headers: { 'Content-Type': 'application/json' },
                    }
                  )

                  if (response.status === 200) {
                    alert('แก้ไขโปรโมชั่นสำเร็จ')
                    setPromotions((prevPromotions) =>
                      prevPromotions.map((promo) =>
                        promo.ID === selectedPromotionId
                          ? { ...promo, ...response.data }
                          : promo
                      )
                    )
                    setIsEditPopupOpen(false)
                    resetPromotionData() // รีเซ็ตข้อมูลหลังการอัปเดต
                  } else {
                    alert(`เกิดข้อผิดพลาด: ${response.data.error}`)
                  }
                } catch (error) {
                  console.error('Error updating promotion:', error)
                  alert('เกิดข้อผิดพลาดในการอัปเดตโปรโมชั่น')
                }
              }}
            >
              {/* ฟอร์มแก้ไขเหมือนฟอร์มเพิ่มโปรโมชั่น */}
              {/* ตัวอย่าง */}
              <div className="mb-4">
                <label className="block text-gray-700">
                  ชื่อโปรโมชั่น (ไทย)
                </label>
                <input
                  type="text"
                  name="name"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.name}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกชื่อโปรโมชั่น ภาษาอังกฤษ */}
              <div className="mb-4">
                <label className="block text-gray-700">
                  ชื่อโปรโมชั่น (อังกฤษ)
                </label>
                <input
                  type="text"
                  name="nameEn"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.nameEn}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกชื่อโปรโมชั่น ภาษาจีน */}
              <div className="mb-4">
                <label className="block text-gray-700">
                  ชื่อโปรโมชั่น (จีน)
                </label>
                <input
                  type="text"
                  name="nameCh"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.nameCh}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกคำอธิบายโปรโมชั่น ภาษาไทย */}
              <div className="mb-4">
                <label className="block text-gray-700">คำอธิบาย (ไทย)</label>
                <textarea
                  name="description"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.description}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกคำอธิบายโปรโมชั่น ภาษาอังกฤษ */}
              <div className="mb-4">
                <label className="block text-gray-700">คำอธิบาย (อังกฤษ)</label>
                <textarea
                  name="descriptionEn"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.descriptionEn}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกคำอธิบายโปรโมชั่น ภาษาจีน */}
              <div className="mb-4">
                <label className="block text-gray-700">คำอธิบาย (จีน)</label>
                <textarea
                  name="descriptionCh"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.descriptionCh}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกวันที่เริ่มต้นและสิ้นสุด */}
              <div className="mb-4 grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-gray-700">วันที่เริ่มต้น</label>
                  <input
                    type="datetime-local" // เปลี่ยนเป็น datetime-local
                    name="startDate"
                    className="w-full px-4 py-2 border border-gray-300 rounded"
                    value={promotionData.startDate}
                    onChange={handleInputChange}
                    required
                  />
                </div>

                <div>
                  <label className="block text-gray-700">วันที่สิ้นสุด</label>
                  <input
                    type="datetime-local" // เปลี่ยนเป็น datetime-local
                    name="endDate"
                    className="w-full px-4 py-2 border border-gray-300 rounded"
                    value={promotionData.endDate}
                    onChange={handleInputChange}
                    required
                  />
                </div>
              </div>

              {/* ช่องกรอกราคาของโปรโมชั่น */}
              <div className="mb-4">
                <label className="block text-gray-700">ราคาโปรโมชั่น</label>
                <input
                  type="number"
                  name="price"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.price}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* สถานะโปรโมชั่น */}
              <div className="mb-4">
                <label className="block text-gray-700">สถานะโปรโมชั่น</label>
                <input
                  type="checkbox"
                  name="isActive"
                  checked={promotionData.isActive}
                  onChange={() =>
                    setPromotionData((prevData) => ({
                      ...prevData,
                      isActive: !prevData.isActive,
                    }))
                  }
                />{' '}
                เปิดใช้งาน
              </div>

              {/* ช่องกรอกรายการสินค้าที่ร่วมโปรโมชั่น */}
              <div className="mb-4">
                <h3 className="text-lg font-semibold">
                  รายการสินค้าที่ร่วมโปรโมชั่น
                </h3>
                {promotionData.items.map((item, index) => (
                  <div key={index} className="mb-4 flex space-x-4">
                    <div>
                      <label className="block text-gray-700">เมนู</label>
                      <select
                        name="menu_item_id"
                        className="w-full px-4 py-2 border border-gray-300 rounded"
                        value={item.menu_item_id}
                        onChange={(e) => handleItemChange(index, e)} // ส่งค่าไปที่ handleItemChange
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
                      <X className="inline" /> ลบ
                    </button>
                  </div>
                ))}
                <button
                  type="button"
                  className="text-blue-500"
                  onClick={handleAddItem}
                >
                  <Plus className="inline" /> เพิ่มรายการสินค้า
                </button>
              </div>
              <div className="mt-6">
                <button
                  type="submit"
                  className="px-4 py-2 bg-green-600 text-white rounded"
                >
                  บันทึกการเปลี่ยนแปลง
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Popup ยืนยันก่อนลบ */}
      {showConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center">
          <div className="bg-white p-6 rounded shadow-lg w-96">
            <h2 className="text-lg font-semibold mb-4">ยืนยันการลบ</h2>
            <p className="mb-4">คุณแน่ใจหรือไม่ว่าต้องการลบโปรโมชั่นนี้?</p>
            <div className="flex justify-end">
              <button
                onClick={() => setShowConfirm(false)}
                className="bg-gray-300 text-black px-4 py-2 rounded hover:bg-gray-400 mr-2"
              >
                ยกเลิก
              </button>
              <button
                onClick={handleDelete}
                className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-700"
              >
                ลบ
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ฟอร์มเพิ่มโปรโมชัน */}
      {isPopupOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center ">
          <div className="bg-white p-6 rounded shadow-lg w-auto h-4/5 overflow-y-auto">
            <h2 className="text-2xl font-semibold mb-4">เพิ่มโปรโมชั่น</h2>
            <button
              onClick={() => {
                setIsPopupOpen(false)
                resetPromotionData() // เคลียร์ข้อมูล
              }}
              className="absolute top-2 right-2 text-gray-500 hover:text-gray-700"
            >
              <X className="w-6 h-6" />
            </button>
            <form onSubmit={handleSubmit}>
              {/* ช่องกรอกชื่อโปรโมชั่น ภาษาไทย */}
              <div className="mb-4">
                <label className="block text-gray-700">
                  ชื่อโปรโมชั่น (ไทย)
                </label>
                <input
                  type="text"
                  name="name"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.name}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกชื่อโปรโมชั่น ภาษาอังกฤษ */}
              <div className="mb-4">
                <label className="block text-gray-700">
                  ชื่อโปรโมชั่น (อังกฤษ)
                </label>
                <input
                  type="text"
                  name="nameEn"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.nameEn}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกชื่อโปรโมชั่น ภาษาจีน */}
              <div className="mb-4">
                <label className="block text-gray-700">
                  ชื่อโปรโมชั่น (จีน)
                </label>
                <input
                  type="text"
                  name="nameCh"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.nameCh}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกคำอธิบายโปรโมชั่น ภาษาไทย */}
              <div className="mb-4">
                <label className="block text-gray-700">คำอธิบาย (ไทย)</label>
                <textarea
                  name="description"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.description}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกคำอธิบายโปรโมชั่น ภาษาอังกฤษ */}
              <div className="mb-4">
                <label className="block text-gray-700">คำอธิบาย (อังกฤษ)</label>
                <textarea
                  name="descriptionEn"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.descriptionEn}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกคำอธิบายโปรโมชั่น ภาษาจีน */}
              <div className="mb-4">
                <label className="block text-gray-700">คำอธิบาย (จีน)</label>
                <textarea
                  name="descriptionCh"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.descriptionCh}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* ช่องกรอกวันที่เริ่มต้นและสิ้นสุด */}
              <div className="mb-4 grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-gray-700">วันที่เริ่มต้น</label>
                  <input
                    type="datetime-local" // เปลี่ยนเป็น datetime-local
                    name="startDate"
                    className="w-full px-4 py-2 border border-gray-300 rounded"
                    value={promotionData.startDate}
                    onChange={handleInputChange}
                    required
                  />
                </div>

                <div>
                  <label className="block text-gray-700">วันที่สิ้นสุด</label>
                  <input
                    type="datetime-local" // เปลี่ยนเป็น datetime-local
                    name="endDate"
                    className="w-full px-4 py-2 border border-gray-300 rounded"
                    value={promotionData.endDate}
                    onChange={handleInputChange}
                    required
                  />
                </div>
              </div>

              {/* ช่องกรอกราคาของโปรโมชั่น */}
              <div className="mb-4">
                <label className="block text-gray-700">ราคาโปรโมชั่น</label>
                <input
                  type="number"
                  name="price"
                  className="w-full px-4 py-2 border border-gray-300 rounded"
                  value={promotionData.price}
                  onChange={handleInputChange}
                  required
                />
              </div>

              {/* สถานะโปรโมชั่น */}
              <div className="mb-4">
                <label className="block text-gray-700">สถานะโปรโมชั่น</label>
                <input
                  type="checkbox"
                  name="isActive"
                  checked={promotionData.isActive}
                  onChange={() =>
                    setPromotionData((prevData) => ({
                      ...prevData,
                      isActive: !prevData.isActive,
                    }))
                  }
                />{' '}
                เปิดใช้งาน
              </div>

              {/* ช่องกรอกรายการสินค้าที่ร่วมโปรโมชั่น */}
              <div className="mb-4">
                <h3 className="text-lg font-semibold">
                  รายการสินค้าที่ร่วมโปรโมชั่น
                </h3>
                {promotionData.items.map((item, index) => (
                  <div key={index} className="mb-4 flex space-x-4">
                    <div>
                      <label className="block text-gray-700">เมนู</label>
                      <select
                        name="menu_item_id"
                        className="w-full px-4 py-2 border border-gray-300 rounded"
                        value={item.menu_item_id}
                        onChange={(e) => handleItemChange(index, e)} // ส่งค่าไปที่ handleItemChange
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
                      <X className="inline" /> ลบ
                    </button>
                  </div>
                ))}
                <button
                  type="button"
                  className="text-blue-500"
                  onClick={handleAddItem}
                >
                  <Plus className="inline" /> เพิ่มรายการสินค้า
                </button>
              </div>

              {/* ปุ่มสำหรับส่งฟอร์ม */}
              <div className="mt-6 flex">
                <button
                  type="button"
                  onClick={() => {
                    setIsPopupOpen(false) // ปิด popup
                    resetPromotionData() // เคลียร์ข้อมูลที่กรอก
                  }}
                  className="w-full py-2 px-4 bg-gray-300 text-gray-700 rounded"
                >
                  ยกเลิก
                </button>
                <button
                  type="submit"
                  className="w-full py-2 px-4 bg-blue-600 text-white rounded"
                >
                  สร้างโปรโมชั่น
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Popup ยืนยันการลบ */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 flex items-center justify-center bg-gray-500 bg-opacity-50 z-50">
          <div className="bg-white p-6 rounded-lg shadow-lg w-1/3">
            <h2 className="text-xl font-semibold mb-4">ยืนยันการลบโปรโมชั่น</h2>
            <p>คุณต้องการลบโปรโมชั่นนี้ใช่หรือไม่?</p>
            <div className="flex justify-end mt-4">
              <button
                onClick={cancelDelete}
                className="px-4 py-2 mr-4 bg-gray-300 text-gray-700 rounded-lg"
              >
                ยกเลิก
              </button>
              <button
                onClick={() => deletePromotion(selectedPromotionId)}
                className="px-4 py-2 bg-red-600 text-white rounded-lg"
              >
                ยืนยัน
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Promotions
