import React, { useState, useEffect } from 'react'
import axios from 'axios'
import { Printer as PrinterIcon, Save, Plus, Trash2 } from 'lucide-react'

const API_BASE_URL = `${import.meta.env.VITE_APP_API_URL}/api/printers`
const API_BASE_URL_CATEGORIES = `${import.meta.env.VITE_APP_API_URL}/api/categories`

const Printer = () => {
  const [printers, setPrinters] = useState([])
  const [categories, setCategories] = useState([])
  const [categoryPrinters, setCategoryPrinters] = useState({})
  const [printerCategories, setPrinterCategories] = useState({})
  const [isLoading, setIsLoading] = useState(false)
  const [selectedPrinter, setSelectedPrinter] = useState(null)

  // Fetch printers
  useEffect(() => {
    const fetchPrintersAndCategories = async () => {
      try {
        const token = localStorage.getItem('token')
        if (!token) return
  
        const [printersRes, categoriesRes] = await Promise.all([
          axios.get(`${API_BASE_URL}`, {
            headers: { Authorization: `Bearer ${token}` },
            accept: 'application/json',
          }),
          axios.get(`${API_BASE_URL_CATEGORIES}`, {
            headers: { Authorization: `Bearer ${token}` },
            Accept: 'application/json',
          }),
        ])
  
        setPrinters(printersRes.data)
        setCategories(categoriesRes.data)
  
        // Fetch categories for all printers in one request
        const categoryRequests = printersRes.data.map((printer) =>
          axios.get(`${API_BASE_URL}/categories/${printer.ID}`, {
            headers: { Authorization: `Bearer ${token}` },
          })
        )
  
        const categoryResponses = await Promise.all(categoryRequests)
  
        // อัปเดต printerCategories ให้เป็น object โดยใช้ printer.ID เป็น key
        const newPrinterCategories = {}
        printersRes.data.forEach((printer, index) => {
          newPrinterCategories[printer.ID] = categoryResponses[index].data
        })
  
        setPrinterCategories(newPrinterCategories)
      } catch (error) {
        console.error('Error fetching printers and categories:', error)
      }
    }
  
    fetchPrintersAndCategories()
  }, [])

  const handleSaveSettings = async () => {
    setIsLoading(true)
    try {
      const token = localStorage.getItem('token')
      if (!token) return

      // เตรียมข้อมูลสำหรับส่งไปยัง API
      const selectedCategories = Object.entries(categoryPrinters)
        .filter(([_, printerId]) => printerId === selectedPrinter.ID)
        .map(([categoryId]) => parseInt(categoryId))

      // ส่งข้อมูลไปยัง API
      const response = await axios.post(
        `${API_BASE_URL}/categories/${selectedPrinter.ID}`,
        { category_ids: selectedCategories },
        {
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      )

      // อัปเดตข้อมูลในหน้าจอ
      const updatedPrinterCategories = { ...printerCategories }
      updatedPrinterCategories[selectedPrinter.ID] = response.data.Categories
      setPrinterCategories(updatedPrinterCategories)

      // รีเซ็ต categoryPrinters สำหรับเครื่องพิมพ์ที่เลือก
      const newCategoryPrinters = {}
      response.data.Categories.forEach(category => {
        newCategoryPrinters[category.ID] = selectedPrinter.ID
      })
      setCategoryPrinters(newCategoryPrinters)

      alert('บันทึกการตั้งค่าหมวดหมู่สำเร็จ')
    } catch (error) {
      console.error('Error saving printer categories:', error)
      alert('เกิดข้อผิดพลาดในการบันทึกหมวดหมู่')
    } finally {
      setIsLoading(false)
    }
  }

  // Delete printer
  const handleDeletePrinter = async (printerId) => {
    if (!window.confirm('คุณแน่ใจหรือไม่ที่จะลบเครื่องพิมพ์นี้?')) return

    try {
      const token = localStorage.getItem('token')
      await axios.delete(`${API_BASE_URL}/${printerId}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      setPrinters(printers.filter((printer) => printer.ID !== printerId))
      alert('ลบเครื่องพิมพ์สำเร็จ')
    } catch (error) {
      console.error('Error deleting printer:', error)
      alert('เกิดข้อผิดพลาดในการลบเครื่องพิมพ์')
    }
  }

  // Get existing categories for a specific printer
  const getExistingCategories = (printerId) => {
    return printerCategories[printerId] || []
  }

  // เลือกเครื่องพิมพ์
  const handleSelectPrinter = (printer) => {
    setSelectedPrinter(printer)
    // อัปเดต categoryPrinters ตามหมวดหมู่ที่มีอยู่
    const existingCategories = getExistingCategories(printer.ID)
    const newCategoryPrinters = { ...categoryPrinters }
    existingCategories.forEach(category => {
      newCategoryPrinters[category.ID] = printer.ID
    })
    setCategoryPrinters(newCategoryPrinters)
  }

  return (
    <div className="max-h-full h-full px-6 pb-8 mx-auto bg-gray-100">
      <div className="flex justify-between items-center p-4 mb-6 bg-gray-800 rounded-lg">
        <h1 className="text-2xl font-bold flex items-center text-white">
          <PrinterIcon className="w-8 h-8 mr-2" />
          ตั้งค่าเครื่องพิมพ์
        </h1>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* รายการเครื่องพิมพ์ */}
        <div className="bg-white rounded-lg shadow-lg p-6">
          <h2 className="text-xl font-semibold mb-4">เครื่องพิมพ์ทั้งหมด</h2>
          <div className="space-y-4">
            {printers.map((printer) => (
              <div
                key={printer.ID}
                className={`p-4 border rounded-lg cursor-pointer transition-all ${
                  selectedPrinter?.ID === printer.ID
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200 hover:border-blue-300'
                }`}
                onClick={() => handleSelectPrinter(printer)}
              >
                <div className="flex justify-between items-center">
                  <div>
                    <h3 className="font-medium">{printer.Name}</h3>
                    <p className="text-sm text-gray-600">{printer.IPAddress}</p>
                  </div>
                  <div className="flex items-center space-x-2">
                    <span
                      className={`px-2 py-1 rounded-full text-xs ${
                        printer.Status === 'active'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {printer.Status === 'active' ? 'พร้อมใช้งาน' : 'ไม่พร้อมใช้งาน'}
                    </span>
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDeletePrinter(printer.ID)
                      }}
                      className="text-red-600 hover:text-red-900"
                    >
                      <Trash2 className="w-5 h-5" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* กำหนดหมวดหมู่ */}
        <div className="md:col-span-2 bg-white rounded-lg shadow-lg p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">
              {selectedPrinter
                ? `กำหนดหมวดหมู่สำหรับ ${selectedPrinter.Name}`
                : 'เลือกเครื่องพิมพ์เพื่อกำหนดหมวดหมู่'}
            </h2>
            {selectedPrinter && (
              <button
                onClick={handleSaveSettings}
                disabled={isLoading}
                className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
              >
                <Save className="w-5 h-5 mr-2" />
                {isLoading ? 'กำลังบันทึก...' : 'บันทึกการตั้งค่า'}
              </button>
            )}
          </div>

          {selectedPrinter ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {categories.map((category) => {
                const isAssigned = categoryPrinters[category.ID] === selectedPrinter.ID
                return (
                  <div
                    key={category.ID}
                    className={`p-4 border rounded-lg cursor-pointer transition-all ${
                      isAssigned
                        ? 'border-green-500 bg-green-50'
                        : 'border-gray-200 hover:border-blue-300'
                    }`}
                    onClick={() => {
                      setCategoryPrinters({
                        ...categoryPrinters,
                        [category.ID]: isAssigned ? null : selectedPrinter.ID,
                      })
                    }}
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-medium">{category.Name}</span>
                      {isAssigned && (
                        <div className="w-4 h-4 rounded-full bg-green-500"></div>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="text-center text-gray-500 py-8">
              กรุณาเลือกเครื่องพิมพ์จากรายการด้านซ้าย
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default Printer