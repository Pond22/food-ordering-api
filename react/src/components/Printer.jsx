import React, { useState, useEffect } from 'react'
import axios from 'axios'
import { Printer as PrinterIcon, Save, Plus, Trash2 } from 'lucide-react'

const Printer = () => {
  const [printers, setPrinters] = useState([])
  const [categories, setCategories] = useState([])
  const [categoryPrinters, setCategoryPrinters] = useState({})
  const [printerCategories, setPrinterCategories] = useState({})
  // const [showAddPrinter, setShowAddPrinter] = useState(false)
  // const [newPrinter, setNewPrinter] = useState({ name: '', ip: '' })
  const [isLoading, setIsLoading] = useState(false)

  // Fetch printers
  useEffect(() => {
    const fetchPrintersAndCategories = async () => {
      try {
        const token = localStorage.getItem('token')
        if (!token) return
  
        const [printersRes, categoriesRes] = await Promise.all([
          axios.get('http://localhost:8080/api/printers', {
            headers: { Authorization: `Bearer ${token}` },
          }),
          axios.get('http://localhost:8080/api/categories', {
            headers: { Authorization: `Bearer ${token}` },
          }),
        ])
  
        setPrinters(printersRes.data)
        setCategories(categoriesRes.data)
  
        // Fetch categories for all printers in one request
        const categoryRequests = printersRes.data.map((printer) =>
          axios.get(`http://localhost:8080/api/printers/categories/${printer.ID}`, {
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

  // Fetch categories
  useEffect(() => {
    const token = localStorage.getItem('token')
    axios
      .get('http://localhost:8080/api/categories', {
        headers: {
          Accept: 'application/json',
          Authorization: `Bearer ${token}`,
        },
      })
      .then((response) => {
        setCategories(response.data)
      })
      .catch((error) => {
        console.error('Error fetching categories data:', error)
      })
  }, [])

  const handleSaveSettings = async () => {
    setIsLoading(true)
    try {
      // Prepare category assignments with existing and new categories
      const categoryAssignments = {}
      const printersToCategoriesMap = {}

      // First, collect existing categories for each printer
      printers.forEach((printer) => {
        const existingCategories = getExistingCategories(printer.ID)
        if (existingCategories.length > 0) {
          printersToCategoriesMap[printer.ID] = existingCategories.map((cat) =>
            parseInt(cat.ID)
          )
        }
      })

      // Process newly selected categories
      categories.forEach((category) => {
        const newPrinterId = categoryPrinters[category.ID]
        if (newPrinterId) {
          // Remove this category from its current printer's categories
          Object.entries(printersToCategoriesMap).forEach(
            ([oldPrinterId, categoryIds]) => {
              printersToCategoriesMap[oldPrinterId] = categoryIds.filter(
                (catId) => catId !== parseInt(category.ID)
              )
            }
          )

          // Add category to new printer
          if (!printersToCategoriesMap[newPrinterId]) {
            printersToCategoriesMap[newPrinterId] = []
          }
          printersToCategoriesMap[newPrinterId].push(parseInt(category.ID))
        }
      })

      // Prepare API calls
      const apiCalls = Object.entries(printersToCategoriesMap).map(
        ([printerId, categoryIds]) =>
          axios.post(
            `http://localhost:8080/api/printers/categories/${printerId}`,
            { category_ids: [...new Set(categoryIds)] }
          )
      )

      // Execute API calls
      await Promise.all(apiCalls)

      // Update local state
      const updatedPrinterCategories = {}
      Object.entries(printersToCategoriesMap).forEach(
        ([printerId, categoryIds]) => {
          updatedPrinterCategories[printerId] = categoryIds.map((catId) =>
            categories.find((cat) => parseInt(cat.ID) === catId)
          )
        }
      )

      setPrinterCategories(updatedPrinterCategories)

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
    try {
      const token = localStorage.getItem('token')
      await axios.delete(`http://localhost:8080/api/printers/${printerId}`)
      setPrinters(printers.filter((printer) => printer.ID !== printerId))
    } catch (error) {
      console.error('Error deleting printer:', error)
      alert('เกิดข้อผิดพลาดในการลบเครื่องพิมพ์')
    }
  }

  // Get existing categories for a specific printer
  const getExistingCategories = (printerId) => {
    return printerCategories[printerId] || []
  }
  return (
    <div className="max-h-full h-full px-6 pb-8 mx-auto bg-gray-100">
      <div className="flex justify-between items-center p-4 mb-6 bg-gray-800 rounded-lg">
        <h1 className="text-2xl font-bold flex items-center text-white">
          <PrinterIcon className="w-8 h-8 mr-2" />
          ตั้งค่าเครื่องพิมพ์
        </h1>
      </div>

      <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">เครื่องพิมพ์ที่ตั้งค่าไว้</h2>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-2 text-left">ชื่อเครื่องพิมพ์</th>
                <th className="px-4 py-2 text-left">IP Address</th>
                <th className="px-4 py-2 text-left">สถานะ</th>
                <th className="px-4 py-2 text-right">จัดการ</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {printers.map((printer) => (
                <tr key={printer.ID}>
                  <td className="px-4 py-2">{printer.Name}</td>
                  <td className="px-4 py-2">{printer.IPAddress}</td>
                  <td className="px-4 py-2">
                    <span
                      className={`px-2 py-1 rounded-full text-xs ${
                        printer.Status === 'active'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {printer.Status === 'active'
                        ? 'พร้อมใช้งาน'
                        : 'ไม่พร้อมใช้งาน'}
                    </span>
                  </td>
                  <td className="px-4 py-2 text-right">
                    <button
                      onClick={() => handleDeletePrinter(printer.ID)}
                      className="text-red-600 hover:text-red-900"
                    >
                      <Trash2 className="w-5 h-5" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-lg p-6">
        <div className="flex justify-between">
          <h2 className="text-xl font-semibold mb-4">
            ตั้งค่าเครื่องพิมพ์ตามหมวดหมู่
          </h2>
          <button
            onClick={handleSaveSettings}
            disabled={isLoading}
            className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 mb-4"
          >
            <Save className="w-5 h-5 mr-2" />
            {isLoading ? 'กำลังบันทึก...' : 'บันทึกการตั้งค่า'}
          </button>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 ">
          {categories.map((category) => (
            <div key={category.ID} className="p-4 border rounded-lg shadow-md">
              <div className="flex items-center mb-2">
                <span className="text-2xl mr-2">{category.Name}</span>
              </div>
              <select
                className="w-full p-2 border rounded-md bg-white"
                value={categoryPrinters[category.ID] || ''}
                onChange={(e) =>
                  setCategoryPrinters({
                    ...categoryPrinters,
                    [category.ID]: e.target.value,
                  })
                }
              >
                <option value="">
                  {printers.filter((printer) => {
                    const existingCategories = getExistingCategories(printer.ID)
                    return existingCategories.some(
                      (existingCategory) => existingCategory.ID === category.ID
                    )
                  }).length > 0
                    ? printers
                        .filter((printer) => {
                          const existingCategories = getExistingCategories(
                            printer.ID
                          )
                          return existingCategories.some(
                            (existingCategory) =>
                              existingCategory.ID === category.ID
                          )
                        })
                        .map((printer) => `${printer.Name} (เลือกอยู่)`)
                        .join(', ')
                    : 'เลือกเครื่องพิมพ์'}
                </option>
                {printers.map((printer) => {
                  const existingCategories = getExistingCategories(printer.ID)
                  const isCategoryAssigned = existingCategories.some(
                    (existingCategory) => existingCategory.ID === category.ID
                  )

                  return (
                    <option
                      key={printer.ID}
                      value={printer.ID}
                      className={isCategoryAssigned ? 'bg-green-100' : ''}
                    >
                      {printer.Name}
                      {isCategoryAssigned ? ' (เลือกอยู่)' : ''}
                    </option>
                  )
                })}
              </select>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default Printer
