import React, { useState, useEffect } from 'react'
import {
  Image,
  Plus,
  Search,
  Filter,
  X,
  Ellipsis,
  Edit,
  Trash2,
} from 'lucide-react'
import axios from 'axios'
import MenuRestore from './MenuRestore'
import Promotions from './Promotions'

const API_BASE_URL = 'http://127.0.0.1:8080/api/menu'
const API_BASE_URL_CATEGORIES = 'http://127.0.0.1:8080/api/categories' // กำหนด URL ของ API

const MenuManagement = () => {
  const [menus, setMenus] = useState([]) // menus เริ่มต้นเป็นอาร์เรย์
  const [categories, setCategories] = useState([]) // categories เริ่มต้นเป็นอาร์เรย์
  const [showAddMenuModal, setShowAddMenuModal] = useState(false)
  const [showConfirmModal, setShowConfirmModal] = useState(false)
  const [showUploadModal, setShowUploadModal] = useState(false) // state สำหรับแสดง popup upload รูป
  const [selectedImage, setSelectedImage] = useState(null) // state สำหรับเก็บรูปที่เลือก
  const [currentMenuId, setCurrentMenuId] = useState(null) // state สำหรับเก็บ ID เมนูที่ต้องการอัพเดทรูป
  const [loading, setLoading] = useState(false) // state สำหรับโหลด
  const [errorMessage, setErrorMessage] = useState('') // state สำหรับข้อผิดพลาด
  const [groupOptions, setGroupOptions] = useState([])
  const [activeTab, setActiveTab] = useState('menu') // state สลับการแสดงตาราง
  const [menuOptions, setMenuOptions] = useState([])
  const [searchTerm, setSearchTerm] = useState('') // สำหรับค้นหาชื่อเมนู
  const [categoryFilter, setCategoryFilter] = useState('') // สำหรับกรองหมวดหมู่
  const [optionGroupFilter, setOptionGroupFilter] = useState('') // สำหรับกรองกลุ่มตัวเลือก
  const [filteredMenus, setFilteredMenus] = useState(menus) // รายการเมนูที่กรองแล้ว
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const [menuToDelete, setMenuToDelete] = useState(null) // เก็บเมนูที่ต้องการลบ
  const [isDropdownOpen, setIsDropdownOpen] = useState(false)
  const [expandedOptions, setExpandedOptions] = useState({}) // เพิ่ม state สำหรับการจัดการการแสดง/ซ่อนตัวเลือก
  const token = localStorage.getItem('token')

  const toggleDropdown = (ID) => {
    setIsDropdownOpen((prev) => (prev === ID ? null : ID))
  }

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (!event.target.closest('#dropdownMenu')) {
        closeDropdown()
      }
    }

    if (isDropdownOpen) {
      document.addEventListener('click', handleClickOutside)
    }

    return () => {
      document.removeEventListener('click', handleClickOutside)
    }
  }, [isDropdownOpen])

  // ฟังก์ชันกรองข้อมูล
  useEffect(() => {
    let result = menus

    // กรองตามคำค้นหา
    if (searchTerm) {
      result = result.filter((menu) =>
        [menu.Name, menu.NameEn, menu.NameCh, menu.Price.toString()].some(
          (name) => name.toLowerCase().includes(searchTerm.toLowerCase())
        )
      )
    }

    // กรองตามหมวดหมู่
    if (categoryFilter) {
      result = result.filter(
        (menu) => menu.CategoryID === parseInt(categoryFilter)
      )
    }

    // กรองตามกลุ่มตัวเลือก
    if (optionGroupFilter) {
      result = result.filter((menu) =>
        menu.OptionGroups.some(
          (group) => group.ID === parseInt(optionGroupFilter)
        )
      )
    }

    setFilteredMenus(result)
  }, [menus, searchTerm, categoryFilter, optionGroupFilter])

  // เมื่อ showUploadModal เปลี่ยนสถานะ จะรีเซ็ตค่า selectedImage และ currentMenuId
  useEffect(() => {
    if (!showUploadModal) {
      setSelectedImage(null) // เคลียร์ภาพเมื่อปิด popup
      setCurrentMenuId(null) // เคลียร์ ID ของเมนูเมื่อปิด popup
    }
  }, [showUploadModal])

  // ดึงข้อมูลหมวดหมู่
  useEffect(() => {
    fetchCategories()
    fetchMenus('getAll')
  }, [])

  // ฟังก์ชันเปิด/ปิด อาหารหมด
  const toggleMenuStatus = async (menuId) => {
    try {
      const token = localStorage.getItem('token')
      const response = await axios.put(
        `${API_BASE_URL}/status/${menuId}`,
        null, // ส่ง null เป็น data เพราะ API นี้ไม่ต้องการ request body
        {
          headers: {
            accept: 'application/json',
            Authorization: `Bearer ${token}`,
          },
        }
      )

      if (response.status === 200) {
        alert(`เปลี่ยนสถานะของเมนู ${response.data.Name} เรียบร้อย`)
        fetchMenus('getAll') // โหลดข้อมูลใหม่
      } else {
        throw new Error('เกิดข้อผิดพลาดในการเปลี่ยนสถานะ')
      }
    } catch (error) {
      console.error('Error toggling menu status:', error)
      alert('ไม่สามารถเปลี่ยนสถานะได้: ' + error.message)
    }
  }

  // ฟังก์ชันดึงข้อมูลหมวดหมู่ จาก api
  const fetchCategories = async () => {
    try {
      setLoading(true)
      const token = localStorage.getItem('token')
      const response = await axios.get(API_BASE_URL_CATEGORIES, {
        headers: {
          accept: 'application/json',
          Authorization: `Bearer ${token}`, // ส่ง JWT ใน Header
        },
      }) //การดึงข้อมูล categories จาก api เส้นอื่น
      setCategories(response.data)
    } catch (error) {
      setErrorMessage('ไม่สามารถดึงข้อมูลหมวดหมู่ได้')
    } finally {
      setLoading(false)
    }
  } // สิ้นสุดฟังก์ชันดึงข้อมูลหมวดหมู่ จาก api

  // ดึงข้อมูล options
  useEffect(() => {
    if (currentMenuId) {
      fetchGroupOptions(currentMenuId)
    }
  }, [currentMenuId])

  // ฟังก์ชันดึงข้อมูล options จาก api
  const fetchGroupOptions = async (menuId) => {
    try {
      const token = localStorage.getItem('token')
      const response = await axios.get(`${API_BASE_URL}/option-groups/${menuId}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
      if (response.status === 200) {
        setMenuOptions(response.data) // ตั้งค่าเมนู options จาก API
      }
    } catch (error) {
      console.error('Error fetching group options:', error)
    }
  } // สิ้นสุดฟังก์ชันดึงข้อมูล options จาก api

  // ฟังก์ชันช่วยในการหาชื่อหมวดหมู่จาก categories
  const getCategoryNameById = (categoryId) => {
    const category = categories.find((cat) => cat.ID === categoryId)
    return category ? category.Name : 'ไม่พบหมวดหมู่'
  } // สิ้นสุดฟังก์ชันช่วยในการหาชื่อหมวดหมู่จาก categories

  // ฟังก์ชันดึงข้อมูลเมนู จาก api
  const fetchMenus = async (action, id = null) => {
    try {
      let url = `${API_BASE_URL}?action=${action}`
      if (action === 'getByID' && id) {
        url += `&id=${id}`
      }

      const token = localStorage.getItem('token')
      const response = await axios.get(url, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      if (response.status === 200 && Array.isArray(response.data)) {
        const sortedMenus = response.data.sort((a, b) => a.ID - b.ID)
        setMenus(sortedMenus) // ตั้งค่า menus ให้เป็นข้อมูลที่เรียงลำดับแล้ว
      } else {
        console.error('Data is not an array', response.data)
      }
    } catch (error) {
      console.error('Error fetching menus:', error)
    }
  } // สิ้นสุดฟังก์ชันดึงข้อมูลเมนู จาก api

  const handleFileChange = (e) => {
    const file = e.target.files[0]
    if (file) {
      setSelectedImage(URL.createObjectURL(file)) // แสดงตัวอย่างรูปภาพที่เลือก
    }
  }

  const handleUploadImage = async () => {
    if (!selectedImage || !currentMenuId) {
      alert('กรุณาเลือกภาพและเมนูที่ต้องการอัพเดท')
      return
    }

    const formData = new FormData()
    const file = document.querySelector('#imageInput').files[0]
    formData.append('image', file)

    try {
      const response = await axios.put(
        `${API_BASE_URL}/image/${currentMenuId}`,
        formData,
        {
          headers: {
            accept: 'application/json',
            Authorization: `Bearer ${token}`,
            'Content-Type': 'multipart/form-data',
          },
        }
      )

      if (response.status === 200) {
        alert('อัพโหลดรูปภาพสำเร็จ')
        setShowUploadModal(false) // ปิด popup หลังจากอัพโหลดเสร็จ
        fetchMenus('getAll') // ดึงข้อมูลเมนูใหม่
      } else {
        alert('เกิดข้อผิดพลาดในการอัพโหลดรูปภาพ')
      }
    } catch (error) {
      console.error('Error uploading image:', error)
      alert('ไม่สามารถอัพโหลดรูปภาพได้')
    }
  }

  // Object ของฟังก์ชันแก้ไขข้อมูลเมนู
  const [menuDetails, setMenuDetails] = useState({
    ID: '',
    name: '',
    nameEn: '',
    nameCh: '',
    price: 0,
    description: '',
    descriptionEn: '',
    descriptionCh: '',
    categoryId: 0,
    optionGroups: [
      {
        name: '',
        nameEn: '',
        nameCh: '',
        MaxSelections: 1,
        isRequired: false,
        options: [
          {
            name: '',
            nameEn: '',
            nameCh: '',
            price: 0,
          },
        ],
      },
    ], // ใช้ array ว่างเป็นค่าเริ่มต้น
  })

  const [showEditMenuModal, setShowEditMenuModal] = useState(false)

  const handleEditMenu = async () => {
    try {
      // ตรวจสอบข้อมูลที่จำเป็น
      if (!menuDetails.name || !menuDetails.categoryId || menuDetails.price < 0) {
        throw new Error('กรุณากรอกข้อมูลที่จำเป็นให้ครบถ้วน และราคาต้องไม่ติดลบ')
      }

      // 1. อัปเดตข้อมูลเมนูหลัก
      const menuPayload = {
        name: menuDetails.name.trim(),
        name_en: menuDetails.nameEn.trim(),
        name_ch: menuDetails.nameCh.trim(),
        description: menuDetails.description.trim(),
        description_en: menuDetails.descriptionEn.trim(),
        description_ch: menuDetails.descriptionCh.trim(),
        category_id: Number(menuDetails.categoryId),
        price: Number(menuDetails.price),
      }

      const token = localStorage.getItem('token')
      const menuResponse = await axios.put(
        `${API_BASE_URL}/${menuDetails.ID}`,
        menuPayload,
        {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
        }
      )

      if (menuResponse.status !== 200) {
        throw new Error('ไม่สามารถอัปเดตข้อมูลเมนูได้')
      }

      // 2. จัดการข้อมูล Option Groups
      for (const group of menuDetails.optionGroups) {
        const groupPayload = {
          name: group.name.trim(),
          name_en: group.nameEn.trim(),
          name_ch: group.nameCh.trim(),
          MaxSelections: group.MaxSelections || 1,
          is_required: group.isRequired || false,
        }

        if (group.ID) {
          // อัปเดต Option Group ที่มีอยู่
          await axios.put(
            `${API_BASE_URL}/option-groups/${group.ID}`,
            groupPayload,
            {
              headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
              },
            }
          )
        } else {
          // สร้าง Option Group ใหม่
          await axios.post(
            `${API_BASE_URL}/option-groups?menu_id=${menuDetails.ID}`,
            groupPayload,
            {
              headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
              },
            }
          )
        }

        // 3. จัดการข้อมูล Options
        if (group.options && group.options.length > 0) {
          for (const option of group.options) {
            const optionPayload = {
              name: option.name.trim(),
              name_en: option.nameEn.trim(),
              name_ch: option.nameCh.trim(),
              price: Number(option.price),
            }

            if (option.ID) {
              // อัปเดต Option ที่มีอยู่
              await axios.put(
                `${API_BASE_URL}/options/${option.ID}`,
                optionPayload,
                {
                  headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`,
                  },
                }
              )
            } else if (group.ID) {
              // สร้าง Option ใหม่
              await axios.post(
                `${API_BASE_URL}/options?OptionGroupID=${group.ID}`,
                optionPayload,
                {
                  headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`,
                  },
                }
              )
            }
          }
        }
      }

      alert('อัปเดตข้อมูลเมนูสำเร็จ')
      fetchMenus('getAll') // โหลดข้อมูลใหม่
      setShowEditMenuModal(false) // ปิด Modal
    } catch (error) {
      console.error('Error updating menu:', error)
      alert('เกิดข้อผิดพลาดในการอัปเดตข้อมูลเมนู: ' + (error.response?.data?.error || error.message))
    }
  }

  const [isModalOpen, setIsModalOpen] = useState(false) // สถานะการเปิด/ปิด Modal
  const [updatedData, setUpdatedData] = useState({
    name: '',
    name_ch: '',
    name_en: '',
    price: 0,
  }) // ข้อมูลที่ผู้ใช้กรอกเพื่ออัพเดท

  // ฟังก์ชันเปิด Modal และตั้งค่า Option ที่จะถูกแก้ไข
  const handleEditClick = (option) => {
    setSelectedOption(option)
    setUpdatedData({
      name: option.Name,
      name_ch: option.NameCh,
      name_en: option.NameEn,
      price: option.Price,
    })
    setIsModalOpen(true)
  }

  // ฟังก์ชันปิด Modal
  const handleCloseModal = () => {
    setIsModalOpen(false)
    setSelectedOption(null)
  }

  // ฟังก์ชันสำหรับการอัพเดท Option ผ่าน API
  const updateOption = async () => {
    try {
      const token = localStorage.getItem('token')
      const response = await axios.put(
        `${API_BASE_URL}/${selectedOption.GroupID}/options/${selectedOption.ID}`,
        updatedData,
        {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
        }
      )
      if (response.status === 200) {
        alert('Option updated successfully!')
        // รีเฟรชข้อมูลในตาราง
        handleCloseModal()
      } else {
        alert('Failed to update Option')
      }
    } catch (error) {
      console.error('Error updating option:', error)
      alert(
        'Error updating Option: ' +
          (error.response?.data?.error || error.message)
      )
    }
  }

  // ฟังก์ชันสำหรับการจัดการการเปลี่ยนแปลงในฟอร์ม
  const handleInputChange = (e) => {
    const { name, value } = e.target
    setUpdatedData((prevData) => ({
      ...prevData,
      [name]: value,
    }))
  }
  //ฟังก์ชันการลบเมนู
  const deleteMenu = async () => {
    if (!menuToDelete) return

    try {
      const token = localStorage.getItem('token')
      const response = await axios.delete(
        `${API_BASE_URL}/${menuToDelete.ID}`,
        {
          headers: {
            accept: 'application/json',
            Authorization: `Bearer ${token}`, // ส่ง JWT ใน Header
          },
        }
      )
      if (response.status === 200) {
        alert('ลบเมนูสำเร็จ')
        setMenus((prevMenus) =>
          prevMenus.filter((menu) => menu.ID !== menuToDelete.ID)
        )
      } else {
        throw new Error('ลบเมนูไม่สำเร็จ')
      }
    } catch (error) {
      console.error('Error deleting menu:', error)
      alert('เกิดข้อผิดพลาดในการลบเมนู')
    } finally {
      setIsDeleteModalOpen(false) // ปิด modal หลังจากลบ
      setMenuToDelete(null) // รีเซ็ตค่าเมนูที่ต้องการลบ
    }
  }

  const ConfirmDeleteModal = ({ isOpen, onClose, onConfirm, menuName }) => {
    if (!isOpen) return null // ซ่อน Modal หาก isOpen เป็น false

    return (
      <div className="fixed inset-0 bg-gray-800 bg-opacity-50 flex justify-center items-center">
        <div className="bg-white rounded-lg p-6 w-96 shadow-lg">
          <h2 className="text-lg font-semibold mb-4">ยืนยันการลบ</h2>
          <p className="text-black">
            คุณต้องการลบเมนู{' '}
            <strong className="text-red-500">{menuName}</strong> ใช่หรือไม่?
          </p>
          <div className="mt-6 flex justify-end">
            <button
              className="px-4 py-2 bg-gray-300 rounded-lg mr-4"
              onClick={onClose}
            >
              ยกเลิก
            </button>
            <button
              className="px-4 py-2 bg-red-500 text-white rounded-lg"
              onClick={onConfirm}
            >
              ลบ
            </button>
          </div>
        </div>
      </div>
    )
  }

  // ลบ Option
  const deleteOption = async (optionId) => {
    try {
      const token = localStorage.getItem('token') // ดึง JWT
  
      const response = await fetch(`${API_BASE_URL}/options/${optionId}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`, // เพิ่ม JWT ใน Header
        },
      })
  
      if (response.ok) {
        alert('ลบตัวเลือกสำเร็จ')
      } else {
        alert('เกิดข้อผิดพลาดในการลบตัวเลือก')
      }
    } catch (error) {
      console.error('Error deleting option:', error)
      alert('ไม่สามารถลบตัวเลือกได้')
    }
  }
  

  const handleDeleteOptionGroup = async (groupIndex) => {
    try {
      const group = menuDetails.optionGroups[groupIndex];
      if (group.ID) {
        // ถ้ามี ID แสดงว่าเป็น group ที่มีอยู่แล้ว ต้องลบผ่าน API
        const token = localStorage.getItem('token');
        const response = await axios.delete(
          `${API_BASE_URL}/option-groups/${group.ID}`,
          {
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
          }
        )
        
        if (response.status === 200) {
          // ลบสำเร็จ อัพเดท state
          const updatedGroups = menuDetails.optionGroups.filter((_, index) => index !== groupIndex);
          setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
          alert('ลบกลุ่มตัวเลือกสำเร็จ');
        } else {
          throw new Error('Failed to delete option group');
        }
      } else {
        // ถ้าไม่มี ID แสดงว่าเป็น group ใหม่ที่ยังไม่ได้บันทึก ลบจาก state ได้เลย
        const updatedGroups = menuDetails.optionGroups.filter((_, index) => index !== groupIndex);
        setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
      }
    } catch (error) {
      console.error('Error deleting option group:', error);
      alert('เกิดข้อผิดพลาดในการลบกลุ่มตัวเลือก: ' + (error.response?.data?.error || error.message));
    }
  };

  const handleDeleteOption = async (groupIndex, optionIndex) => {
    try {
      const option = menuDetails.optionGroups[groupIndex].options[optionIndex];
      if (option.ID) {
        // ถ้ามี ID แสดงว่าเป็น option ที่มีอยู่แล้ว ต้องลบผ่าน API
        const token = localStorage.getItem('token');
        const response = await axios.delete(
          `${API_BASE_URL}/options/${option.ID}`,
          {
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${token}`,
            },
          }
        )
        
        if (response.status === 200) {
          // ลบสำเร็จ อัพเดท state
          const updatedGroups = [...menuDetails.optionGroups];
          updatedGroups[groupIndex].options = updatedGroups[groupIndex].options.filter(
            (_, index) => index !== optionIndex
          );
          setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
          alert('ลบตัวเลือกสำเร็จ');
        }
      } else {
        // ถ้าไม่มี ID แสดงว่าเป็น option ใหม่ที่ยังไม่ได้บันทึก ลบจาก state ได้เลย
        const updatedGroups = [...menuDetails.optionGroups];
        updatedGroups[groupIndex].options = updatedGroups[groupIndex].options.filter(
          (_, index) => index !== optionIndex
        );
        setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
      }
    } catch (error) {
      console.error('Error deleting option:', error);
      alert('เกิดข้อผิดพลาดในการลบตัวเลือก');
    }
  };

  // เพิ่ม Group opptions ใหม่ใน modal แก้ไข menu
  const handleAddOptionGroup = () => {
    const newGroup = {
      ID: null, // สำหรับ Group ใหม่ยังไม่มี ID
      name: '',
      nameEn: '',
      nameCh: '',
      MaxSelections: 1, // ค่าเริ่มต้น
      isRequired: false, // ค่าเริ่มต้น
      options: [], // เริ่มต้นไม่มี Option
    }
    setMenuDetails((prevDetails) => ({
      ...prevDetails,
      optionGroups: [...prevDetails.optionGroups, newGroup],
    }))
  }

  // เพิ่ม opptions ใหม่ใน modal แก้ไข menu
  const handleAddOption = (groupIndex) => {
    const updatedGroups = [...menuDetails.optionGroups]
    updatedGroups[groupIndex].options.push({
      name: '',
      nameEn: '',
      nameCh: '',
      price: '',
    })
    console.log('Updated Option Groups:', updatedGroups) // ดูค่าที่อัปเดต
    setMenuDetails({ ...menuDetails, optionGroups: updatedGroups })
  }

  // เพิ่มฟังก์ชันสำหรับการสลับการแสดง/ซ่อนตัวเลือก
  const toggleOptions = (menuId) => {
    setExpandedOptions(prev => ({
      ...prev,
      [menuId]: !prev[menuId]
    }))
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 p-6">
      {/* Header Section */}
      <div className="mb-8">
        <div className="bg-white rounded-xl shadow-lg p-6">
          <h1 className="text-2xl font-bold text-gray-800 mb-2">จัดการเมนูอาหาร</h1>
          <p className="text-gray-600">จัดการรายการอาหาร เพิ่ม แก้ไข และปรับแต่งตัวเลือกต่างๆ</p>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="mb-6">
        <div className="bg-white rounded-xl shadow-lg p-2 inline-flex space-x-2">
          <button
            className={`px-6 py-3 rounded-lg transition-all duration-200 ${
              activeTab === 'menu'
                ? 'bg-blue-500 text-white shadow-md'
                : 'text-gray-600 hover:bg-gray-100'
            }`}
            onClick={() => setActiveTab('menu')}
          >
            ข้อมูลสินค้า
          </button>
          <button
            className={`px-6 py-3 rounded-lg transition-all duration-200 ${
              activeTab === 'promotion'
                ? 'bg-blue-500 text-white shadow-md'
                : 'text-gray-600 hover:bg-gray-100'
            }`}
            onClick={() => setActiveTab('promotion')}
          >
            ข้อมูลโปรโมชัน
          </button>
          <button
            className={`px-6 py-3 rounded-lg transition-all duration-200 ${
              activeTab === 'restore'
                ? 'bg-blue-500 text-white shadow-md'
                : 'text-gray-600 hover:bg-gray-100'
            }`}
            onClick={() => setActiveTab('restore')}
          >
            ข้อมูลสินค้าที่ถูกลบ
          </button>
        </div>
      </div>

      {/* Main Content Area */}
      {activeTab === 'menu' && (
        <div className="bg-white rounded-xl shadow-lg p-6">
          {/* Search and Filter Section */}
          <div className="flex flex-wrap gap-4 mb-6">
            <div className="flex-1 min-w-[200px]">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={20} />
                <input
                  type="text"
                  className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                  placeholder="ค้นหาชื่อเมนู ราคา..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
            </div>
            <div className="flex-1 min-w-[200px]">
              <select
                className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all appearance-none cursor-pointer bg-white"
                value={categoryFilter}
                onChange={(e) => setCategoryFilter(e.target.value)}
              >
                <option value="">ค้นหาตามหมวดหมู่</option>
                {categories.map((category) => (
                  <option key={category.ID} value={category.ID}>
                    {category.Name}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex-1 min-w-[200px]">
              <select
                className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all appearance-none cursor-pointer bg-white"
                value={optionGroupFilter}
                onChange={(e) => setOptionGroupFilter(e.target.value)}
              >
                <option value="">ตัวเลือกทั้งหมด</option>
                {menus.flatMap((menu) =>
                  menu.OptionGroups.map((group) => (
                    <option key={group.ID} value={group.ID}>
                      {group.Name}
                    </option>
                  ))
                )}
              </select>
            </div>
            <button
              className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-all duration-200 shadow-md hover:shadow-lg flex items-center gap-2"
              onClick={() => setShowAddMenuModal(true)}
            >
              <Plus size={20} />
              เพิ่มเมนูอาหาร
            </button>
          </div>

          {/* Menu Table */}
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="bg-gradient-to-r from-blue-500 to-blue-600 text-white">
                  <th className="px-4 py-3 text-center rounded-tl-lg">สถานะ</th>
                  <th className="px-4 py-3">รหัสสินค้า</th>
                  <th className="px-4 py-3">รูปสินค้า</th>
                  <th className="px-4 py-3">ชื่อเมนู</th>
                  <th className="px-4 py-3">คำอธิบาย</th>
                  <th className="px-4 py-3">หมวดหมู่</th>
                  <th className="px-4 py-3">ราคา (THB)</th>
                  <th className="px-4 py-3">ตัวเลือก</th>
                  <th className="px-4 py-3 rounded-tr-lg">จัดการ</th>
                </tr>
              </thead>
              <tbody>
                {Array.isArray(filteredMenus) && filteredMenus.length > 0 ? (
                  filteredMenus.map((menu, index) => (
                    <React.Fragment key={menu.ID}>
                      <tr className={`border-b border-gray-200 ${index % 2 === 0 ? 'bg-white' : 'bg-gray-50'} hover:bg-blue-50 transition-colors`}>
                        <td className="px-4 py-3 text-center">
                          <label className="relative inline-flex items-center cursor-pointer">
                            <input
                              type="checkbox"
                              className="sr-only peer"
                              checked={menu.Is_available}
                              onChange={() => toggleMenuStatus(menu.ID)}
                            />
                            <div className="w-11 h-6 bg-gray-200 rounded-full peer peer-focus:ring-4 peer-focus:ring-blue-300 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-500"></div>
                          </label>
                        </td>
                        <td className="px-4 py-3 text-center">
                          <span className="font-medium text-gray-700 bg-gray-100 px-3 py-1 rounded-full">
                            {menu.ID}
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          {menu.Image ? (
                            <div className="group relative">
                              <img
                                src={`data:image/png;base64,${menu.Image}`}
                                alt={menu.Name}
                                className="w-32 h-32 object-cover rounded-xl shadow-md transition-transform transform hover:scale-105"
                              />
                              <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-20 rounded-xl transition-all duration-200">
                                <button
                                  onClick={() => {
                                    setCurrentMenuId(menu.ID)
                                    setShowUploadModal(true)
                                  }}
                                  className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100"
                                >
                                  <div className="bg-white p-2 rounded-lg shadow-lg">
                                    <Image className="w-5 h-5 text-gray-600" />
                                  </div>
                                </button>
                              </div>
                            </div>
                          ) : (
                            <div className="w-32 h-32 bg-gray-100 rounded-xl flex flex-col items-center justify-center text-gray-400 hover:bg-gray-200 transition-colors cursor-pointer"
                                 onClick={() => {
                                   setCurrentMenuId(menu.ID)
                                   setShowUploadModal(true)
                                 }}>
                              <Image className="w-8 h-8 mb-2" />
                              <span className="text-sm">เพิ่มรูปภาพ</span>
                            </div>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <div className="space-y-1">
                            <p className="font-medium text-gray-800 text-lg">{menu.Name}</p>
                            <p className="text-sm text-gray-600">{menu.NameEn}</p>
                            <p className="text-sm text-gray-600">{menu.NameCh}</p>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <div className="space-y-2">
                            <p className="text-sm text-gray-800">{menu.Description || '-'}</p>
                            <p className="text-sm text-gray-600">{menu.DescriptionEn || '-'}</p>
                            <p className="text-sm text-gray-600">{menu.DescriptionCh || '-'}</p>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <span className="px-4 py-2 bg-gradient-to-r from-blue-50 to-blue-100 text-blue-800 rounded-lg text-sm font-medium shadow-sm">
                            {getCategoryNameById(menu.CategoryID)}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-center">
                          <div className="flex items-center space-x-2">
                            <span className="font-medium text-lg text-gray-800">
                              {menu.Price.toLocaleString()} ฿
                            </span>
                            {menu.IsRecommended && (
                              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                                แนะนำ
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          {Array.isArray(menu.OptionGroups) && menu.OptionGroups.length > 0 ? (
                            <button
                              onClick={() => toggleOptions(menu.ID)}
                              className="px-4 py-2 bg-white border border-gray-200 rounded-lg shadow-sm hover:bg-gray-50 transition-colors flex items-center justify-between w-full"
                            >
                              <span className="text-sm font-medium text-gray-700">
                                {menu.OptionGroups.length} กลุ่มตัวเลือก
                              </span>
                              <svg
                                className={`w-5 h-5 text-gray-500 transform transition-transform ${
                                  expandedOptions[menu.ID] ? 'rotate-180' : ''
                                }`}
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 9l-7 7-7-7" />
                              </svg>
                            </button>
                          ) : (
                            <div className="text-gray-400 italic text-sm">ไม่มีตัวเลือก</div>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <div className="relative">
                            <button
                              className="p-2 hover:bg-gray-100 rounded-full transition-colors"
                              onClick={() => toggleDropdown(menu.ID)}
                            >
                              <Ellipsis />
                            </button>
                            {isDropdownOpen === menu.ID && (
                              <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-xl border border-gray-100 z-10 overflow-hidden">
                                <button
                                  onClick={() => {
                                    setCurrentMenuId(menu.ID)
                                    setMenuDetails({
                                      ID: menu.ID,
                                      name: menu.Name,
                                      nameEn: menu.NameEn,
                                      nameCh: menu.NameCh,
                                      price: menu.Price,
                                      description: menu.Description,
                                      descriptionEn: menu.DescriptionEn,
                                      descriptionCh: menu.DescriptionCh,
                                      categoryId: menu.CategoryID,
                                      optionGroups: menu.OptionGroups.map((group) => ({
                                        ID: group.ID,
                                        name: group.Name,
                                        nameEn: group.NameEn,
                                        nameCh: group.NameCh,
                                        MaxSelections: group.MaxSelections,
                                        isRequired: group.IsRequired,
                                        options: group.Options.map((option) => ({
                                          ID: option.ID,
                                          name: option.Name,
                                          nameEn: option.NameEn,
                                          nameCh: option.NameCh,
                                          price: option.Price,
                                        })),
                                      })),
                                    })
                                    setIsDropdownOpen(false)
                                    setShowEditMenuModal(true)
                                  }}
                                  className="w-full px-4 py-3 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2 border-b border-gray-100"
                                >
                                  <Edit className="w-4 h-4" />
                                  แก้ไขเมนู
                                </button>
                                <button
                                  onClick={() => {
                                    setIsDropdownOpen(false)
                                    setCurrentMenuId(menu.ID)
                                    setShowUploadModal(true)
                                  }}
                                  className="w-full px-4 py-3 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2 border-b border-gray-100"
                                >
                                  <Image className="w-4 h-4" />
                                  อัปโหลดรูปภาพ
                                </button>
                                <button
                                  onClick={() => {
                                    setIsDropdownOpen(false)
                                    setMenuToDelete(menu)
                                    setIsDeleteModalOpen(true)
                                  }}
                                  className="w-full px-4 py-3 text-left text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
                                >
                                  <Trash2 className="w-4 h-4" />
                                  ลบเมนู
                                </button>
                                <button
                                  onClick={async () => {
                                    try {
                                      const token = localStorage.getItem('token')
                                      await axios.put(
                                        `${API_BASE_URL}/${menu.ID}/recommend`,
                                        {},
                                        {
                                          headers: {
                                            Authorization: `Bearer ${token}`,
                                          },
                                        }
                                      )
                                      fetchMenus('getAll') // รีเฟรชข้อมูล
                                      setIsDropdownOpen(false)
                                    } catch (error) {
                                      console.error('Error toggling recommendation:', error)
                                      alert('เกิดข้อผิดพลาดในการเปลี่ยนสถานะเมนูแนะนำ')
                                    }
                                  }}
                                  className="w-full px-4 py-3 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2 border-b border-gray-100"
                                >
                                  <svg
                                    className="w-4 h-4"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                  >
                                    <path
                                      strokeLinecap="round"
                                      strokeLinejoin="round"
                                      strokeWidth="2"
                                      d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
                                  />
                                </svg>
                                {menu.IsRecommended ? 'ยกเลิกเมนูแนะนำ' : 'ตั้งเป็นเมนูแนะนำ'}
                              </button>
                              </div>
                            )}
                          </div>
                        </td>
                      </tr>
                      {/* Option Groups Row */}
                      {expandedOptions[menu.ID] && menu.OptionGroups.length > 0 && (
                        <tr className="border-b border-gray-200">
                          <td colSpan="9" className="px-6 py-4 bg-gradient-to-r from-blue-50/80 to-white">
                            <div className="space-y-4">
                              {menu.OptionGroups.map((group) => (
                                <div key={group.ID} className="bg-white rounded-lg shadow-md border-2 border-gray-100 overflow-hidden hover:border-blue-200 transition-all duration-200">
                                  <div className="px-4 py-3 bg-gradient-to-r from-blue-100/50 to-blue-50/50 border-b-2 border-gray-100">
                                    <div className="flex items-center justify-between">
                                      <div>
                                        <h4 className="font-medium text-gray-800">{group.Name}</h4>
                                        <div className="flex items-center space-x-2 mt-1">
                                          <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                                            group.IsRequired ? 'bg-red-100 text-red-700 border border-red-200' : 'bg-green-100 text-green-700 border border-green-200'
                                          }`}>
                                            {group.IsRequired ? 'บังคับเลือก' : 'เลือกได้'}
                                          </span>
                                          <span className="bg-blue-100 text-blue-700 px-2 py-1 rounded-full text-xs font-medium border border-blue-200">
                                            เลือกได้ {group.MaxSelections} รายการ
                                          </span>
                                        </div>
                                      </div>
                                      {group.NameEn && group.NameCh && (
                                        <div className="text-sm text-gray-500">
                                          <div>{group.NameEn}</div>
                                          <div>{group.NameCh}</div>
                                        </div>
                                      )}
                                    </div>
                                  </div>
                                  <div className="p-4 bg-gradient-to-b from-white to-gray-50">
                                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                      {group.Options.map((option) => (
                                        <div key={option.ID} className="bg-white rounded-lg p-3 shadow-sm border border-gray-200 hover:border-blue-300 hover:shadow-md transition-all duration-200">
                                          <div className="flex justify-between items-start">
                                            <div>
                                              <div className="font-medium text-gray-800">{option.Name}</div>
                                              {(option.NameEn || option.NameCh) && (
                                                <div className="text-sm text-gray-500 mt-1">
                                                  {option.NameEn && <div>{option.NameEn}</div>}
                                                  {option.NameCh && <div>{option.NameCh}</div>}
                                                </div>
                                              )}
                                            </div>
                                            <div className="text-green-600 font-medium bg-green-50 px-2 py-1 rounded-lg border border-green-100">
                                              +{option.Price.toLocaleString()} ฿
                                            </div>
                                          </div>
                                        </div>
                                      ))}
                                    </div>
                                  </div>
                                </div>
                              ))}
                            </div>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  ))
                ) : (
                  <tr>
                    <td colSpan="9" className="px-4 py-8 text-center text-gray-500 bg-gray-50 border-b border-gray-200">
                      ไม่พบเมนู
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === 'promotion' && <Promotions />}
      {activeTab === 'restore' && <MenuRestore />}

      {/* Modals */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full mx-4 transform transition-all">
            <div className="p-6">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-xl font-bold text-gray-800">อัพโหลดรูปภาพ</h2>
                <button
                  onClick={() => setShowUploadModal(false)}
                  className="text-gray-400 hover:text-gray-600 transition-colors"
                >
                  <X size={24} />
                </button>
              </div>
              
              <div className="space-y-4">
                <div className="border-2 border-dashed border-gray-300 rounded-lg p-4 text-center">
                  <input
                    type="file"
                    id="imageInput"
                    onChange={handleFileChange}
                    className="hidden"
                  />
                  <label
                    htmlFor="imageInput"
                    className="cursor-pointer block"
                  >
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
                  onClick={() => setShowUploadModal(false)}
                >
                  ยกเลิก
                </button>
                <button
                  className="px-4 py-2 text-white bg-blue-500 rounded-lg hover:bg-blue-600 transition-colors"
                  onClick={handleUploadImage}
                >
                  อัพโหลด
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      <ConfirmDeleteModal
        isOpen={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        onConfirm={deleteMenu}
        menuName={menuToDelete?.Name}
      />

      {/* Add Menu Modal */}
      {showAddMenuModal && (
        <AddMenuModal
          onClose={() => setShowAddMenuModal(false)}
          onMenuAdded={() => fetchMenus('getAll')}
        />
      )}

      {/* Edit Menu Modal */}
      {showEditMenuModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-4xl mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex justify-between items-center mb-6 border-b pb-4">
                <div>
                  <h2 className="text-xl font-bold text-gray-800">แก้ไขข้อมูลเมนู</h2>
                  <p className="text-sm text-gray-600">รหัสสินค้า: {menuDetails.ID}</p>
                </div>
                <button
                  onClick={() => setShowEditMenuModal(false)}
                  className="text-gray-400 hover:text-gray-600 transition-colors"
                >
                  <X size={24} />
                </button>
              </div>

              <div className="space-y-6">
                {/* ข้อมูลพื้นฐาน */}
                <div className="bg-gray-50 p-4 rounded-lg space-y-4">
                  <h3 className="text-lg font-semibold text-gray-700">ข้อมูลพื้นฐาน</h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ชื่อเมนู (ไทย)
                      </label>
                      <input
                        type="text"
                        value={menuDetails.name}
                        onChange={(e) => setMenuDetails({ ...menuDetails, name: e.target.value })}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ชื่อเมนู (อังกฤษ)
                      </label>
                      <input
                        type="text"
                        value={menuDetails.nameEn}
                        onChange={(e) => setMenuDetails({ ...menuDetails, nameEn: e.target.value })}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ชื่อเมนู (จีน)
                      </label>
                      <input
                        type="text"
                        value={menuDetails.nameCh}
                        onChange={(e) => setMenuDetails({ ...menuDetails, nameCh: e.target.value })}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ราคา (บาท)
                      </label>
                      <input
                        type="number"
                        value={menuDetails.price}
                        onChange={(e) => setMenuDetails({ ...menuDetails, price: e.target.value })}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        หมวดหมู่
                      </label>
                      <select
                        value={menuDetails.categoryId || ''}
                        onChange={(e) => setMenuDetails({
                          ...menuDetails,
                          categoryId: e.target.value ? Number(e.target.value) : null,
                        })}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      >
                        <option value="">กรุณาเลือกหมวดหมู่</option>
                        {categories.map((cat) => (
                          <option key={cat.ID} value={cat.ID}>
                            {cat.Name}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                </div>

                {/* คำอธิบาย */}
                <div className="bg-gray-50 p-4 rounded-lg space-y-4">
                  <h3 className="text-lg font-semibold text-gray-700">คำอธิบายเมนู</h3>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (ไทย)
                      </label>
                      <textarea
                        value={menuDetails.description}
                        onChange={(e) => setMenuDetails({
                          ...menuDetails,
                          description: e.target.value,
                        })}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="2"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (อังกฤษ)
                      </label>
                      <textarea
                        value={menuDetails.descriptionEn}
                        onChange={(e) => setMenuDetails({
                          ...menuDetails,
                          descriptionEn: e.target.value,
                        })}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="2"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        คำอธิบาย (จีน)
                      </label>
                      <textarea
                        value={menuDetails.descriptionCh}
                        onChange={(e) => setMenuDetails({
                          ...menuDetails,
                          descriptionCh: e.target.value,
                        })}
                        className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        rows="2"
                      />
                    </div>
                  </div>
                </div>

                {/* ตัวเลือก */}
                <div className="bg-gray-50 p-4 rounded-lg">
                  <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold text-gray-700">ตัวเลือกเพิ่มเติม</h3>
                    <button
                      type="button"
                      onClick={handleAddOptionGroup}
                      className="flex items-center px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
                    >
                      <Plus className="w-4 h-4 mr-2" />
                      เพิ่มกลุ่มตัวเลือก
                    </button>
                  </div>

                  <div className="space-y-4">
                    {menuDetails.optionGroups.map((group, groupIndex) => (
                      <div key={groupIndex} className="border border-gray-200 rounded-lg p-4 bg-white">
                        <div className="flex justify-between items-center mb-4">
                          <div className="flex items-center space-x-2">
                            <h4 className="text-md font-medium text-gray-700">
                              กลุ่มตัวเลือกที่ {groupIndex + 1}
                            </h4>
                            {group.ID && (
                              <span className="text-sm text-gray-500">
                                (รหัส: {group.ID})
                              </span>
                            )}
                          </div>
                          <button
                            type="button"
                            onClick={() => handleDeleteOptionGroup(groupIndex)}
                            className="text-red-500 hover:text-red-700 transition-colors"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                              จำนวนที่เลือกได้สูงสุด
                            </label>
                            <input
                              type="number"
                              value={group.MaxSelections}
                              onChange={(e) => {
                                const updatedGroups = [...menuDetails.optionGroups]
                                updatedGroups[groupIndex].MaxSelections = Number(e.target.value)
                                setMenuDetails({
                                  ...menuDetails,
                                  optionGroups: updatedGroups,
                                })
                              }}
                              min="1"
                              className="w-full p-2 border border-gray-300 rounded-lg"
                            />
                          </div>
                          <div className="flex items-center">
                            <label className="flex items-center cursor-pointer">
                              <input
                                type="checkbox"
                                checked={group.isRequired}
                                onChange={(e) => {
                                  const updatedGroups = [...menuDetails.optionGroups]
                                  updatedGroups[groupIndex].isRequired = e.target.checked
                                  setMenuDetails({
                                    ...menuDetails,
                                    optionGroups: updatedGroups,
                                  })
                                }}
                                className="form-checkbox h-5 w-5 text-blue-500"
                              />
                              <span className="ml-2 text-sm text-gray-700">บังคับให้เลือก</span>
                            </label>
                          </div>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                              ชื่อกลุ่ม (ไทย)
                            </label>
                            <input
                              type="text"
                              value={group.name}
                              onChange={(e) => {
                                const updatedGroups = [...menuDetails.optionGroups]
                                updatedGroups[groupIndex].name = e.target.value
                                setMenuDetails({
                                  ...menuDetails,
                                  optionGroups: updatedGroups,
                                })
                              }}
                              className="w-full p-2 border border-gray-300 rounded-lg"
                            />
                          </div>
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                              ชื่อกลุ่ม (อังกฤษ)
                            </label>
                            <input
                              type="text"
                              value={group.nameEn}
                              onChange={(e) => {
                                const updatedGroups = [...menuDetails.optionGroups]
                                updatedGroups[groupIndex].nameEn = e.target.value
                                setMenuDetails({
                                  ...menuDetails,
                                  optionGroups: updatedGroups,
                                })
                              }}
                              className="w-full p-2 border border-gray-300 rounded-lg"
                            />
                          </div>
                          <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                              ชื่อกลุ่ม (จีน)
                            </label>
                            <input
                              type="text"
                              value={group.nameCh}
                              onChange={(e) => {
                                const updatedGroups = [...menuDetails.optionGroups]
                                updatedGroups[groupIndex].nameCh = e.target.value
                                setMenuDetails({
                                  ...menuDetails,
                                  optionGroups: updatedGroups,
                                })
                              }}
                              className="w-full p-2 border border-gray-300 rounded-lg"
                            />
                          </div>
                        </div>

                        <div className="border-t pt-4">
                          <div className="flex justify-between items-center mb-4">
                            <h5 className="text-sm font-medium text-gray-700">ตัวเลือกในกลุ่ม</h5>
                            <button
                              type="button"
                              onClick={() => handleAddOption(groupIndex)}
                              className="flex items-center px-3 py-1 bg-blue-500 text-white text-sm rounded-lg hover:bg-blue-600 transition-colors"
                            >
                              <Plus className="w-4 h-4 mr-1" />
                              เพิ่มตัวเลือก
                            </button>
                          </div>

                          <div className="space-y-4">
                            {group.options.map((option, optionIndex) => (
                              <div key={optionIndex} className="flex items-center space-x-4 bg-gray-50 p-3 rounded-lg">
                                <div className="grid grid-cols-4 gap-4 flex-grow">
                                  <input
                                    type="text"
                                    placeholder="ชื่อ (ไทย)"
                                    value={option.name}
                                    onChange={(e) => {
                                      const updatedGroups = [...menuDetails.optionGroups]
                                      updatedGroups[groupIndex].options[optionIndex].name = e.target.value
                                      setMenuDetails({
                                        ...menuDetails,
                                        optionGroups: updatedGroups,
                                      })
                                    }}
                                    className="w-full p-2 border border-gray-300 rounded-lg"
                                  />
                                  <input
                                    type="text"
                                    placeholder="ชื่อ (อังกฤษ)"
                                    value={option.nameEn}
                                    onChange={(e) => {
                                      const updatedGroups = [...menuDetails.optionGroups]
                                      updatedGroups[groupIndex].options[optionIndex].nameEn = e.target.value
                                      setMenuDetails({
                                        ...menuDetails,
                                        optionGroups: updatedGroups,
                                      })
                                    }}
                                    className="w-full p-2 border border-gray-300 rounded-lg"
                                  />
                                  <input
                                    type="text"
                                    placeholder="ชื่อ (จีน)"
                                    value={option.nameCh}
                                    onChange={(e) => {
                                      const updatedGroups = [...menuDetails.optionGroups]
                                      updatedGroups[groupIndex].options[optionIndex].nameCh = e.target.value
                                      setMenuDetails({
                                        ...menuDetails,
                                        optionGroups: updatedGroups,
                                      })
                                    }}
                                    className="w-full p-2 border border-gray-300 rounded-lg"
                                  />
                                  <div className="flex items-center space-x-2">
                                    <input
                                      type="number"
                                      placeholder="ราคา"
                                      value={option.price}
                                      onChange={(e) => {
                                        const updatedGroups = [...menuDetails.optionGroups]
                                        updatedGroups[groupIndex].options[optionIndex].price = e.target.value
                                        setMenuDetails({
                                          ...menuDetails,
                                          optionGroups: updatedGroups,
                                        })
                                      }}
                                      className="w-full p-2 border border-gray-300 rounded-lg"
                                    />
                                    <button
                                      type="button"
                                      onClick={() => handleDeleteOption(groupIndex, optionIndex)}
                                      className="text-red-500 hover:text-red-700 transition-colors"
                                    >
                                      <Trash2 className="w-4 h-4" />
                                    </button>
                                  </div>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              <div className="mt-6 flex justify-end space-x-4 pt-4 border-t">
                <button
                  type="button"
                  onClick={() => setShowEditMenuModal(false)}
                  className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  ยกเลิก
                </button>
                <button
                  type="button"
                  onClick={handleEditMenu}
                  className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                >
                  บันทึก
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

{
  /* ---------------------------------------------------------------------------------------------------------- */
}
// Modal สำหรับเพิ่มเมนูอาหาร
const AddMenuModal = ({ onClose, onMenuAdded }) => {
  const [menuName, setMenuName] = useState('')
  const [menuNameEn, setMenuNameEn] = useState('')
  const [menuNameCh, setMenuNameCh] = useState('')
  const [price, setPrice] = useState('')
  const [description, setDescription] = useState('')
  const [descriptionEn, setDescriptionEn] = useState('')
  const [descriptionCh, setDescriptionCh] = useState('')
  const [category, setCategory] = useState('')
  const [categories, setCategories] = useState([])
  const [options, setOptions] = useState([]) // เก็บข้อมูลตัวเลือก
  const [image, setImage] = useState(null) // เก็บไฟล์รูปภาพที่เลือก
  const [isRecommended, setIsRecommended] = useState(false) // เพิ่ม state สำหรับ isRecommended

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        const token = localStorage.getItem('token')
        const response = await axios.get(API_BASE_URL_CATEGORIES, {
          method: 'DELETE',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`, // เพิ่ม JWT ใน Header
          },
        })
        if (response.status === 200) {
          setCategories(response.data) // เก็บข้อมูลหมวดหมู่
        }
      } catch (error) {
        console.error('Error fetching categories:', error)
      }
    }
    fetchCategories()
  }, [])

  // ฟังก์ชันเพิ่มกลุ่มตัวเลือก
  const handleAddOptionGroup = () => {
    setOptions([
      ...options,
      {
        MaxSelections: 1,
        is_required: false,
        name: '',
        name_en: '',
        name_ch: '',
        options: [],
      },
    ])
  }

  // ฟังก์ชันแก้ไขข้อมูลของกลุ่มตัวเลือก
  const handleOptionGroupChange = (groupIndex, field, value) => {
    const newOptions = [...options]
    newOptions[groupIndex][field] = value
    setOptions(newOptions)
  }

  // ฟังก์ชันเพิ่มตัวเลือกในกลุ่ม
  const handleAddOption = (groupIndex) => {
    const updatedGroups = [...options]
    updatedGroups[groupIndex].options.push({
      name: '',
      nameEn: '',
      nameCh: '',
      price: 0,
    })
    setOptions(updatedGroups)
  }

  // ฟังก์ชันแก้ไขข้อมูลของตัวเลือกในกลุ่ม
  const handleOptionChange = (groupIndex, optionIndex, field, value) => {
    const updatedOptions = [...options]
    updatedOptions[groupIndex].options[optionIndex][field] = value
    setOptions(updatedOptions)
  }

  // ฟังก์ชันลบตัวเลือกในกลุ่ม
  const handleRemoveOption = (groupIndex, optionIndex) => {
    const updatedOptions = [...options]
    updatedOptions[groupIndex].options.splice(optionIndex, 1)
    setOptions(updatedOptions)
  }

  // ฟังก์ชันลบกลุ่มตัวเลือก
  const handleRemoveOptionGroup = (groupIndex) => {
    const updatedOptions = [...options]
    updatedOptions.splice(groupIndex, 1)
    setOptions(updatedOptions)
  }

  const handleAddMenu = async (e) => {
    e.preventDefault()

    // ตรวจสอบข้อมูลที่จำเป็นก่อนส่ง
    if (!menuName || !category || !price) {
      alert('กรุณากรอกข้อมูลที่จำเป็น')
      return
    }

    // ตรวจสอบตัวเลือก
    const validOptionGroups = options.map((group) => ({
      MaxSelections: group.MaxSelections || 1,
      is_required: group.is_required ?? true, // ตรวจสอบว่ามีค่าหรือไม่ ถ้าไม่มีใช้ค่า default
      name: group.name || '',
      name_en: group.name_en || '',
      name_ch: group.name_ch || '',
      options: group.options.map((option) => ({
        name: option.name || '',
        name_en: option.name_en || '',
        name_ch: option.name_ch || '',
        price: parseFloat(option.price) || 0, // ราคาต้องเป็นตัวเลข
      })),
    }))

    const menuData = {
      menu_item: {
        name: menuName,
        name_en: menuNameEn,
        name_ch: menuNameCh,
        price: parseFloat(price) || 0,
        description: description || '',
        description_en: descriptionEn || '',
        description_ch: descriptionCh || '',
        category_id: parseInt(category, 10),
        image: [], // แก้ไขให้ image เป็น array ที่ API ยอมรับ
        is_recommended: isRecommended, // เพิ่มบรรทัดนี้
      },
      option_groups: validOptionGroups,
    }

    try {
      const token = localStorage.getItem('token')
      const response = await axios.post(
        API_BASE_URL,
        menuData,
        {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`, 
          },
        }
      )

      if (response.status === 200) {
        alert('เพิ่มเมนูสำเร็จ')
        onMenuAdded()
        onClose()
      }
    } catch (error) {
      console.error('Error adding menu:', error)
      const errorMessage =
        error.response?.data || 'เกิดข้อผิดพลาดในการเพิ่มเมนู'
      alert(errorMessage)
    }
  }

  return (
    <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center z-50 py-4">
      <div className="bg-white p-8 rounded-xl lg:w-6/12 sm:w-8/12 md:8/12 ml-12 h-screen overflow-y-auto shadow-2xl">
        <div className="flex justify-between items-center border-b pb-4">
          <h2 className="text-2xl font-bold text-gray-800">เพิ่มเมนูอาหาร</h2>
          <button
            onClick={onClose}
            className="hover:bg-gray-100 p-2 rounded-full transition-colors"
          >
            <X className="text-gray-500 hover:text-red-500" />
          </button>
        </div>
        <form onSubmit={handleAddMenu} className="mt-6 space-y-6">
          {/* ส่วนข้อมูลพื้นฐาน */}
          <div className="bg-gray-50 p-4 rounded-lg space-y-4">
            <h3 className="text-lg font-semibold text-gray-700 mb-4">
              ข้อมูลพื้นฐาน
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  ชื่อเมนู (ไทย) <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={menuName}
                  onChange={(e) => setMenuName(e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                  placeholder="กรุณากรอกชื่อเมนู"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  ชื่อเมนู (อังกฤษ)
                </label>
                <input
                  type="text"
                  value={menuNameEn}
                  onChange={(e) => setMenuNameEn(e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Enter menu name"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  ชื่อเมนู (จีน)
                </label>
                <input
                  type="text"
                  value={menuNameCh}
                  onChange={(e) => setMenuNameCh(e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="输入菜单名称"
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  ราคา (บาท) <span className="text-red-500">*</span>
                </label>
                <input
                  type="number"
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                  onWheel={(e) => e.preventDefault()}
                  className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                  min="0"
                  step="0.01"
                  placeholder="0.00"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  หมวดหมู่ <span className="text-red-500">*</span>
                </label>
                <select
                  value={category}
                  onChange={(e) => setCategory(e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                >
                  <option value="">เลือกหมวดหมู่</option>
                  {categories.map((cat) => (
                    <option key={cat.ID} value={cat.ID}>
                      {cat.Name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {/* ส่วนคำอธิบาย */}
          <div className="bg-gray-50 p-4 rounded-lg space-y-4">
            <h3 className="text-lg font-semibold text-gray-700 mb-4">
              คำอธิบายเมนู
            </h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  คำอธิบาย (ไทย)
                </label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  rows="2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  คำอธิบาย (อังกฤษ)
                </label>
                <textarea
                  value={descriptionEn}
                  onChange={(e) => setDescriptionEn(e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  rows="2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  คำอธิบาย (จีน)
                </label>
                <textarea
                  value={descriptionCh}
                  onChange={(e) => setDescriptionCh(e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  rows="2"
                />
              </div>
            </div>
          </div>

          {/* ส่วนตัวเลือก */}
          <div className="bg-gray-50 p-4 rounded-lg">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold text-gray-700">
                ตัวเลือกเพิ่มเติม
              </h3>
              <button
                type="button"
                onClick={handleAddOptionGroup}
                className="flex items-center px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
              >
                <Plus className="w-4 h-4 mr-2" />
                เพิ่มกลุ่มตัวเลือก
              </button>
            </div>

            <div className="space-y-4">
              {options.map((group, groupIndex) => (
                <div
                  key={groupIndex}
                  className="border border-gray-200 rounded-lg p-4 bg-white"
                >
                  <div className="flex justify-between items-center mb-4">
                    <h4 className="text-md font-medium text-gray-700">
                      กลุ่มตัวเลือกที่ {groupIndex + 1}
                    </h4>
                    <button
                      type="button"
                      onClick={() => handleRemoveOptionGroup(groupIndex)}
                      className="text-red-500 hover:text-red-700 transition-colors"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        จำนวนที่เลือกได้สูงสุด
                      </label>
                      <input
                        type="number"
                        value={group.MaxSelections}
                        onChange={(e) =>
                          handleOptionGroupChange(
                            groupIndex,
                            'MaxSelections',
                            e.target.value
                          )
                        }
                        min="1"
                        className="w-full p-2 border border-gray-300 rounded-lg"
                      />
                    </div>
                    <div className="flex items-center">
                      <label className="flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={group.is_required}
                          onChange={(e) =>
                            handleOptionGroupChange(
                              groupIndex,
                              'is_required',
                              e.target.checked
                            )
                          }
                          className="form-checkbox h-5 w-5 text-blue-500"
                        />
                        <span className="ml-2 text-sm text-gray-700">
                          บังคับให้เลือก
                        </span>
                      </label>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                    <input
                      type="text"
                      placeholder="ชื่อกลุ่ม (ไทย)"
                      value={group.name}
                      onChange={(e) =>
                        handleOptionGroupChange(
                          groupIndex,
                          'name',
                          e.target.value
                        )
                      }
                      className="w-full p-2 border border-gray-300 rounded-lg"
                    />
                    <input
                      type="text"
                      placeholder="ชื่อกลุ่ม (อังกฤษ)"
                      value={group.nameEn}
                      onChange={(e) =>
                        handleOptionGroupChange(
                          groupIndex,
                          'name_en',
                          e.target.value
                        )
                      }
                      className="w-full p-2 border border-gray-300 rounded-lg"
                    />
                    <input
                      type="text"
                      placeholder="ชื่อกลุ่ม (จีน)"
                      value={group.nameCh}
                      onChange={(e) =>
                        handleOptionGroupChange(
                          groupIndex,
                          'name_ch',
                          e.target.value
                        )
                      }
                      className="w-full p-2 border border-gray-300 rounded-lg"
                    />
                  </div>

                  <div className="border-t pt-4">
                    <div className="flex justify-between items-center mb-4">
                      <h5 className="text-sm font-medium text-gray-700">
                        ตัวเลือกในกลุ่ม
                      </h5>
                      <button
                        type="button"
                        onClick={() => handleAddOption(groupIndex)}
                        className="flex items-center px-3 py-1 bg-blue-500 text-white text-sm rounded-lg hover:bg-blue-600 transition-colors"
                      >
                        <Plus className="w-4 h-4 mr-1" />
                        เพิ่มตัวเลือก
                      </button>
                    </div>

                    <div className="space-y-4">
                      {group.options.map((option, optionIndex) => (
                        <div
                          key={optionIndex}
                          className="flex items-center space-x-4 bg-gray-50 p-3 rounded-lg"
                        >
                          <div className="grid grid-cols-4 gap-4 flex-grow">
                            <input
                              type="text"
                              placeholder="ชื่อ (ไทย)"
                              value={option.name}
                              onChange={(e) =>
                                handleOptionChange(
                                  groupIndex,
                                  optionIndex,
                                  'name',
                                  e.target.value
                                )
                              }
                              className="w-full p-2 border border-gray-300 rounded-lg"
                            />
                            <input
                              type="text"
                              placeholder="ชื่อ (อังกฤษ)"
                              value={option.nameEn}
                              onChange={(e) =>
                                handleOptionChange(
                                  groupIndex,
                                  optionIndex,
                                  'name_en',
                                  e.target.value
                                )
                              }
                              className="w-full p-2 border border-gray-300 rounded-lg"
                            />
                            {/* <input
                              type="text"
                              placeholder="ชื่อ (อังกฤษ)"
                              value={option.nameEn}
                              onChange={(e) => handleOptionChange(groupIndex, optionIndex, 'name_en', e.target.value)}
                              className="w-full p-2 border border-gray-300 rounded-lg"
                            /> */}
                            <input
                              type="text"
                              placeholder="ชื่อ (จีน)"
                              value={option.nameCh}
                              onChange={(e) =>
                                handleOptionChange(
                                  groupIndex,
                                  optionIndex,
                                  'name_ch',
                                  e.target.value
                                )
                              }
                              className="w-full p-2 border border-gray-300 rounded-lg"
                            />
                            <div className="flex items-center space-x-2">
                              <input
                                type="number"
                                placeholder="ราคา"
                                value={option.price}
                                onChange={(e) =>
                                  handleOptionChange(
                                    groupIndex,
                                    optionIndex,
                                    'price',
                                    e.target.value
                                  )
                                }
                                className="w-full p-2 border border-gray-300 rounded-lg"
                              />
                              <button
                                type="button"
                                onClick={() =>
                                  handleRemoveOption(groupIndex, optionIndex)
                                }
                                className="text-red-500 hover:text-red-700 transition-colors"
                              >
                                <Trash2 className="w-4 h-4" />
                              </button>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* ปุ่มดำเนินการ */}
          <div className="flex justify-end space-x-4 pt-4 border-t">
            <button
              type="button"
              onClick={onClose}
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
  )
}

export default MenuManagement