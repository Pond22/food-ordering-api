import React, { useState, useEffect } from 'react'
import { X, Plus, Trash2 } from 'lucide-react'
import axios from 'axios'

const API_BASE_URL = 'http://127.0.0.1:8080/api/printers'

const PrintBillCheckModal = ({
  isOpen,
  onClose,
  mainTable,
  occupiedTables,
  billableItems,
  tableBillableItems,
  selectedDiscounts,
  selectedCharges,
  serviceCharge,
  user,
  discounts,
  charges
}) => {
  const [isPrinting, setIsPrinting] = useState(false)
  const [printSelectedTables, setPrintSelectedTables] = useState([])
  const [error, setError] = useState(null)
  const [localDiscounts, setLocalDiscounts] = useState([{ discountID: '', reason: '' }])
  const [localCharges, setLocalCharges] = useState([{ chargeID: '', quantity: 1, note: '' }])

  useEffect(() => {
    if (!isOpen) {
      setPrintSelectedTables([])
      setError(null)
      setLocalDiscounts([{ discountID: '', reason: '' }])
      setLocalCharges([{ chargeID: '', quantity: 1, note: '' }])
    }
  }, [isOpen])

  const handleTableSelection = (tableId) => {
    setPrintSelectedTables(prev => {
      if (prev.includes(tableId)) {
        return prev.filter(id => id !== tableId)
      }
      return [...prev, tableId]
    })
  }

  const handleAddDiscount = () => {
    setLocalDiscounts([...localDiscounts, { discountID: '', reason: '' }])
  }

  const handleRemoveDiscount = (index) => {
    setLocalDiscounts(prev => prev.filter((_, i) => i !== index))
  }

  const handleDiscountChange = (index, field, value) => {
    setLocalDiscounts(prev => {
      const newDiscounts = [...prev]
      newDiscounts[index] = { ...newDiscounts[index], [field]: value }
      return newDiscounts
    })
  }

  const handleAddCharge = () => {
    setLocalCharges([...localCharges, { chargeID: '', quantity: 1, note: '' }])
  }

  const handleRemoveCharge = (index) => {
    setLocalCharges(prev => prev.filter((_, i) => i !== index))
  }

  const handleChargeChange = (index, field, value) => {
    setLocalCharges(prev => {
      const newCharges = [...prev]
      newCharges[index] = { ...newCharges[index], [field]: value }
      return newCharges
    })
  }

  const handlePrintBillCheck = async () => {
    setIsPrinting(true)
    setError(null)

    try {
      const token = localStorage.getItem('token')
      if (!token) {
        throw new Error('Session หมดอายุ กรุณาเข้าสู่ระบบใหม่')
      }

      // รวบรวม table IDs ที่จะพิมพ์
      const tableIDsToPrint = printSelectedTables.length > 0 
        ? printSelectedTables 
        : [mainTable.ID]

      // สร้าง request body ที่มีข้อมูลครบถ้วน
      const requestBody = {
        table_ids: tableIDsToPrint,
        service_charge: serviceCharge,
        discounts: localDiscounts
          .filter(discount => discount.discountID)
          .map(discount => ({
            discount_type_id: parseInt(discount.discountID),
            reason: discount.reason || ''
          })),
        extra_charges: localCharges
          .filter(charge => charge.chargeID)
          .map(charge => ({
            charge_type_id: parseInt(charge.chargeID),
            quantity: parseInt(charge.quantity) || 1,
            note: charge.note || ''
          }))
      }

      const response = await axios.post(
        `${API_BASE_URL}/bill-check`,
        requestBody,
        {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
        }
      )

      if (response.data) {
        alert('ส่งรายการไปยังเครื่องพิมพ์แล้ว')
        onClose()
      }
    } catch (error) {
      console.error('Error printing bill check:', error)
      setError(error.response?.data?.error || 'ไม่สามารถพิมพ์ใบรายการอาหารได้')
    } finally {
      setIsPrinting(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl p-8 max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="flex flex-col">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-2xl font-medium">เลือกโต๊ะและรายละเอียดเพื่อพิมพ์รายการอาหาร</h2>
            <button 
              onClick={onClose}
              className="text-gray-500 hover:text-gray-700"
            >
              <X className="w-6 h-6" />
            </button>
          </div>

          {/* ส่วนเลือกโต๊ะ */}
          <div className="mb-6">
            <h3 className="text-lg font-medium mb-3">เลือกโต๊ะ</h3>
            <div className="grid grid-cols-2 gap-4">
              {/* โต๊ะหลัก */}
              <div
                className={`p-4 rounded-xl border-2 cursor-pointer ${
                  printSelectedTables.includes(mainTable.ID)
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200'
                }`}
                onClick={() => handleTableSelection(mainTable.ID)}
              >
                <div className="font-medium">โต๊ะ {mainTable.Name} (หลัก)</div>
                <div className="text-gray-500 mt-1">
                  {billableItems.length} รายการ
                </div>
                <div className="text-gray-600 mt-1">
                  ยอดรวม: {calculateTableTotal(billableItems).toLocaleString()} ฿
                </div>
              </div>

              {/* โต๊ะอื่นๆ */}
              {occupiedTables?.map((table) => (
                table.ID !== mainTable.ID && (
                  <div
                    key={table.ID}
                    className={`p-4 rounded-xl border-2 cursor-pointer ${
                      printSelectedTables.includes(table.ID)
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200'
                    }`}
                    onClick={() => handleTableSelection(table.ID)}
                  >
                    <div className="font-medium">โต๊ะ {table.Name}</div>
                    <div className="text-gray-500 mt-1">
                      {Array.isArray(tableBillableItems[table.ID]) 
                        ? `${tableBillableItems[table.ID].length} รายการ`
                        : '0 รายการ'}
                    </div>
                    <div className="text-gray-600 mt-1">
                      ยอดรวม: {calculateTableTotal(tableBillableItems[table.ID] || []).toLocaleString()} ฿
                    </div>
                  </div>
                )
              ))}
            </div>
          </div>

          {/* ส่วนส่วนลด */}
          <div className="mb-6">
            <h3 className="text-lg font-medium mb-3">ส่วนลด</h3>
            {localDiscounts.map((discount, index) => (
              <div key={index} className="flex gap-4 mb-3">
                <select
                  className="flex-1 border rounded-lg p-2"
                  value={discount.discountID}
                  onChange={(e) => handleDiscountChange(index, 'discountID', e.target.value)}
                >
                  <option value="">เลือกส่วนลด</option>
                  {discounts?.map((d) => (
                    <option key={d.ID} value={d.ID}>
                      {d.Name} - {d.Type === 'percentage' ? `${d.Value}%` : `${d.Value}฿`}
                    </option>
                  ))}
                </select>
                <input
                  type="text"
                  placeholder="เหตุผล"
                  className="flex-1 border rounded-lg p-2"
                  value={discount.reason}
                  onChange={(e) => handleDiscountChange(index, 'reason', e.target.value)}
                />
                <button
                  onClick={() => handleRemoveDiscount(index)}
                  className="text-red-500 p-2"
                >
                  <Trash2 className="w-5 h-5" />
                </button>
              </div>
            ))}
            <button
              onClick={handleAddDiscount}
              className="flex items-center text-blue-500 mt-2"
            >
              <Plus className="w-4 h-4 mr-1" /> เพิ่มส่วนลด
            </button>
          </div>

          {/* ส่วนค่าใช้จ่ายเพิ่มเติม */}
          <div className="mb-6">
            <h3 className="text-lg font-medium mb-3">ค่าใช้จ่ายเพิ่มเติม</h3>
            {localCharges.map((charge, index) => (
              <div key={index} className="flex gap-4 mb-3">
                <select
                  className="flex-1 border rounded-lg p-2"
                  value={charge.chargeID}
                  onChange={(e) => handleChargeChange(index, 'chargeID', e.target.value)}
                >
                  <option value="">เลือกค่าใช้จ่าย</option>
                  {charges?.map((c) => (
                    <option key={c.ID} value={c.ID}>
                      {c.Name} - {c.DefaultAmount}฿
                    </option>
                  ))}
                </select>
                <input
                  type="number"
                  min="1"
                  placeholder="จำนวน"
                  className="w-24 border rounded-lg p-2"
                  value={charge.quantity}
                  onChange={(e) => handleChargeChange(index, 'quantity', e.target.value)}
                />
                <input
                  type="text"
                  placeholder="หมายเหตุ"
                  className="flex-1 border rounded-lg p-2"
                  value={charge.note}
                  onChange={(e) => handleChargeChange(index, 'note', e.target.value)}
                />
                <button
                  onClick={() => handleRemoveCharge(index)}
                  className="text-red-500 p-2"
                >
                  <Trash2 className="w-5 h-5" />
                </button>
              </div>
            ))}
            <button
              onClick={handleAddCharge}
              className="flex items-center text-blue-500 mt-2"
            >
              <Plus className="w-4 h-4 mr-1" /> เพิ่มค่าใช้จ่าย
            </button>
          </div>

          {error && (
            <div className="text-red-500 mb-4 text-center">
              {error}
            </div>
          )}

          <div className="flex justify-end gap-4">
            <button
              onClick={onClose}
              className="px-6 py-3 rounded-xl border border-gray-300"
            >
              ยกเลิก
            </button>
            <button
              onClick={handlePrintBillCheck}
              disabled={isPrinting}
              className={`px-6 py-3 rounded-xl ${
                isPrinting ? 'bg-gray-400' : 'bg-blue-500'
              } text-white`}
            >
              {isPrinting ? 'กำลังพิมพ์...' : 'พิมพ์รายการอาหาร'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

// ฟังก์ชันช่วยคำนวณยอดรวมของแต่ละโต๊ะ
const calculateTableTotal = (items) => {
  return items.reduce((total, item) => {
    let itemTotal = item.price * (item.quantity || 1)
    
    // เพิ่มราคาตัวเลือกเสริม
    if (item.options) {
      itemTotal += item.options.reduce((optTotal, opt) => 
        optTotal + (opt.price * opt.quantity), 0)
    }

    // ปรับราคาตามโปรโมชั่น (ถ้ามี)
    if (item.promotion) {
      itemTotal = item.promotion.price
    }

    return total + itemTotal
  }, 0)
}

export default PrintBillCheckModal 