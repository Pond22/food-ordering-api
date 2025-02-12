import React, { useState, useEffect } from 'react'
import axios from 'axios'
import { Plus, X, Ellipsis, Trash2, Edit, Image, Search, Calendar, Tag, Clock } from 'lucide-react'

const Promotions = () => {
  const [promotionData, setPromotionData] = useState({
    name: '',
    nameEn: '',
    nameCh: '',
    description: '',
    startDate: '',
    endDate: '',
    price: '',
    isActive: true,
    maxSelections: 0, // เพิ่มฟิลด์ใหม่
    minSelections: 0, // เพิ่มฟิลด์ใหม่
    items: [{ menu_item_id: '', quantity: 1 }],
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

  // แก้ไขฟังก์ชัน handleSubmit
  const handleSubmit = async (e) => {
    e.preventDefault()

    try {
      const {
        name,
        nameEn,
        nameCh,
        description,
        startDate,
        endDate,
        price,
        maxSelections,
        minSelections,
        items,
      } = promotionData

      // ตรวจสอบว่ามีการเลือกรายการสินค้าอย่างน้อย 1 รายการ
      if (!items || items.length === 0) {
        alert('กรุณาเพิ่มรายการสินค้าอย่างน้อย 1 รายการ')
        return
      }

      // ตรวจสอบว่าทุกรายการมีการเลือกเมนูและระบุจำนวน
      const isValidItems = items.every(
        (item) => item.menu_item_id && item.quantity > 0
      )
      if (!isValidItems) {
        alert('กรุณาระบุเมนูและจำนวนให้ครบทุกรายการ')
        return
      }

      const promotionPayload = {
        name,
        nameEn,
        nameCh,
        description,
        start_date: new Date(startDate).toISOString(),
        end_date: new Date(endDate).toISOString(),
        price: parseFloat(price),
        max_selections: parseInt(maxSelections, 10),
        min_selections: parseInt(minSelections, 10),
        items: items.map((item) => ({
          menu_item_id: parseInt(item.menu_item_id, 10),
          quantity: parseInt(item.quantity, 10),
        })),
      }

      console.log('Sending payload:', promotionPayload) // เพิ่ม log เพื่อตรวจสอบข้อมูล

      const response = await axios.post(
        'http://localhost:8080/api/promotions',
        promotionPayload,
        {
          headers: { 'Content-Type': 'application/json' },
        }
      )

      if (response.status === 201 || response.status === 200) {
        alert('เพิ่มโปรโมชันสำเร็จ')
        setIsPopupOpen(false)
        resetPromotionData()
        fetchPromotions()
      }
    } catch (error) {
      console.error('Error creating promotion:', error)
      const errorMessage = error.response?.data?.message || 'เกิดข้อผิดพลาดในการเพิ่มโปรโมชัน'
      alert(errorMessage)
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
    if (!selectedPromotion || !selectedImage) {
      alert('กรุณาเลือกไฟล์รูปภาพ')
      return
    }

    const formData = new FormData()
    formData.append('image', selectedImage)

    try {
      const response = await axios.put(
        `http://localhost:8080/api/promotions/image/${selectedPromotion.ID}`,
        formData,
        {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        }
      )

      if (response.status === 200) {
        alert('อัปเดตรูปภาพสำเร็จ')
        setIsImageUploadPopupOpen(false)
        fetchPromotions() // รีเฟรชข้อมูลหลังอัพเดท
        setSelectedImage(null)
      } else {
        alert('เกิดข้อผิดพลาดในการอัปเดตรูปภาพ')
      }
    } catch (error) {
      console.error('Error uploading image:', error)
      alert('เกิดข้อผิดพลาดในการอัปเดตรูปภาพ')
    }
  }

  const handleUpdatePromotion = async (e) => {
    e.preventDefault()

    if (!selectedPromotionId) {
      alert('ไม่พบข้อมูลโปรโมชั่นที่ต้องการแก้ไข')
      return
    }

    try {
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

      const formattedItems = items.map((item) => ({
        menu_item_id: parseInt(item.menu_item_id, 10),
        quantity: parseInt(item.quantity, 10),
      }))

      const promotionPayload = {
        name,
        nameEn,
        nameCh,
        description,
        descriptionEn,
        descriptionCh,
        start_date,
        end_date,
        price: parseFloat(price),
        is_active: isActive,
        items: formattedItems,
      }

      const response = await axios.put(
        `http://localhost:8080/api/promotions/${selectedPromotionId}`,
        promotionPayload,
        {
          headers: { 'Content-Type': 'application/json' },
        }
      )

      if (response.status === 200) {
        alert('แก้ไขโปรโมชั่นสำเร็จ')
        setIsEditPopupOpen(false)
        fetchPromotions() // รีเฟรชข้อมูลหลังแก้ไข
      } else {
        alert(`เกิดข้อผิดพลาด: ${response.data.error}`)
      }
    } catch (error) {
      console.error('Error updating promotion:', error)
      alert('เกิดข้อผิดพลาดในการแก้ไขโปรโมชั่น')
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
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 p-6">
      {/* Header Section */}
      <div className="mb-8">
        <div className="bg-white rounded-xl shadow-lg p-6">
          <h1 className="text-2xl font-bold text-gray-800 mb-2">
            จัดการโปรโมชัน
          </h1>
          <p className="text-gray-600">
            จัดการรายการโปรโมชัน เพิ่ม แก้ไข และปรับแต่งรายละเอียดต่างๆ
          </p>
        </div>
      </div>

      {/* Search and Filter Section */}
      <div className="bg-white rounded-xl shadow-lg p-6 mb-6">
        <div className="flex flex-wrap gap-4 items-center justify-between">
          <div className="flex-1 min-w-[300px]">
            <div className="relative">
              <Search
                className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400"
                size={20}
              />
              <input
                type="text"
                className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                placeholder="ค้นหาโปรโมชัน..."
                value={searchTerm}
                onChange={handleSearchChange}
              />
            </div>
          </div>
          <button
            onClick={togglePopup}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-all duration-200 shadow-md hover:shadow-lg flex items-center gap-2"
          >
            <Plus size={20} />
            เพิ่มโปรโมชันใหม่
          </button>
        </div>
      </div>

      {/* Promotions List */}
      <div className="bg-white rounded-xl shadow-lg overflow-hidden">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 p-6">
          {filteredPromotions.map((promotion) => (
            <div
              key={promotion.ID}
              className="bg-white rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-all duration-200"
            >
              {/* Promotion Image */}
              <div className="relative h-48 bg-gray-100 rounded-t-lg overflow-hidden">
                {promotion.Image ? (
                  <img
                    src={`data:image/png;base64,${promotion.Image}`}
                    alt={promotion.Name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <Image className="w-12 h-12 text-gray-400" />
                  </div>
                )}
                <div className="absolute top-4 right-4">
                  <button
                    onClick={() => toggleDropdown(promotion.ID)}
                    className="p-2 bg-white rounded-full shadow-md hover:bg-gray-50 transition-colors"
                  >
                    <Ellipsis className="w-5 h-5 text-gray-600" />
                  </button>
                  {isDropdownOpen === promotion.ID && (
                    <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-xl border border-gray-100 z-10">
                      <button
                        onClick={() => handleEdit(promotion)}
                        className="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2"
                      >
                        <Edit className="w-4 h-4" />
                        แก้ไขโปรโมชัน
                      </button>
                      <button
                        onClick={() => {
                          setSelectedPromotion(promotion)
                          setIsImageUploadPopupOpen(true)
                        }}
                        className="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2"
                      >
                        <Image className="w-4 h-4" />
                        อัปโหลดรูปภาพ
                      </button>
                      <button
                        onClick={() => confirmDelete(promotion.ID)}
                        className="w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
                      >
                        <Trash2 className="w-4 h-4" />
                        ลบโปรโมชัน
                      </button>
                    </div>
                  )}
                </div>
              </div>

              {/* Promotion Details */}
              <div className="p-4">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-lg font-semibold text-gray-800">
                    {promotion.Name}
                  </h3>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      className="sr-only peer"
                      checked={promotion.IsActive}
                      onChange={() =>
                        toggleStatus(promotion.ID, promotion.IsActive)
                      }
                    />
                    <div className="w-11 h-6 bg-gray-200 rounded-full peer peer-focus:ring-4 peer-focus:ring-blue-300 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-500"></div>
                  </label>
                </div>

                <div className="space-y-2 mb-4">
                  <p className="text-sm text-gray-600">{promotion.NameEn}</p>
                  <p className="text-sm text-gray-600">{promotion.NameCh}</p>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm text-gray-600">
                    <Calendar className="w-4 h-4" />
                    <span>
                      {new Date(promotion.StartDate).toLocaleDateString()} -{' '}
                      {new Date(promotion.EndDate).toLocaleDateString()}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 text-sm text-gray-600">
                    <Tag className="w-4 h-4" />
                    <span>{promotion.Price.toLocaleString()} ฿</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm text-gray-600">
                    <Clock className="w-4 h-4" />
                    <span>
                      {promotion.IsActive ? 'กำลังใช้งาน' : 'ปิดใช้งาน'}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 text-sm text-gray-600">
                    <span>
                      จำนวนรายการที่เลือกได้: {promotion.MinSelections} -{' '}
                      {promotion.MaxSelections}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Add/Edit Promotion Modal */}
      {isPopupOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-4xl mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex justify-between items-center mb-6 border-b pb-4">
                <h2 className="text-xl font-bold text-gray-800">
                  เพิ่มโปรโมชันใหม่
                </h2>
                <button
                  onClick={togglePopup}
                  className="text-gray-400 hover:text-gray-600 transition-colors"
                >
                  <X size={24} />
                </button>
              </div>

              <form onSubmit={handleSubmit} className="space-y-6">
                {/* Basic Information */}
                <div className="bg-gray-50 p-4 rounded-lg space-y-4">
                  <h3 className="text-lg font-semibold text-gray-700">
                    ข้อมูลพื้นฐาน
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ชื่อโปรโมชัน (ไทย)
                      </label>
                      <input
                        type="text"
                        name="name"
                        value={promotionData.name}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        required
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ชื่อโปรโมชัน (อังกฤษ)
                      </label>
                      <input
                        type="text"
                        name="nameEn"
                        value={promotionData.nameEn}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ชื่อโปรโมชัน (จีน)
                      </label>
                      <input
                        type="text"
                        name="nameCh"
                        value={promotionData.nameCh}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        วันที่เริ่มต้น
                      </label>
                      <input
                        type="datetime-local"
                        name="startDate"
                        value={promotionData.startDate}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        required
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        วันที่สิ้นสุด
                      </label>
                      <input
                        type="datetime-local"
                        name="endDate"
                        value={promotionData.endDate}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        required
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ราคา
                      </label>
                      <input
                        type="number"
                        name="price"
                        value={promotionData.price}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        required
                      />
                    </div>
                  </div>
                </div>

                {/* Description */}
                <div className="bg-gray-50 p-4 rounded-lg space-y-4">
                  <h3 className="text-lg font-semibold text-gray-700">
                    คำอธิบาย
                  </h3>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (ไทย)
                      </label>
                      <textarea
                        name="description"
                        value={promotionData.description}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="2"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (อังกฤษ)
                      </label>
                      <textarea
                        name="descriptionEn"
                        value={promotionData.descriptionEn}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="2"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (จีน)
                      </label>
                      <textarea
                        name="descriptionCh"
                        value={promotionData.descriptionCh}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="2"
                      />
                    </div>
                  </div>
                </div>

                {/* เพิ่มฟิลด์ใหม่ */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      จำนวนรายการขั้นต่ำที่เลือกได้
                    </label>
                    <input
                      type="number"
                      name="minSelections"
                      value={promotionData.minSelections}
                      onChange={handleInputChange}
                      className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      min="0"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      จำนวนรายการสูงสุดที่เลือกได้
                    </label>
                    <input
                      type="number"
                      name="maxSelections"
                      value={promotionData.maxSelections}
                      onChange={handleInputChange}
                      className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      min="0"
                    />
                  </div>
                </div>

                {/* Items */}
                <div className="bg-gray-50 p-4 rounded-lg">
                  <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold text-gray-700">
                      รายการสินค้า
                    </h3>
                    <button
                      type="button"
                      onClick={handleAddItem}
                      className="flex items-center px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
                    >
                      <Plus className="w-4 h-4 mr-2" />
                      เพิ่มรายการ
                    </button>
                  </div>

                  <div className="space-y-4">
                    {promotionData.items.map((item, index) => (
                      <div
                        key={index}
                        className="flex items-center gap-4 bg-white p-4 rounded-lg border border-gray-200"
                      >
                        <div className="flex-1">
                          <label className="block text-sm font-medium text-gray-700 mb-1">
                            เมนู
                          </label>
                          <select
                            name="menu_item_id"
                            value={item.menu_item_id}
                            onChange={(e) => handleItemChange(index, e)}
                            className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            required
                          >
                            <option value="">เลือกเมนู</option>
                            {menuItems.map((menuItem) => (
                              <option key={menuItem.ID} value={menuItem.ID}>
                                {menuItem.Name}
                              </option>
                            ))}
                          </select>
                        </div>
                        <div className="w-32">
                          <label className="block text-sm font-medium text-gray-700 mb-1">
                            จำนวน
                          </label>
                          <input
                            type="number"
                            name="quantity"
                            value={item.quantity}
                            onChange={(e) => handleItemChange(index, e)}
                            className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            min="1"
                            required
                          />
                        </div>
                        <button
                          type="button"
                          onClick={() => handleRemoveItem(index)}
                          className="self-end p-2 text-red-500 hover:text-red-700 transition-colors"
                        >
                          <Trash2 className="w-5 h-5" />
                        </button>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="flex justify-end space-x-4 pt-4 border-t">
                  <button
                    type="button"
                    onClick={() => {
                      resetPromotionData()
                      togglePopup()
                    }}
                    className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                  >
                    ยกเลิก
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                  >
                    บันทึก
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Image Upload Modal */}
      {isImageUploadPopupOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full mx-4">
            <div className="p-6">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-xl font-bold text-gray-800">
                  อัพโหลดรูปภาพ
                </h2>
                <button
                  onClick={() => setIsImageUploadPopupOpen(false)}
                  className="text-gray-400 hover:text-gray-600 transition-colors"
                >
                  <X size={24} />
                </button>
              </div>

              <div className="space-y-4">
                <div className="border-2 border-dashed border-gray-300 rounded-lg p-4 text-center">
                  <input
                    type="file"
                    onChange={handleImageChange}
                    className="hidden"
                    id="imageInput"
                  />
                  <label htmlFor="imageInput" className="cursor-pointer block">
                    {selectedImage ? (
                      <img
                        src={selectedImage}
                        alt="Preview"
                        className="max-w-full h-48 object-contain mx-auto"
                      />
                    ) : (
                      <div className="py-8">
                        <Image className="mx-auto h-12 w-12 text-gray-400" />
                        <p className="mt-2 text-sm text-gray-600">
                          คลิกเพื่อเลือกรูปภาพ
                        </p>
                      </div>
                    )}
                  </label>
                </div>
              </div>

              <div className="mt-6 flex justify-end gap-3">
                <button
                  className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
                  onClick={() => setIsImageUploadPopupOpen(false)}
                >
                  ยกเลิก
                </button>
                <button
                  className="px-4 py-2 text-white bg-blue-500 rounded-lg hover:bg-blue-600 transition-colors"
                  onClick={handleImageUpload}
                >
                  อัพโหลด
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Edit Promotion Modal */}
      {isEditPopupOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-4xl mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex justify-between items-center mb-6 border-b pb-4">
                <h2 className="text-xl font-bold text-gray-800">
                  แก้ไขโปรโมชัน
                </h2>
                <button
                  onClick={() => setIsEditPopupOpen(false)}
                  className="text-gray-400 hover:text-gray-600 transition-colors"
                >
                  <X size={24} />
                </button>
              </div>

              <form onSubmit={handleUpdatePromotion} className="space-y-6">
                {/* Basic Information */}
                <div className="bg-gray-50 p-4 rounded-lg space-y-4">
                  <h3 className="text-lg font-semibold text-gray-700">
                    ข้อมูลพื้นฐาน
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ชื่อโปรโมชัน (ไทย)
                      </label>
                      <input
                        type="text"
                        name="name"
                        value={promotionData.name}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        required
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ชื่อโปรโมชัน (อังกฤษ)
                      </label>
                      <input
                        type="text"
                        name="nameEn"
                        value={promotionData.nameEn}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ชื่อโปรโมชัน (จีน)
                      </label>
                      <input
                        type="text"
                        name="nameCh"
                        value={promotionData.nameCh}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        วันที่เริ่มต้น
                      </label>
                      <input
                        type="datetime-local"
                        name="startDate"
                        value={promotionData.startDate}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        required
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        วันที่สิ้นสุด
                      </label>
                      <input
                        type="datetime-local"
                        name="endDate"
                        value={promotionData.endDate}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        required
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ราคา
                      </label>
                      <input
                        type="number"
                        name="price"
                        value={promotionData.price}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        required
                      />
                    </div>
                  </div>
                </div>

                {/* Description */}
                <div className="bg-gray-50 p-4 rounded-lg space-y-4">
                  <h3 className="text-lg font-semibold text-gray-700">
                    คำอธิบาย
                  </h3>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (ไทย)
                      </label>
                      <textarea
                        name="description"
                        value={promotionData.description}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="2"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (อังกฤษ)
                      </label>
                      <textarea
                        name="descriptionEn"
                        value={promotionData.descriptionEn}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="2"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (จีน)
                      </label>
                      <textarea
                        name="descriptionCh"
                        value={promotionData.descriptionCh}
                        onChange={handleInputChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="2"
                      />
                    </div>
                  </div>
                </div>

                {/* Items */}
                <div className="bg-gray-50 p-4 rounded-lg">
                  <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold text-gray-700">
                      รายการสินค้า
                    </h3>
                    <button
                      type="button"
                      onClick={handleAddItem}
                      className="flex items-center px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
                    >
                      <Plus className="w-4 h-4 mr-2" />
                      เพิ่มรายการ
                    </button>
                  </div>

                  <div className="space-y-4">
                    {promotionData.items.map((item, index) => (
                      <div
                        key={index}
                        className="flex items-center gap-4 bg-white p-4 rounded-lg border border-gray-200"
                      >
                        <div className="flex-1">
                          <label className="block text-sm font-medium text-gray-700 mb-1">
                            เมนู
                          </label>
                          <select
                            name="menu_item_id"
                            value={item.menu_item_id}
                            onChange={(e) => handleItemChange(index, e)}
                            className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            required
                          >
                            <option value="">เลือกเมนู</option>
                            {menuItems.map((menuItem) => (
                              <option key={menuItem.ID} value={menuItem.ID}>
                                {menuItem.Name}
                              </option>
                            ))}
                          </select>
                        </div>
                        <div className="w-32">
                          <label className="block text-sm font-medium text-gray-700 mb-1">
                            จำนวน
                          </label>
                          <input
                            type="number"
                            name="quantity"
                            value={item.quantity}
                            onChange={(e) => handleItemChange(index, e)}
                            className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            min="1"
                            required
                          />
                        </div>
                        <button
                          type="button"
                          onClick={() => handleRemoveItem(index)}
                          className="self-end p-2 text-red-500 hover:text-red-700 transition-colors"
                        >
                          <Trash2 className="w-5 h-5" />
                        </button>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="flex justify-end space-x-4 pt-4 border-t">
                  <button
                    type="button"
                    onClick={() => setIsEditPopupOpen(false)}
                    className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                  >
                    ยกเลิก
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                  >
                    บันทึก
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-xl p-6 max-w-md w-full mx-4">
            <h2 className="text-xl font-bold text-gray-800 mb-4">
              ยืนยันการลบ
            </h2>
            <p className="text-gray-600 mb-6">
              คุณแน่ใจหรือไม่ว่าต้องการลบโปรโมชันนี้?
            </p>
            <div className="flex justify-end gap-3">
              <button
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
                onClick={() => setShowDeleteConfirm(false)}
              >
                ยกเลิก
              </button>
              <button
                className="px-4 py-2 text-white bg-red-500 rounded-lg hover:bg-red-600 transition-colors"
                onClick={() => {
                  handleDelete(selectedPromotionId)
                  setShowDeleteConfirm(false)
                }}
              >
                ลบ
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Promotions
