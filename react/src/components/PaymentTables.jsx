import React, { useState, useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import { ChevronLeft, CreditCard, Wallet, Info, Check } from 'lucide-react'
import axios from 'axios'

const PaymentTables = ({ user }) => {
  const [selectedTable, setSelectedTable] = useState(null)
  const [selectedPayment, setSelectedPayment] = useState('')
  const [vat, setVat] = useState(7)
  const [tipPercentage, setTipPercentage] = useState(7)
  const [isProcessing, setIsProcessing] = useState(false)
  const [showSuccessModal, setShowSuccessModal] = useState(false)
  const [promoCode, setPromoCode] = useState('')
  const [promoError, setPromoError] = useState('')
  const [appliedDiscount, setAppliedDiscount] = useState(0)
  const location = useLocation()
  const { tableID, uuid, occupiedTables } = location.state || {}

  const [billableItems, setBillableItems] = useState([])
  const [charges, setCharges] = useState([])
  const [selectedCharges, setSelectedCharges] = useState([])
  const [discounts, setDiscounts] = useState([])
  const [selectedDiscounts, setSelectedDiscounts] = useState([
    { discountID: '', value: '' },
  ])
  console.log('Occupied Tables:', occupiedTables)
  console.log('User:', user)

  const [cardDetails, setCardDetails] = useState({
    number: '',
    expiry: '',
    cvc: '',
    name: '',
  })
  const [cardErrors, setCardErrors] = useState({})

  useEffect(() => {
    const fetchBillableItems = async () => {
      const token = localStorage.getItem('token')
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
        const response = await axios.get(
          'http://localhost:8080/api/payment/discount-types/Active',
          {
            headers: { accept: 'application/json' },
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
        const response = await axios.get(
          'http://localhost:8080/api/payment/charge-types/Active',
          {
            headers: { accept: 'application/json' },
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
  const handleTableSelection = (table) => {
    setSelectedTable(table) // Set selected table
  }

  const calculateTotal = () => {
    const subtotal = calculateSubtotal()
    const tipAmount = (subtotal * tipPercentage) / 100
    const vatAmount = calculateVAT()

    const totalDiscount = selectedDiscounts.reduce((sum, discount) => {
      if (discount.value) {
        const discountAmount =
          discount.discountID &&
          discounts.find((d) => d.ID === discount.discountID)
            ? discounts.find((d) => d.ID === discount.discountID).Type ===
              'percentage'
              ? (subtotal * parseFloat(discount.value)) / 100
              : parseFloat(discount.value)
            : 0
        return sum + discountAmount
      }
      return sum
    }, 0)

    const totalCharges = selectedCharges.reduce((sum, charge) => {
      if (charge.value) {
        const chargeAmount =
          charge.chargeID && charges.find((c) => c.ID === charge.chargeID)
            ? parseFloat(charge.value)
            : 0
        return sum + chargeAmount
      }
      return sum
    }, 0)

    const total =
      subtotal + tipAmount + vatAmount - totalDiscount + totalCharges

    return Math.max(0, total)
  }

  const calculateSubtotal = () => {
    return billableItems.reduce((sum, item) => sum + item.price, 0)
  }

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

    const totalAmount = calculateTotal()

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

      // ตรวจสอบว่าเลือกโต๊ะเพื่อรวมค่าชำระหรือไม่
      let apiEndpoint = '/api/payment/process'
      if (selectedTable) {
        // เช็คว่าเลือกโต๊ะในการรวมค่าชำระ
        apiEndpoint = '/api/v2/payment/merge'
        paymentData.table_ids = [tableID, Number(selectedTable.ID)] // ส่งโต๊ะที่เลือกไปในการรวม
      } else {
        // ถ้าไม่ได้เลือกโต๊ะเพื่อรวมชำระ ให้ส่ง table_id
        paymentData.table_id = tableID
      }

      // ตรวจสอบการส่งข้อมูล
      console.log(paymentData)

      const response = await axios.post(
        `http://localhost:8080${apiEndpoint}`,
        paymentData,
        {
          headers: {
            accept: 'application/json',
            Authorization: `Bearer ${token}`,
          },
        }
      )

      if (response.data.success) {
        setShowSuccessModal(true)
      } else {
        alert('Payment failed. Please try again.')
      }
    } catch (error) {
      console.error('Error during payment:', error)
      alert('Payment failed. Please try again.')
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
            onClick={() => (window.location.href = '/')}
            className="w-full bg-black text-white py-4 rounded-xl"
          >
            Return to Menu
          </button>
        </div>
      </div>
    </div>
  )

  return (
    <div className="flex flex-col min-h-screen bg-neutral-100 p-5">
      {showSuccessModal && <SuccessModal />}

      <header className="bg-black text-white p-6 rounded-lg">
        <div className="flex items-center gap-4 text-gold">
          <ChevronLeft className="w-6 h-6" />
          <h1 className="text-xl font-light tracking-wider">PAYMENT</h1>
        </div>
      </header>

      <main className="flex-1 p-6 pb-24">
        <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-medium">Order Summary</h2>
            <span className="text-sm text-gray-500">Table {tableID}</span>
            <span className="text-sm text-gray-500">Table {uuid}</span>
          </div>
          {billableItems.map((item) => (
            <div key={item.id} className="flex justify-between mb-3">
              <div>
                <span className="text-gray-800">{item.quantity}x </span>
                <span>{item.name}</span>
              </div>
              <span>{item.price.toLocaleString()} ฿</span>
            </div>
          ))}

          <div className="border-t border-gray-100 mt-4 pt-4 space-y-3">
            <div className="flex justify-between text-gray-600">
              <span>VAT (7%)</span>
              <span>{calculateVAT().toLocaleString()} ฿</span>
            </div>
            <div className="flex justify-between text-gray-600">
              <span>Subtotal</span>
              <span>{calculateTotal().toLocaleString()} ฿</span>
            </div>
          </div>
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
                    handleDiscountChange(index, 'discountID', e.target.value)
                  }
                >
                  <option value="">None</option>
                  {discounts.length > 0 ? (
                    discounts.map((discountOption) => (
                      <option key={discountOption.ID} value={discountOption.ID}>
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
                  className="bg-red-500 text-white p-2 rounded-md"
                >
                  Remove
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

        {/* Additional charges Section */}
        <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
          <h2 className="text-lg font-medium mb-4">ค่าใช้จ่ายเพิ่มเติม</h2>
          <div className="space-y-3">
            {selectedCharges.map((charge, index) => (
              <div key={index} className="flex space-x-3">
                <select
                  className="border p-2 rounded-md"
                  value={charge.chargeID}
                  onChange={(e) =>
                    handleChargeChange(index, 'chargeID', e.target.value)
                  }
                >
                  <option value="">None</option>
                  {charges.length > 0 ? (
                    charges.map((chargeOption) => (
                      <option key={chargeOption.ID} value={chargeOption.ID}>
                        {chargeOption.Name} - {chargeOption.DefaultAmount} ฿
                      </option>
                    ))
                  ) : (
                    <option value="">No additional charges available</option>
                  )}
                </select>
                <input
                  type="number"
                  className="border p-2 rounded-md"
                  placeholder="Amount"
                  value={charge.value}
                  onChange={(e) =>
                    handleChargeChange(index, 'value', e.target.value)
                  }
                />
                {/* ปุ่มลบค่าใช้จ่ายเพิ่มเติม */}
                <button
                  onClick={() => removeCharge(index)}
                  className="bg-red-500 text-white p-2 rounded-md"
                >
                  Remove
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

        {/* Table Selection Section */}
        <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
          <h2 className="text-lg font-medium mb-4">
            เลือกโต๊ะที่ต้องการรวมค่าชำระ
          </h2>
          <div className="space-y-3">
            {occupiedTables && occupiedTables.length > 0 ? (
              <div className="flex flex-col space-y-3">
                {occupiedTables.map((table, index) => (
                  <div
                    key={index}
                    className={`flex cursor-pointer p-3 rounded-lg border ${
                      selectedTable?.tableID === table.tableID
                        ? 'bg-gray-200 border-gray-500'
                        : 'border-gray-300'
                    }`}
                    onClick={() => handleTableSelection(table)}
                  >
                    <span className="text-gray-700">Table {table.ID}</span>
                    <span className="text-gray-500">{table.Name}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-gray-500">No occupied tables available.</p>
            )}
          </div>
        </div>

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
      </main>

      <div className="flex justify-center">
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
      </div>
    </div>
  )
}

export default PaymentTables
