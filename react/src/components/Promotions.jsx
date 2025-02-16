import React, { useState, useEffect } from 'react'
import axios from 'axios'
import { Plus, X, Ellipsis, Trash2, Edit, Image, Search, Calendar, Tag, Clock } from 'lucide-react'

const API_BASE_URL_PROMOTIONS = 'http://127.0.0.1:8080/api/promotions'
const API_BASE_URL_MENU = 'http://127.0.0.1:8080/api/menu'

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
  const [isDropdownOpen, setIsDropdownOpen] = useState(null)
  const [isEditMode, setIsEditMode] = useState(false)

  // สถานะสำหรับฟอร์มแก้ไข
  const [editFormData, setEditFormData] = useState({
    name: '',
    nameEn: '',
    nameCh: '',
    description: '',
    descriptionEn: '',
    descriptionCh: '',
    startDate: '',
    endDate: '',
    price: 0,
    maxSelections: 0,
    minSelections: 0,
    items: []
  })

  const token = localStorage.getItem('token')

  // ฟังก์ชันดึงข้อมูลเมนูทั้งหมด
  const fetchMenuItems = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL_MENU}`, {
        headers: {
          Authorization: `Bearer ${token}`,
          accept: 'application/json'
        },
        params: {
          action: 'getAll'
        }
      })
      setMenuItems(response.data)
    } catch (error) {
      console.error('Error fetching menu items:', error)
      setError('ไม่สามารถดึงข้อมูลเมนูได้')
    }
  }

  // ใช้ useEffect เพื่อดึงข้อมูลเมื่อโหลด component
  useEffect(() => {
    const fetchData = async () => {
      try {
        await Promise.all([
          fetchMenuItems(),
          fetchPromotions(),
          fetchPromotionsActive()
        ])
      } catch (error) {
        console.error('Error fetching data:', error)
        setError('เกิดข้อผิดพลาดในการดึงข้อมูล')
      }
    }
    fetchData()
  }, [])

  // ฟังก์ชันดึงข้อมูลโปรโมชั่นทั้งหมด
  const fetchPromotions = async () => {
    try {
      const response = await axios.get(API_BASE_URL_PROMOTIONS, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
      setPromotions(response.data)
    } catch (error) {
      console.error('Error fetching promotions:', error)
      alert('เกิดข้อผิดพลาดในการดึงข้อมูลโปรโมชั่น')
    }
  }

  const fetchPromotionsActive = async () => {
    try {
      const response = await fetch(
        `${API_BASE_URL_PROMOTIONS}/Active`,
        {
          method: 'GET',
          headers: {
            accept: 'application/json',
            Authorization: `Bearer ${token}`,
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

  // ฟังก์ชันจัดการการแก้ไขข้อมูลในฟอร์ม
  const handleEditFormChange = (e) => {
    const { name, value } = e.target
    setEditFormData(prev => ({
      ...prev,
      [name]: value
    }))
  }

  // ฟังก์ชันเปิดโหมดแก้ไข
  const handleEdit = (promotion) => {
    setSelectedPromotion(promotion)
    setEditFormData({
      name: promotion.Name,
      nameEn: promotion.NameEn,
      nameCh: promotion.NameCh,
      description: promotion.Description,
      descriptionEn: promotion.DescriptionEn,
      descriptionCh: promotion.DescriptionCh,
      startDate: new Date(promotion.StartDate).toISOString().split('T')[0],
      endDate: new Date(promotion.EndDate).toISOString().split('T')[0],
      price: promotion.Price,
      maxSelections: promotion.MaxSelections,
      minSelections: promotion.MinSelections,
      items: promotion.Items.map(item => ({
        menu_item_id: item.MenuItem.ID,
        quantity: item.Quantity,
        ID: item.ID,
        MenuItem: item.MenuItem
      }))
    })
    setIsEditMode(true)
    setIsDropdownOpen(null)
  }

  // ฟังก์ชันบันทึกการแก้ไข
  const handleUpdatePromotion = async (e) => {
    e.preventDefault()
    try {
      // อัพเดทข้อมูลพื้นฐานของโปรโมชัน
      const response = await axios.put(
        `${API_BASE_URL_PROMOTIONS}/${selectedPromotion.ID}`,
        {
          name: editFormData.name,
          nameEn: editFormData.nameEn,
          nameCh: editFormData.nameCh,
          description: editFormData.description,
          descriptionEn: editFormData.descriptionEn,
          descriptionCh: editFormData.descriptionCh,
          start_date: new Date(editFormData.startDate).toISOString(),
          end_date: new Date(editFormData.endDate).toISOString(),
          price: parseFloat(editFormData.price),
          max_selections: parseInt(editFormData.maxSelections),
          min_selections: parseInt(editFormData.minSelections)
        },
        {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
        }
      )

      if (response.status === 200) {
        // เพิ่มรายการอาหารใหม่ (ถ้ามี)
        const newItems = editFormData.items.filter(item => !item.ID)
        if (newItems.length > 0) {
          const itemsToAdd = newItems.map(item => ({
            menu_item_id: parseInt(item.menu_item_id),
            quantity: parseInt(item.quantity)
          }))
          
          await axios.post(
            `${API_BASE_URL_PROMOTIONS}/${selectedPromotion.ID}/items`,
            { items: itemsToAdd },
            {
              headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
              },
            }
          )
        }

        alert('อัพเดทโปรโมชั่นสำเร็จ')
        setIsEditMode(false)
        setSelectedPromotion(null)
        fetchPromotions() // รีเฟรชข้อมูลทั้งหมด
      }
    } catch (error) {
      console.error('Error updating promotion:', error)
      alert(error.response?.data?.error || 'เกิดข้อผิดพลาดในการอัพเดทโปรโมชั่น')
    }
  }

  // ฟังก์ชันลบรายการอาหารในโปรโมชั่น
  const handleDeletePromotionItem = async (promoId, itemId) => {
    if (window.confirm('คุณแน่ใจหรือไม่ว่าต้องการลบรายการอาหารนี้?')) {
      try {
        const response = await axios.delete(
          `${API_BASE_URL_PROMOTIONS}/${promoId}/items/${itemId}`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        )
        if (response.status === 200) {
          alert('ลบรายการอาหารสำเร็จ')
          // อัพเดท state โดยตรง
          setPromotions(prevPromotions =>
            prevPromotions.map(promo => {
              if (promo.ID === promoId) {
                return {
                  ...promo,
                  Items: promo.Items.filter(item => item.ID !== itemId),
                  TotalItems: promo.TotalItems - 1
                }
              }
              return promo
            })
          )
          // อัพเดท editFormData ถ้าอยู่ในโหมดแก้ไข
          if (isEditMode && selectedPromotion?.ID === promoId) {
            setEditFormData(prev => ({
              ...prev,
              items: prev.items.filter(item => item.ID !== itemId)
            }))
          }
        }
      } catch (error) {
        console.error('Error deleting promotion item:', error)
        if (error.response?.status === 404) {
          alert('ไม่พบรายการอาหารที่ระบุ')
        } else {
          alert('เกิดข้อผิดพลาดในการลบรายการอาหาร')
        }
      }
    }
  }

  // ฟังก์ชันเพิ่มรายการอาหารในโปรโมชั่น
  const handleAddPromotionItems = async (promoId, newItems) => {
    try {
      const response = await axios.post(
        `${API_BASE_URL_PROMOTIONS}/${promoId}/items`,
        { items: newItems },
        {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
        }
      )

      if (response.status === 200) {
        setPromotions(prevPromotions =>
          prevPromotions.map(promo => {
            if (promo.ID === promoId) {
              return {
                ...promo,
                Items: [...promo.Items, ...response.data.Items],
                TotalItems: response.data.TotalItems
              }
            }
            return promo
          })
        )
        // อัพเดท editFormData ถ้าอยู่ในโหมดแก้ไข
        if (isEditMode && selectedPromotion?.ID === promoId) {
          setEditFormData(prev => ({
            ...prev,
            items: [...prev.items, ...response.data.Items]
          }))
        }
        alert('เพิ่มรายการอาหารสำเร็จ')
      }
    } catch (error) {
      console.error('Error adding promotion items:', error)
      if (error.response?.status === 404) {
        alert('ไม่พบโปรโมชันที่ระบุ')
      } else if (error.response?.status === 400) {
        alert(error.response.data.error || 'ข้อมูลไม่ถูกต้อง')
      } else {
        alert('เกิดข้อผิดพลาดในการเพิ่มรายการอาหาร')
      }
    }
  }

  // ฟังก์ชันปิดโหมดแก้ไข
  const handleCancelEdit = () => {
    setIsEditMode(false)
    setSelectedPromotion(null)
    setEditFormData({
      name: '',
      nameEn: '',
      nameCh: '',
      description: '',
      descriptionEn: '',
      descriptionCh: '',
      startDate: '',
      endDate: '',
      price: 0,
      maxSelections: 0,
      minSelections: 0,
      items: []
    })
  }

  // ฟังก์ชันจัดการการเปลี่ยนแปลงของ input ทั่วไป
  const handleInputChange = (e) => {
    const { name, value } = e.target // ดึงชื่อและค่าจาก input
    setPromotionData((prevData) => ({ ...prevData, [name]: value }))
  }

  // ฟังก์ชันจัดการการเปลี่ยนแปลงในรายการสินค้า
  const handleItemChange = (index, e) => {
    const { name, value } = e.target
    
    setPromotionData(prevData => {
      const updatedItems = [...prevData.items]
      
      // สร้าง item ใหม่ถ้ายังไม่มี
      if (!updatedItems[index]) {
        updatedItems[index] = { menu_item_id: '', quantity: 1 }
      }
      
      // อัพเดทค่าใน item
      updatedItems[index] = {
        ...updatedItems[index],
        [name]: name === 'menu_item_id' ? parseInt(value) : value
      }
      
      return {
        ...prevData,
        items: updatedItems
      }
    })
  }

  // ฟังก์ชันอัพเดทสถานะการใช้งานโปรโมชั่น
  const toggleStatus = async (id, currentStatus) => {
    try {
      const newStatus = !currentStatus // สลับสถานะของโปรโมชั่น
      const response = await axios.patch(
        `${API_BASE_URL_PROMOTIONS}/status/${id}`,
        {
          is_active: newStatus,
        },
        {
          headers: {
            accept: 'application/json',
            Authorization: `Bearer ${token}`,
          },
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

      if (!items || items.length === 0) {
        alert('กรุณาเพิ่มรายการสินค้าอย่างน้อย 1 รายการ')
        return
      }

      // ตรวจสอบเมนูซ้ำ
      const menuIds = items.map(item => item.menu_item_id)
      const hasDuplicates = menuIds.length !== new Set(menuIds).size
      if (hasDuplicates) {
        alert('ไม่สามารถเพิ่มเมนูซ้ำในโปรโมชันเดียวกันได้')
        return
      }

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
        `${API_BASE_URL_PROMOTIONS}`,
        promotionPayload,
        {
          headers: { accept: 'application/json',
           Authorization: `Bearer ${token}` },
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
        `${API_BASE_URL_PROMOTIONS}/image/${selectedPromotion.ID}`,
        formData,
        {
          headers: {
            'Content-Type': 'multipart/form-data',
            Authorization: `Bearer ${token}`,
            accept: 'application/json',
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

  // ฟังก์ชันจัดการการเปลี่ยนแปลงในรายการสินค้าสำหรับโหมดแก้ไข
  const handleEditItemChange = (index, e) => {
    const { name, value } = e.target
    
    setEditFormData(prevData => {
      const updatedItems = [...prevData.items]
      
      // สร้าง item ใหม่ถ้ายังไม่มี
      if (!updatedItems[index]) {
        updatedItems[index] = { menu_item_id: '', quantity: 1 }
      }
      
      // อัพเดทค่าใน item
      updatedItems[index] = {
        ...updatedItems[index],
        [name]: name === 'menu_item_id' ? parseInt(value) : parseInt(value)
      }
      
      return {
        ...prevData,
        items: updatedItems
      }
    })
  }

  // ฟังก์ชันเพิ่มรายการสินค้าในโหมดแก้ไข
  const handleAddEditItem = () => {
    setEditFormData(prevData => ({
      ...prevData,
      items: [...prevData.items, { menu_item_id: '', quantity: 1 }]
    }))
  }

  // ฟังก์ชันลบรายการสินค้าในโหมดแก้ไข
  const handleRemoveEditItem = (index) => {
    setEditFormData(prevData => ({
      ...prevData,
      items: prevData.items.filter((_, i) => i !== index)
    }))
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
                        onClick={() => {
                          setSelectedPromotion(promotion)
                          setShowDeleteConfirm(true)
                        }}
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
                  
                  {/* เพิ่มส่วนแสดงรายการอาหาร */}
                  <div className="mt-4">
                    <h4 className="font-medium text-gray-700 mb-2">รายการอาหาร</h4>
                    <div className="space-y-2">
                      {promotion.Items?.map((item) => (
                        <div key={item.ID} className="flex items-center justify-between bg-gray-50 p-2 rounded-lg">
                          <div className="flex items-center gap-2">
                            <span className="text-sm text-gray-600">{item.MenuItem.Name}</span>
                            <span className="text-xs text-gray-500">x{item.Quantity}</span>
                          </div>
                          <button
                            onClick={() => handleDeletePromotionItem(promotion.ID, item.ID)}
                            className="p-1 text-red-500 hover:text-red-700 transition-colors"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      ))}
                    </div>
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
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => {
                          if (promotionData.items.length > 0) {
                            handleAddPromotionItems(selectedPromotionId, promotionData.items)
                          } else {
                            alert('กรุณาเพิ่มรายการอาหารอย่างน้อย 1 รายการ')
                          }
                        }}
                        className="flex items-center px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                      >
                        <Plus className="w-4 h-4 mr-2" />
                        บันทึกรายการใหม่
                      </button>
                      <button
                        type="button"
                        onClick={handleAddItem}
                        className="flex items-center px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
                      >
                        <Plus className="w-4 h-4 mr-2" />
                        เพิ่มรายการ
                      </button>
                    </div>
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
      {isEditMode && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-4xl mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex justify-between items-center mb-6 border-b pb-4">
                <h2 className="text-xl font-bold text-gray-800">
                  แก้ไขโปรโมชัน
                </h2>
                <button
                  onClick={handleCancelEdit}
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
                        value={editFormData.name}
                        onChange={handleEditFormChange}
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
                        value={editFormData.nameEn}
                        onChange={handleEditFormChange}
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
                        value={editFormData.nameCh}
                        onChange={handleEditFormChange}
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
                        type="date"
                        name="startDate"
                        value={editFormData.startDate}
                        onChange={handleEditFormChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        required
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        วันที่สิ้นสุด
                      </label>
                      <input
                        type="date"
                        name="endDate"
                        value={editFormData.endDate}
                        onChange={handleEditFormChange}
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
                        value={editFormData.price}
                        onChange={handleEditFormChange}
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
                        value={editFormData.description}
                        onChange={handleEditFormChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="3"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (อังกฤษ)
                      </label>
                      <textarea
                        name="descriptionEn"
                        value={editFormData.descriptionEn}
                        onChange={handleEditFormChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="3"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (จีน)
                      </label>
                      <textarea
                        name="descriptionCh"
                        value={editFormData.descriptionCh}
                        onChange={handleEditFormChange}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="3"
                      />
                    </div>
                  </div>
                </div>

                {/* Items */}
                <div className="bg-gray-50 p-4 rounded-lg">
                  <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold text-gray-700">
                      รายการอาหาร
                    </h3>
                    <button
                      type="button"
                      onClick={handleAddEditItem}
                      className="flex items-center px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
                    >
                      <Plus className="w-4 h-4 mr-2" />
                      เพิ่มรายการ
                    </button>
                  </div>

                  <div className="space-y-4">
                    {editFormData.items.map((item, index) => (
                      <div
                        key={index}
                        className="flex items-center gap-4 bg-white p-4 rounded-lg border border-gray-200"
                      >
                        <div className="flex-1">
                          <label className="block text-sm font-medium text-gray-700 mb-1">
                            เมนู
                          </label>
                          {item.ID ? (
                            // ถ้าเป็นรายการที่มีอยู่แล้ว
                            <div className="flex items-center justify-between">
                              <span className="text-gray-700">{item.MenuItem.Name}</span>
                              <span className="text-sm text-gray-500">x{item.quantity}</span>
                            </div>
                          ) : (
                            // ถ้าเป็นรายการใหม่
                            <select
                              name="menu_item_id"
                              value={item.menu_item_id}
                              onChange={(e) => handleEditItemChange(index, e)}
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
                          )}
                        </div>
                        {!item.ID && (
                          <div className="w-32">
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                              จำนวน
                            </label>
                            <input
                              type="number"
                              name="quantity"
                              value={item.quantity}
                              onChange={(e) => handleEditItemChange(index, e)}
                              className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                              min="1"
                              required
                            />
                          </div>
                        )}
                        <button
                          type="button"
                          onClick={() => item.ID ? 
                            handleDeletePromotionItem(selectedPromotion.ID, item.ID) : 
                            handleRemoveEditItem(index)}
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
                    onClick={handleCancelEdit}
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
