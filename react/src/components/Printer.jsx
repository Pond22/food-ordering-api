import React, { useState } from 'react'
import { Printer as PrinterIcon, Save, Plus, Trash2 } from 'lucide-react' // เปลี่ยนชื่อ Printer เป็น PrinterIcon

const Printer = () => {
  const [printers, setPrinters] = useState([
    { id: 1, name: 'Kitchen Printer 1', ip: '192.168.1.101', status: 'online' },
    { id: 2, name: 'Sushi Bar Printer', ip: '192.168.1.102', status: 'online' },
    { id: 3, name: 'Bar Printer', ip: '192.168.1.103', status: 'offline' },
  ])

  const [categoryPrinters, setCategoryPrinters] = useState({
    sushi: 1,
    ramen: 1,
    tempura: 1,
    drinks: 3,
    desserts: 2,
  })

  const [showAddPrinter, setShowAddPrinter] = useState(false)
  const [newPrinter, setNewPrinter] = useState({ name: '', ip: '' })

  const foodCategories = [
    { id: 'sushi', name: 'ซูชิ/ซาซิมิ', icon: '🍣' },
    { id: 'ramen', name: 'ราเมน/อุด้ง', icon: '🍜' },
    { id: 'tempura', name: 'เทมปุระ/ทอด', icon: '🍤' },
    { id: 'drinks', name: 'เครื่องดื่ม', icon: '🍺' },
    { id: 'desserts', name: 'ของหวาน', icon: '🍡' },
  ]

  const handleSaveSettings = () => {
    // Logic to save printer settings
    alert('บันทึกการตั้งค่าเรียบร้อย')
  }

  const handleAddPrinter = (e) => {
    e.preventDefault()
    const newId = printers.length + 1
    setPrinters([...printers, { ...newPrinter, id: newId, status: 'online' }])
    setShowAddPrinter(false)
    setNewPrinter({ name: '', ip: '' })
  }

  return (
    <div className="max-h-full h-full  px-6 pb-8 mx-auto bg-gray-100">
      <div className="flex justify-between items-center p-4 mb-6 bg-gray-800 rounded-lg">
        <h1 className="text-2xl font-bold flex items-center text-white">
          <PrinterIcon className="w-8 h-8 mr-2 " /> {/* ใช้ PrinterIcon แทน */}
          ตั้งค่าเครื่องพิมพ์
        </h1>
        <button
          onClick={handleSaveSettings}
          className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Save className="w-5 h-5 mr-2" />
          บันทึกการตั้งค่า
        </button>
      </div>

      <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">เครื่องพิมพ์ที่ตั้งค่าไว้</h2>
          <button
            onClick={() => setShowAddPrinter(true)}
            className="flex items-center px-3 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
          >
            <Plus className="w-5 h-5 mr-1" />
            เพิ่มเครื่องพิมพ์
          </button>
        </div>

        {showAddPrinter && (
          <form
            onSubmit={handleAddPrinter}
            className="mb-4 p-4 border rounded-lg"
          >
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div>
                <label className="block text-sm font-medium mb-1">
                  ชื่อเครื่องพิมพ์
                </label>
                <input
                  type="text"
                  className="w-full p-2 border rounded-md"
                  value={newPrinter.name}
                  onChange={(e) =>
                    setNewPrinter({ ...newPrinter, name: e.target.value })
                  }
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">
                  IP Address
                </label>
                <input
                  type="text"
                  className="w-full p-2 border rounded-md"
                  value={newPrinter.ip}
                  onChange={(e) =>
                    setNewPrinter({ ...newPrinter, ip: e.target.value })
                  }
                  required
                  pattern="^(\d{1,3}\.){3}\d{1,3}$"
                />
              </div>
            </div>
            <div className="flex justify-end space-x-2">
              <button
                type="button"
                onClick={() => setShowAddPrinter(false)}
                className="px-4 py-2 text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200"
              >
                ยกเลิก
              </button>
              <button
                type="submit"
                className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
              >
                เพิ่มเครื่องพิมพ์
              </button>
            </div>
          </form>
        )}

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
                <tr key={printer.id}>
                  <td className="px-4 py-2">{printer.name}</td>
                  <td className="px-4 py-2">{printer.ip}</td>
                  <td className="px-4 py-2">
                    <span
                      className={`px-2 py-1 rounded-full text-xs ${
                        printer.status === 'online'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {printer.status === 'online'
                        ? 'พร้อมใช้งาน'
                        : 'ไม่พร้อมใช้งาน'}
                    </span>
                  </td>
                  <td className="px-4 py-2 text-right">
                    <button className="text-red-600 hover:text-red-900">
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
        <h2 className="text-xl font-semibold mb-4">
          ตั้งค่าเครื่องพิมพ์ตามหมวดหมู่
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {foodCategories.map((category) => (
            <div key={category.id} className="p-4 border rounded-lg">
              <div className="flex items-center mb-2">
                <span className="text-2xl mr-2">{category.icon}</span>
                <h3 className="font-medium">{category.name}</h3>
              </div>
              <select
                className="w-full p-2 border rounded-md bg-white"
                value={categoryPrinters[category.id]}
                onChange={(e) =>
                  setCategoryPrinters({
                    ...categoryPrinters,
                    [category.id]: parseInt(e.target.value),
                  })
                }
              >
                {printers.map((printer) => (
                  <option key={printer.id} value={printer.id}>
                    {printer.name}
                  </option>
                ))}
              </select>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default Printer
