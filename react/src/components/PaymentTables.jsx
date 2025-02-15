import React, { useState, useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import { ChevronLeft, CreditCard, Wallet, Check, Trash2, X } from 'lucide-react'
import axios from 'axios'
import OrderSummaryDetail from './OrderSummaryDetail'
import PrintBillCheckModal from './PrintBillCheckModal'

const PaymentTables = ({ user, posToken }) => {
  const [selectedPayment, setSelectedPayment] = useState('')
  const [vat, setVat] = useState(7)
  const [isProcessing, setIsProcessing] = useState(false)
  const [showSuccessModal, setShowSuccessModal] = useState(false)
  const [showPrintModal, setShowPrintModal] = useState(false)
  const [isPrinting, setIsPrinting] = useState(false)
  const [printSelectedTables, setPrintSelectedTables] = useState([])
  // const [promoCode, setPromoCode] = useState('')
  // const [promoError, setPromoError] = useState('')
  // const [appliedDiscount, setAppliedDiscount] = useState(0)
  const location = useLocation()
  const { tableID, uuid, tableName, occupiedTables } = location.state || {}
  const [selectedTables, setSelectedTables] = useState([])
  const [billableItems, setBillableItems] = useState([])
  const [charges, setCharges] = useState([])
  const [selectedCharges, setSelectedCharges] = useState([])
  const [discounts, setDiscounts] = useState([])
  const [selectedDiscounts, setSelectedDiscounts] = useState([
    { discountID: '', value: '' },
  ])
  const [tableBillableItems, setTableBillableItems] = useState({})
  const [isClosingTable, setIsClosingTable] = useState(false)
  const [showCloseConfirmModal, setShowCloseConfirmModal] = useState(false)
  // console.log('Occupied Tables:', occupiedTables)
  // console.log('User:', user)

  useEffect(() => {
    const fetchBillableItems = async () => {
      const token = localStorage.getItem('posToken')
      if (!token) {
        alert('Session หมดอายุ กรุณาเข้าสู่ระบบใหม่')
        window.location.href = '/login'
      }

      try {
        const response = await axios.get(
          `http://localhost:8080/api/table/billable/${uuid}`,
          {
            headers: {
              accept: 'application/json',
              Authorization: `Bearer ${token}`,
            },
          }
        )
        if (response.data && response.data.items) {
          setBillableItems(response.data.items)
        }
      } catch (error) {
        console.error('Error fetching billable items:', error)
      }
    }

    if (uuid) fetchBillableItems()
  }, [uuid])


  useEffect(() => {
    const fetchDiscounts = async () => {
      try {
        const token = localStorage.getItem('token')
        const response = await axios.get(
          'http://localhost:8080/api/payment/discount-types/Active',
          {
            headers: {
              accept: 'application/json',
              Authorization: `Bearer ${token}`,
            },
          }
        )
        if (response.data && Array.isArray(response.data)) {
          setDiscounts(response.data)
        }
      } catch (error) {
        console.error('Error fetching discount types:', error)
      }
    }

    fetchDiscounts()
  }, [])

  useEffect(() => {
    const fetchCharges = async () => {
      try {
        const token = localStorage.getItem('token')
        const response = await axios.get(
          'http://localhost:8080/api/payment/charge-types/Active',
          {
            headers: {
              accept: 'application/json',
              Authorization: `Bearer ${token}`,
            },
          }
        )
        if (response.data && Array.isArray(response.data)) {
          setCharges(response.data)
        }
      } catch (error) {
        console.error('Error fetching charge types:', error)
      }
    }

    fetchCharges()
  }, [])

  // Add a function to handle selecting a table
  const handleTableSelection = async (table) => {
    setSelectedTables((prevSelectedTables) => {
      const newSelectedTables = prevSelectedTables.includes(table.ID)
        ? prevSelectedTables.filter((id) => id !== table.ID)
        : [...prevSelectedTables, table.ID]

      // ถ้าเลือกโต๊ะใหม่ ให้ดึงข้อมูลรายการอาหาร
      if (!prevSelectedTables.includes(table.ID)) {
        fetchTableBillableItems(table.ID, table.UUID)
      }

      return newSelectedTables
    })
  }

  const calculateTotal = () => {
    const subtotal = calculateTotalWithSelectedTables()
    const vatAmount = subtotal * 0.07

    // คำนวณส่วนลดทั้งหมด
    const totalDiscount = selectedDiscounts.reduce((sum, discount) => {
      if (discount.discountID) {
        const discountType = discounts.find(d => d.ID === parseInt(discount.discountID))
        if (discountType) {
          const discountAmount = discountType.Type === 'percentage'
            ? (subtotal * (discountType.Value / 100))
            : discountType.Value
          return sum + discountAmount
        }
      }
      return sum
    }, 0)

    // คำนวณค่าใช้จ่ายเพิ่มเติมทั้งหมด
    const totalCharges = selectedCharges.reduce((sum, charge) => {
      if (charge.chargeID && charge.value) {
        const chargeType = charges.find(c => c.ID === parseInt(charge.chargeID))
        if (chargeType) {
          const quantity = parseInt(charge.value) || 1
          return sum + (chargeType.DefaultAmount * quantity)
        }
      }
      return sum
    }, 0)

    // คำนวณยอดรวมสุทธิ
    const total = subtotal + vatAmount - totalDiscount + totalCharges
    return Math.max(0, total)
  }

  const calculateTotalWithSelectedTables = () => {
    // คำนวณยอดรวมของโต๊ะหลัก
    const mainTableTotal = calculateSubtotal()

    // คำนวณยอดรวมของโต๊ะที่เลือกเพิ่ม
    const selectedTablesTotal = selectedTables.reduce((sum, tableId) => {
      if (tableId === tableID) return sum // ข้ามโต๊ะหลัก
      const tableItems = tableBillableItems[tableId] || []
      const tableTotal = tableItems.reduce((itemSum, item) => {
        let itemTotal = item.price * item.quantity
        if (item.options) {
          itemTotal += item.options.reduce((optSum, opt) => 
            optSum + (opt.price * opt.quantity), 0)
        }
        return itemSum + itemTotal
      }, 0)
      return sum + tableTotal
    }, 0)

    return mainTableTotal + selectedTablesTotal
  }

  const calculateSubtotal = () => {
    return billableItems.reduce((sum, item) => {
      let itemTotal = item.price * item.quantity
      // เพิ่มราคา options
      if (item.options && item.options.length > 0) {
        itemTotal += item.options.reduce((optSum, opt) => {
          return optSum + opt.price * opt.quantity
        }, 0)
      }

      return sum + itemTotal
    }, 0)
  }

  const calculateDiscounts = (subtotal) => {
    return selectedDiscounts.reduce((sum, discount) => {
      if (discount.value && discount.discountID) {
        const discountType = discounts.find((d) => d.ID === discount.discountID)
        const discountAmount =
          discountType.Type === 'percentage'
            ? (subtotal * parseFloat(discount.value)) / 100
            : parseFloat(discount.value)
        return sum + discountAmount
      }
      return sum
    }, 0)
  }

  const calculateCharges = () => {
    return selectedCharges.reduce((sum, charge) => {
      if (charge.value && charge.chargeID) {
        const chargeType = charges.find((c) => c.ID === charge.chargeID)
        return sum + parseFloat(charge.value)
      }
      return sum
    }, 0)
  }

  const justtotal = calculateSubtotal()
  const calculateVAT = () => {
    const subtotal = calculateSubtotal()
    return subtotal * 0.07
  }

  const addDiscount = () => {
    setSelectedDiscounts([...selectedDiscounts, { discountID: '', value: '' }])
  }

  const removeDiscount = (index) => {
    setSelectedDiscounts((prevDiscounts) =>
      prevDiscounts.filter((_, i) => i !== index)
    )
  }

  const handleDiscountChange = (index, field, value) => {
    const updatedDiscounts = [...selectedDiscounts]
    updatedDiscounts[index][field] = value
    setSelectedDiscounts(updatedDiscounts)
  }

  const handleChargeChange = (index, field, value) => {
    const newSelectedCharges = [...selectedCharges]
    newSelectedCharges[index][field] = value
    setSelectedCharges(newSelectedCharges)
  }

  const addCharge = () => {
    setSelectedCharges([...selectedCharges, { chargeID: '', value: 0 }])
  }

  const removeCharge = (index) => {
    const newSelectedCharges = selectedCharges.filter((_, i) => i !== index)
    setSelectedCharges(newSelectedCharges)
  }

  const handlePayment = async () => {
    if (!selectedPayment) {
      alert('Please select a payment method')
      return
    }

    // const totalAmount = calculateTotal()

    setIsProcessing(true)
    try {
      const token = localStorage.getItem('token')

      // ตรวจสอบการเลือกโต๊ะ
      const paymentData = {
        uuid: uuid,
        payment_method: selectedPayment, // วิธีการชำระเงิน
        service_charge: vat, // vat
        staff_id: user.id, // รหัสพนักงาน (เปลี่ยนให้เหมาะสม)

        discounts: selectedDiscounts
          .filter((discount) => discount.discountID) // กรองส่วนลดที่ถูกเลือก
          .map((discount) => ({
            discount_type_id: Number(discount.discountID),
            reason: discount.discountName || '', // ใช้ชื่อส่วนลด หรือ null หากไม่มีชื่อ
          })),

        extra_charges: selectedCharges
          .filter((charge) => charge.chargeID) // กรองค่าใช้จ่ายที่ถูกเลือก
          .map((charge) => ({
            charge_type_id: Number(charge.chargeID),
            note: charge.chargeOption || '',
            quantity: charge.quantity || 1,
          })),
      }

      let apiEndpoint = '/api/payment/process' // API สำหรับกรณีไม่มีโต๊ะที่เลือก
      if (selectedTables && selectedTables.length > 0) {
        apiEndpoint = '/api/v2/payment/merge' // API สำหรับกรณีเลือกโต๊ะหลายโต๊ะ
        paymentData.table_ids = selectedTables // ส่ง array ของ ID โต๊ะที่เลือกไป
      } else {
        paymentData.table_id = tableID // ถ้าไม่มีโต๊ะที่เลือก ให้ส่ง tableID
      }

      // ตรวจสอบการส่งข้อมูล
      console.log(paymentData)

      const response = await axios.post(
        `http://localhost:8080${apiEndpoint}`,
        paymentData,
        {
          headers: {
            accept: 'application/json',
            Authorization: `Bearer ${posToken}`,
          },
        }
      )

      if (response.data.success) {
        setShowSuccessModal(true)
      } else {
        // alert('Payment failed. Please try again.')
        setShowSuccessModal(true)
      }
    } catch (error) {
      console.error('Error during payment:', error)
      // alert('Payment failed. Please try again.')
      setShowSuccessModal(true)
    } finally {
      setIsProcessing(false)
    }
  }

  const SuccessModal = () => (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl p-8 max-w-md w-full mx-4">
        <div className="flex flex-col items-center text-center">
          <div className="w-16 h-16 bg-green-50 rounded-full flex items-center justify-center mb-4">
            <Check className="w-8 h-8 text-green-500" />
          </div>
          <h2 className="text-2xl font-medium mb-2">Payment Successful</h2>
          <p className="text-gray-600 mb-6">Thank you for dining with us!</p>
          <button
            onClick={() => (window.location.href = '/pos')}
            className="w-full bg-black text-white py-4 rounded-xl"
          >
            Return to Menu
          </button>
        </div>
      </div>
    </div>
  )

  const fetchTableBillableItems = async (tableId, uuid) => {
    try {
      const token = localStorage.getItem('token')
      const response = await axios.get(
        `http://localhost:8080/api/table/billable/${uuid}`,
        {
          headers: {
            accept: 'application/json',
            Authorization: `Bearer ${token}`,
          },
        }
      )

      if (response.data && response.data.items) {
        setTableBillableItems(prev => ({
          ...prev,
          [tableId]: response.data.items
        }))
      }
    } catch (error) {
      console.error('Error fetching table billable items:', error)
    }
  }

  // const calculateTotalForSelectedTables = () => {
  //   return selectedTables.reduce((total, tableId) => {
  //     const items = tableBillableItems[tableId] || []
  //     const tableTotal = items.reduce((sum, item) => {
  //       let itemTotal = item.price
  //       if (item.options) {
  //         itemTotal += item.options.reduce((optSum, opt) => 
  //           optSum + (opt.price * opt.quantity), 0)
  //       }
  //       return sum + itemTotal
  //     }, 0)
  //     return total + tableTotal
  //   }, 0)
  // }

  const handlePrintTableSelection = (tableId) => {
    setPrintSelectedTables(prev => {
      if (prev.includes(tableId)) {
        return prev.filter(id => id !== tableId)
      }
      return [...prev, tableId]
    })
  }

  // const handlePrintBillCheck = async () => {
  //   setIsPrinting(true)
  //   try {
  //     const token = localStorage.getItem('token')
  //     if (!token) {
  //       alert('Session หมดอายุ กรุณาเข้าสู่ระบบใหม่')
  //       window.location.href = '/login'
  //       return
  //     }

  //     // รวบรวม table IDs ที่จะพิมพ์
  //     const tableIDsToPrint = printSelectedTables.length > 0 
  //       ? printSelectedTables 
  //       : [tableID]

  //     const response = await axios.post(
  //       // 'http://localhost:8080/api/v2/printers/bill-check',
  //       'http://localhost:8080/api/printers/bill-check',
  //       {
  //         table_ids: tableIDsToPrint
  //       },
  //       {
  //         headers: {
  //           'Content-Type': 'application/json',
  //           Authorization: `Bearer ${token}`,
  //         },
  //       }
  //     )

  //     if (response.data) {
  //       alert('ส่งรายการไปยังเครื่องพิมพ์แล้ว')
  //       setShowPrintModal(false) // ปิด popup เมื่อพิมพ์สำเร็จ
  //       setPrintSelectedTables([]) // รีเซ็ตรายการโต๊ะที่เลือก
  //     }
  //   } catch (error) {
  //     console.error('Error printing bill check:', error)
  //     if (error.response) {
  //       console.error('Error details:', error.response.data)
  //       alert(error.response.data.error || 'ไม่สามารถพิมพ์ใบรายการอาหารได้')
  //     } else {
  //       alert('ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์')
  //     }
  //   } finally {
  //     setIsPrinting(false)
  //   }
  // }
  {showPrintModal && (
    <PrintBillCheckModal
      isOpen={showPrintModal}
      onClose={() => {
        setShowPrintModal(false)
        setPrintSelectedTables([])
      }}
      mainTable={{
        ID: tableID,
        Name: tableName
      }}
      occupiedTables={occupiedTables}
      billableItems={billableItems}
      tableBillableItems={tableBillableItems}
      selectedDiscounts={selectedDiscounts}
      selectedCharges={selectedCharges}
      serviceCharge={vat}
      user={user}
      discounts={discounts}
      charges={charges}
    />
  )}
  
  const CloseTableConfirmModal = () => (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl p-8 max-w-md w-full mx-4">
        <div className="flex flex-col items-center text-center">
          <div className="w-16 h-16 bg-red-50 rounded-full flex items-center justify-center mb-4">
            <X className="w-8 h-8 text-red-500" />
          </div>
          <h2 className="text-2xl font-medium mb-2">ยืนยันการปิดโต๊ะ</h2>
          <p className="text-gray-600 mb-6">
            การปิดโต๊ะจะทำได้เมื่อไม่มีรายการอาหารหรือรายการอาหารถูกยกเลิกทั้งหมดเท่านั้น
          </p>
          <div className="flex gap-4 w-full">
            <button
              onClick={() => setShowCloseConfirmModal(false)}
              className="flex-1 py-4 rounded-xl border border-gray-300"
            >
              ยกเลิก
            </button>
            <button
              onClick={handleCloseTable}
              disabled={isClosingTable}
              className={`flex-1 py-4 rounded-xl ${
                isClosingTable ? 'bg-gray-400' : 'bg-red-500'
              } text-white`}
            >
              {isClosingTable ? 'กำลังปิดโต๊ะ...' : 'ยืนยัน'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )

  // เพิ่ม useEffect สำหรับดึงข้อมูลรายการอาหารของโต๊ะอื่นๆ เมื่อ popup เปิด
  useEffect(() => {
    const fetchOtherTablesBillableItems = async () => {
      if (!showPrintModal || !occupiedTables) return;

      try {
        const token = localStorage.getItem('token')
        if (!token) return;

        // ดึงข้อมูลเฉพาะโต๊ะที่ยังไม่มีข้อมูล
        for (const table of occupiedTables) {
          if (table.ID !== tableID && !tableBillableItems[table.ID]) {
            try {
              const response = await axios.get(
                `http://localhost:8080/api/table/billable/${table.UUID}`,
                {
                  headers: {
                    accept: 'application/json',
                    Authorization: `Bearer ${posToken}`,
                  },
                }
              )
              if (response.data && response.data.items) {
                setTableBillableItems(prev => ({
                  ...prev,
                  [table.ID]: response.data.items
                }))
              }
            } catch (error) {
              console.error(`Error fetching items for table ${table.ID}:`, error)
            }
          }
        }
      } catch (error) {
        console.error('Error fetching other tables billable items:', error)
      }
    }

    fetchOtherTablesBillableItems()
  }, [showPrintModal, occupiedTables, tableID, tableBillableItems, posToken])

  const handleCloseTable = async () => {
    setIsClosingTable(true)
    try {
      const token = localStorage.getItem('token')
      if (!token) {
        alert('Session หมดอายุ กรุณาเข้าสู่ระบบใหม่')
        window.location.href = '/login'
        return
      }

      const response = await axios.post(
        `http://localhost:8080/api/table/close/${tableID}?uuid=${uuid}`,
        {},
        {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
        }
      )

      if (response.data) {
        alert('ปิดโต๊ะสำเร็จ')
        window.location.href = '/pos'
      }
    } catch (error) {
      console.error('Error closing table:', error)
      if (error.response) {
        alert(error.response.data.error || 'ไม่สามารถปิดโต๊ะได้')
      } else {
        alert('ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์')
      }
    } finally {
      setIsClosingTable(false)
      setShowCloseConfirmModal(false)
    }
  }

  return (
    <div className="flex flex-col min-h-screen bg-neutral-100 p-5 ">
      {showSuccessModal && <SuccessModal />}
      {showPrintModal && (
        <PrintBillCheckModal
          isOpen={showPrintModal}
          onClose={() => {
            setShowPrintModal(false)
            setPrintSelectedTables([])
          }}
          mainTable={{
            ID: tableID,
            Name: tableName
          }}
          occupiedTables={occupiedTables}
          billableItems={billableItems}
          tableBillableItems={tableBillableItems}
          selectedDiscounts={selectedDiscounts}
          selectedCharges={selectedCharges}
          serviceCharge={vat}
          user={user}
          discounts={discounts}
          charges={charges}
        />
      )}
      {showCloseConfirmModal && <CloseTableConfirmModal />}

      <header className="bg-black text-white p-6 rounded-lg">
        <button onClick={() => (window.location.href = '/pos')}>
          <div className="flex items-center gap-4 text-gold">
            <ChevronLeft className="w-6 h-6" />
            <h1 className="text-xl font-light tracking-wider">PAYMENT</h1>
          </div>
        </button>
      </header>

      <main className="flex justify-center p-2 ">
        <div className="flex col-3 gap-4">
          <div className="w-3/12">
            {/* Payment Method */}
            <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
              <h2 className="text-lg font-medium mb-4">ช่องทางการชำระ</h2>
              <div className="space-y-3">
                <button
                  onClick={() => setSelectedPayment('credit')}
                  className={`w-full flex items-center p-4 rounded-xl border ${
                    selectedPayment === 'credit'
                      ? 'border-black bg-gray-50'
                      : 'border-gray-200'
                  }`}
                >
                  <CreditCard className="w-5 h-5 mr-3" />
                  <span>Credit Card</span>
                </button>
                <button
                  onClick={() => setSelectedPayment('qr')}
                  className={`w-full flex items-center p-4 rounded-xl border ${
                    selectedPayment === 'qr'
                      ? 'border-black bg-gray-50'
                      : 'border-gray-200'
                  }`}
                >
                  <Wallet className="w-5 h-5 mr-3" />
                  <span>QR Payment</span>
                </button>
              </div>
            </div>
            {/* End payment method */}

            {/* เพิ่มปุ่มพิมพ์ใบรายการอาหาร */}
            <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
              <button
                onClick={() => setShowPrintModal(true)}
                className="w-full py-4 rounded-xl bg-blue-500 text-white text-lg mb-4"
              >
                พิมพ์ใบรายการอาหาร
              </button>
              <button
                onClick={() => setShowCloseConfirmModal(true)}
                className="w-full py-4 rounded-xl bg-red-500 text-white text-lg"
              >
                ปิดโต๊ะ
              </button>
            </div>

            {/* Discount Section */}
            <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
              <h2 className="text-lg font-medium mb-4">ส่วนลด</h2>
              <div className="space-y-3">
                {selectedDiscounts.map((discount, index) => (
                  <div key={index} className="flex space-x-3">
                    <select
                      className="border p-2 rounded-md"
                      value={discount.discountID}
                      onChange={(e) =>
                        handleDiscountChange(
                          index,
                          'discountID',
                          e.target.value
                        )
                      }
                    >
                      <option value="">เลือกส่วนลด</option>
                      {discounts.length > 0 ? (
                        discounts.map((discountOption) => (
                          <option
                            key={discountOption.ID}
                            value={discountOption.ID}
                          >
                            {discountOption.Name} -{' '}
                            {discountOption.Type === 'percentage'
                              ? `${discountOption.Value}%`
                              : `${discountOption.Value} ฿`}
                          </option>
                        ))
                      ) : (
                        <option value="">No discounts available</option>
                      )}
                    </select>
                    {/* ปุ่มลบส่วนลด */}
                    <button
                      onClick={() => removeDiscount(index)}
                      className=" text-red-500"
                    >
                      <X />
                    </button>
                  </div>
                ))}
                <button
                  onClick={addDiscount}
                  className="bg-blue-500 text-white p-2 rounded-md mt-2"
                >
                  เพิ่มส่วนลด
                </button>
              </div>
            </div>
            {/*  */}

            {/* Additional charges Section */}
            <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
              <h2 className="text-lg font-medium mb-4">ค่าใช้จ่ายเพิ่มเติม</h2>
              <div className="space-y-3">
                {selectedCharges.map((charge, index) => (
                  <div key={index} className="flex space-x-3">
                    <select
                      className="border p-2 rounded-md flex-grow"
                      value={charge.chargeID}
                      onChange={(e) =>
                        handleChargeChange(index, 'chargeID', e.target.value)
                      }
                    >
                      <option value="">เลือกค่าใช้จ่ายเพิ่มเติม</option>
                      {charges.map((chargeOption) => (
                        <option key={chargeOption.ID} value={chargeOption.ID}>
                          {chargeOption.Name} - {chargeOption.DefaultAmount} ฿
                        </option>
                      ))}
                    </select>
                    <input
                      type="number"
                      min="1"
                      className="border p-2 rounded-md w-14"
                      placeholder="จำนวน"
                      value={charge.value}
                      onChange={(e) =>
                        handleChargeChange(index, 'value', e.target.value)
                      }
                    />
                    <button
                      onClick={() => removeCharge(index)}
                      className=" text-red-500 "
                    >
                      <X />
                    </button>
                  </div>
                ))}
                <button
                  onClick={addCharge}
                  className="bg-blue-500 text-white p-2 rounded-md mt-2"
                >
                  เพิ่มค่าใช้จ่าย
                </button>
              </div>
            </div>
          </div>
          {/* End additional charges */}

          {/* main payment */}
          <div className="bg-white rounded-2xl p-6 mb-6 shadow-md w-7/12 ">
            <div className="mx-60"></div>
            <div className="flex justify-between items-center mb-4 ">
              <h2 className="text-lg font-medium">Order Summary</h2>
              <span className="text-lg font-medium text-gray-500">
                Table {tableName}
                {tableID}-{uuid}
              </span>
              {/* <span className="text-sm text-gray-500">Table {uuid}</span> */}
            </div>

            {/* <OrderSummaryDetail billableItems={billableItems} /> */}
            <OrderSummaryDetail
              billableItems={billableItems}
              user={user} // เพิ่ม user
              tableID={tableID} // เพิ่ม tableID
              onCancelItem={(updatedItems) => {
                // Optional: อัพเดท billableItems หลังยกเลิกรายการ
                setBillableItems(updatedItems)
              }}
            />

            <div className="bottom-0 ">
              <div className="border-t border-gray-100 mt-4 pt-4 space-y-3">
                {/* แสดงราคาโต๊ะหลัก */}
                <div className="flex justify-between text-gray-600">
                  <span>โต๊ะ {tableName} (หลัก)</span>
                  <span>{justtotal.toLocaleString()} ฿</span>
                </div>

                {/* แสดงราคาโต๊ะที่เลือกเพิ่ม */}
                {selectedTables.length > 0 && (
                  <div className="space-y-2">
                    <div className="font-medium">โต๊ะที่รวมบิล:</div>
                    {selectedTables.map((tableId) => {
                      const tableItems = tableBillableItems[tableId] || []
                      const tableTotal = tableItems.reduce((sum, item) => {
                        let itemTotal = item.price * item.quantity
                        if (item.options) {
                          itemTotal += item.options.reduce(
                            (optSum, opt) => optSum + opt.price * opt.quantity,
                            0
                          )
                        }
                        return sum + itemTotal
                      }, 0)

                      const table = occupiedTables.find((t) => t.ID === tableId)
                      if (table && tableId !== tableID) {
                        // ไม่แสดงโต๊ะหลัก
                        return (
                          <div
                            key={tableId}
                            className="flex justify-between text-gray-600 pl-4"
                          >
                            <span>โต๊ะ {table.Name}</span>
                            <span>{tableTotal.toLocaleString()} ฿</span>
                          </div>
                        )
                      }
                      return null
                    })}
                  </div>
                )}

                {/* แสดงยอดรวมก่อนส่วนลดและค่าใช้จ่ายเพิ่มเติม */}
                <div className="flex justify-between text-gray-600 font-medium border-t pt-2">
                  <span>ยอดรวมทุกโต๊ะ</span>
                  <span>
                    {calculateTotalWithSelectedTables().toLocaleString()} ฿
                  </span>
                </div>

                {/* แสดงรายการส่วนลด */}
                {selectedDiscounts.some((d) => d.discountID) && (
                  <div className="space-y-2">
                    <div className="font-medium">ส่วนลด:</div>
                    {selectedDiscounts.map((discount, index) => {
                      if (discount.discountID) {
                        const discountType = discounts.find(
                          (d) => d.ID === parseInt(discount.discountID)
                        )
                        if (discountType) {
                          const discountAmount =
                            discountType.Type === 'percentage'
                              ? justtotal * (discountType.Value / 100)
                              : discountType.Value
                          return (
                            <div
                              key={index}
                              className="flex justify-between text-green-600 pl-4"
                            >
                              <span>
                                {discountType.Name} (
                                {discountType.Type === 'percentage'
                                  ? `${discountType.Value}%`
                                  : `${discountType.Value} ฿`}
                                )
                              </span>
                              <span>-{discountAmount.toLocaleString()} ฿</span>
                            </div>
                          )
                        }
                      }
                      return null
                    })}
                  </div>
                )}

                {/* แสดงรายการค่าใช้จ่ายเพิ่มเติม */}
                {selectedCharges.some((c) => c.chargeID) && (
                  <div className="space-y-2">
                    <div className="font-medium">ค่าใช้จ่ายเพิ่มเติม:</div>
                    {selectedCharges.map((charge, index) => {
                      if (charge.chargeID && charge.value) {
                        const chargeType = charges.find(
                          (c) => c.ID === parseInt(charge.chargeID)
                        )
                        if (chargeType) {
                          const quantity = parseInt(charge.value) || 1
                          const totalCharge =
                            chargeType.DefaultAmount * quantity
                          return (
                            <div
                              key={index}
                              className="flex justify-between text-red-600 pl-4"
                            >
                              <span>
                                {chargeType.Name} ({quantity} x{' '}
                                {chargeType.DefaultAmount} ฿)
                              </span>
                              <span>+{totalCharge.toLocaleString()} ฿</span>
                            </div>
                          )
                        }
                      }
                      return null
                    })}
                  </div>
                )}

                <div className="flex justify-between text-gray-600">
                  <span>VAT (7%)</span>
                  <span>
                    {(
                      calculateTotalWithSelectedTables() * 0.07
                    ).toLocaleString()}{' '}
                    ฿
                  </span>
                </div>

                <div className="flex justify-between font-bold text-lg">
                  <span>ยอดรวมทั้งหมด</span>
                  <span>{calculateTotal().toLocaleString()} ฿</span>
                </div>
                <button
                  onClick={handlePayment}
                  disabled={isProcessing}
                  className={`px-8 py-4 rounded-xl w-full ${
                    isProcessing ? 'bg-gray-400' : 'bg-black'
                  } text-white text-lg`}
                >
                  {isProcessing ? 'Processing...' : 'ชำระเงิน'}
                </button>
              </div>
            </div>
          </div>
          {/* End main payment */}

          <div className="w-4/12">
            {/* Table Selection Section */}
            <div className="bg-white rounded-2xl p-6 mb-6 shadow-md min-h-screen ">
              <h2 className="text-lg font-medium mb-4">
                เลือกโต๊ะเพื่อรวมค่าชำระ
              </h2>
              <div className="mx-24"></div>
              <div className="space-y-3">
                {occupiedTables && occupiedTables.length > 0 ? (
                  <div className="flex flex-col space-y-3">
                    {occupiedTables.map((table) => (
                      <div key={table.ID}>
                        <div
                          className={`flex justify-between cursor-pointer p-3 rounded-lg border ${
                            selectedTables.includes(table.ID)
                              ? 'bg-gray-200 border-gray-500'
                              : 'border-gray-300'
                          }`}
                          onClick={() => handleTableSelection(table)}
                        >
                          <div>
                            <span className="text-gray-700 mr-2">
                              Table {table.ID}
                            </span>
                            <span className="text-gray-500">{table.Name}</span>
                          </div>
                          {tableBillableItems[table.ID] && (
                            <span className="text-gray-600">
                              {tableBillableItems[table.ID]
                                .reduce((sum, item) => sum + item.price, 0)
                                .toLocaleString()}{' '}
                              ฿
                            </span>
                          )}
                        </div>

                        {/* แสดงรายการอาหารเมื่อเลือกโต๊ะ */}
                        {selectedTables.includes(table.ID) &&
                          table.UUID !== uuid && // ไม่แสดงถ้าเป็นโต๊ะเดียวกับโต๊ะหลัก
                          tableBillableItems[table.ID] && (
                            <div className="ml-4 mt-2">
                              <OrderSummaryDetail
                                billableItems={tableBillableItems[table.ID]}
                                user={user}
                                tableID={table.ID}
                                readOnly={true} // เพิ่ม prop เพื่อไม่ให้แก้ไขรายการได้
                              />
                            </div>
                          )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-gray-500">No occupied tables available.</p>
                )}
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* <div className="flex justify-center">
        <div className="fixed bottom-0 w-full lg:w-10/12 bg-white p-6 border-t border-gray-200 shadow-xl">
          <div className="flex justify-between items-center mb-4 px-5">
            <div>
              <p className="text-gray-600">Total Amount</p>
              <p className="text-2xl font-medium">
                {calculateTotal().toLocaleString()} ฿
              </p>
            </div>
            <button
              onClick={handlePayment}
              disabled={isProcessing}
              className={`px-8 py-4 rounded-xl ${
                isProcessing ? 'bg-gray-400' : 'bg-black'
              } text-white text-lg`}
            >
              {isProcessing ? 'Processing...' : 'Pay Now'}
            </button>
          </div>
        </div>
      </div> */}
    </div>
  )
}

export default PaymentTables