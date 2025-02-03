import React, { useState, useEffect } from 'react'
import { ChevronLeft, CreditCard, Wallet, Info, Check } from 'lucide-react'
import axios from 'axios'

const PaymentTables = () => {
  const [selectedPayment, setSelectedPayment] = useState('')
  const [tipPercentage, setTipPercentage] = useState(10)
  const [isProcessing, setIsProcessing] = useState(false)
  const [showSuccessModal, setShowSuccessModal] = useState(false)
  const [promoCode, setPromoCode] = useState('')
  const [promoError, setPromoError] = useState('')
  const [appliedDiscount, setAppliedDiscount] = useState(0)
  const [cardDetails, setCardDetails] = useState({
    number: '',
    expiry: '',
    cvc: '',
    name: '',
  })
  const [cardErrors, setCardErrors] = useState({})
  // ข้อมูลจริง
  const [charges, setCharges] = useState([]) //เก็บข้อมูลค่าใช้จ่ายเพิ่มเติม
  const [discounts, setDiscounts] = useState([]) // เก็บข้อมูลประเภทส่วนลด
  const [selectedDiscounts, setSelectedDiscounts] = useState([
    { discountID: '', value: '' },
  ])

  // ดึงข้อมูลประเภทส่วนลดจาก API
  useEffect(() => {
    const fetchDiscounts = async () => {
      try {
        const response = await axios.get(
          'http://localhost:8080/api/payment/discount-types/Active',
          {
            headers: {
              accept: 'application/json',
            },
          }
        )

        // ตรวจสอบว่า API ส่งข้อมูลมาหรือไม่
        if (response.data && Array.isArray(response.data)) {
          setDiscounts(response.data) // เก็บข้อมูลประเภทส่วนลดที่ได้
        }
      } catch (error) {
        console.error('Error fetching discount types:', error)
      }
    }

    fetchDiscounts()
  }, [])

  // ฟังก์ชันที่ใช้ในการดึงข้อมูลประเภทค่าใช้จ่ายจาก API
  useEffect(() => {
    const fetchCharges = async () => {
      try {
        const response = await axios.get(
          'http://localhost:8080/api/payment/charge-types/Active',
          {
            headers: {
              accept: 'application/json',
            },
          }
        )

        if (response.data && Array.isArray(response.data)) {
          setCharges(response.data) // เก็บข้อมูลค่าใช้จ่ายเพิ่มเติมที่ได้รับ
        }
      } catch (error) {
        console.error('Error fetching charge types:', error)
      }
    }

    fetchCharges() // เรียกฟังก์ชันนี้เมื่อคอมโพเนนต์โหลด
  }, [])

  // ฟังก์ชั่นที่ใช้ในการเพิ่ม dropdown ใหม่
  const addDiscount = () => {
    setSelectedDiscounts([...selectedDiscounts, { discountID: '', value: '' }])
  }

  // ฟังก์ชั่นที่ใช้ในการอัปเดตค่าของ dropdown เมื่อเลือกประเภทส่วนลด
  const handleDiscountChange = (index, field, value) => {
    const updatedDiscounts = [...selectedDiscounts]
    updatedDiscounts[index][field] = value
    setSelectedDiscounts(updatedDiscounts)
  }
  // สิ้นสุดข้อมูลจริง

  const orderDetails = {
    tableNo: 'A15',
    orderId: 'ORD' + Math.random().toString(36).substring(2, 9).toUpperCase(),
    items: [
      { id: 1, name: 'Wagyu Beef Tenderloin', price: 2900, quantity: 1 },
      { id: 2, name: 'Norwegian Salmon', price: 890, quantity: 2 },
      { id: 3, name: 'Fine Wine Selection', price: 3500, quantity: 1 },
    ],
    subtotal: 8180,
    serviceCharge: 571.6,
    vat: 571.6,
  }

  const validateCard = () => {
    const errors = {}
    if (!cardDetails.number.replace(/\s/g, '').match(/^\d{16}$/)) {
      errors.number = 'Invalid card number'
    }
    if (!cardDetails.expiry.match(/^(0[1-9]|1[0-2])\/([0-9]{2})$/)) {
      errors.expiry = 'Invalid expiry date'
    }
    if (!cardDetails.cvc.match(/^\d{3,4}$/)) {
      errors.cvc = 'Invalid CVC'
    }
    if (!cardDetails.name.trim()) {
      errors.name = 'Name is required'
    }
    setCardErrors(errors)
    return Object.keys(errors).length === 0
  }

  const formatCardNumber = (value) => {
    const v = value.replace(/\s+/g, '').replace(/[^0-9]/gi, '')
    const matches = v.match(/\d{4,16}/g)
    const match = (matches && matches[0]) || ''
    const parts = []
    for (let i = 0, len = match.length; i < len; i += 4) {
      parts.push(match.substring(i, i + 4))
    }
    return parts.length ? parts.join(' ') : value
  }

  const handleCardInput = (e) => {
    const { name, value } = e.target
    if (name === 'number') {
      setCardDetails({ ...cardDetails, [name]: formatCardNumber(value) })
    } else if (name === 'expiry') {
      const v = value.replace(/\D/g, '').slice(0, 4)
      if (v.length >= 2) {
        setCardDetails({
          ...cardDetails,
          [name]: `${v.slice(0, 2)}/${v.slice(2)}`,
        })
      } else {
        setCardDetails({ ...cardDetails, [name]: v })
      }
    } else {
      setCardDetails({ ...cardDetails, [name]: value })
    }
    setCardErrors({ ...cardErrors, [name]: '' })
  }

  const calculateTotal = () => {
    const tipAmount = (orderDetails.subtotal * tipPercentage) / 100
    const total =
      orderDetails.subtotal +
      orderDetails.serviceCharge +
      orderDetails.vat +
      tipAmount -
      appliedDiscount
    return Math.max(0, total)
  }

  const handlePromoCode = () => {
    setPromoError('')
    if (promoCode.toUpperCase() === 'WELCOME20') {
      const discount = orderDetails.subtotal * 0.2
      setAppliedDiscount(discount)
      setPromoCode('')
    } else {
      setPromoError('Invalid promo code')
    }
  }

  const handlePayment = async () => {
    if (!selectedPayment) {
      alert('Please select a payment method')
      return
    }

    if (selectedPayment === 'credit' && !validateCard()) {
      return
    }

    setIsProcessing(true)
    try {
      await new Promise((resolve) => setTimeout(resolve, 2000))
      setShowSuccessModal(true)
    } catch (error) {
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
          <p className="text-sm text-gray-500 mb-4">
            Order ID: {orderDetails.orderId}
          </p>
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
    <div className="flex flex-col min-h-screen bg-neutral-100 p-5 mid-w-md ">
      {showSuccessModal && <SuccessModal />}

      <header className="bg-black text-white p-6 rounded-lg">
        <div className="flex items-center gap-4 text-gold">
          <ChevronLeft className="w-6 h-6" />
          <h1 className="text-xl font-light  tracking-wider">PAYMENT</h1>
        </div>
      </header>

      <main className="flex-1 p-6 pb-24">
        <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-medium">Order Summary</h2>
            <span className="text-sm text-gray-500">
              Table {orderDetails.tableNo}
            </span>
          </div>
          {orderDetails.items.map((item) => (
            <div key={item.id} className="flex justify-between mb-3">
              <div>
                <span className="text-gray-800">{item.quantity}x </span>
                <span>{item.name}</span>
              </div>
              <span>{(item.price * item.quantity).toLocaleString()} ฿</span>
            </div>
          ))}

          <div className="border-t border-gray-100 mt-4 pt-4 space-y-3 ">
            <div className="flex justify-between text-gray-600">
              <span>Subtotal</span>
              <span>{orderDetails.subtotal.toLocaleString()} ฿</span>
            </div>
            <div className="flex justify-between text-gray-600">
              <span>Service Charge (7%)</span>
              <span>{orderDetails.serviceCharge.toLocaleString()} ฿</span>
            </div>
            <div className="flex justify-between text-gray-600">
              <span>VAT (7%)</span>
              <span>{orderDetails.vat.toLocaleString()} ฿</span>
            </div>
            {appliedDiscount > 0 && (
              <div className="flex justify-between text-green-600">
                <span>Discount</span>
                <span>-{appliedDiscount.toLocaleString()} ฿</span>
              </div>
            )}
          </div>
        </div>

        {/* ส่วนลด */}
        <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
          <h2 className="text-lg font-medium mb-4">ส่วนลด</h2>
          <div className="space-y-3">
            {selectedDiscounts.map((discount, index) => (
              <div key={index} className="flex space-x-3">
                <select
                  className="border p-2 rounded-md"
                  value={discount.discountID} // ใช้ discountID แทน discount.ID
                  onChange={
                    (e) =>
                      handleDiscountChange(index, 'discountID', e.target.value) // ใช้ discountID แทน
                  }
                >
                  <option value="">ไม่มี</option>
                  {/* ตรวจสอบว่า discounts มีข้อมูลหรือไม่ */}
                  {discounts.length > 0 ? (
                    discounts.map((discountOption) => (
                      <option key={discountOption.ID} value={discountOption.ID}>
                        {discountOption.Name} -{' '}
                        {discountOption.Type === 'percentage'
                          ? `${discountOption.Value}%`
                          : `${discountOption.Value} บาท`}
                      </option>
                    ))
                  ) : (
                    <option value="">ไม่มีข้อมูลส่วนลด</option>
                  )}
                </select>
                <input
                  type="number"
                  className="border p-2 rounded-md"
                  placeholder="จำนวนเงิน"
                  value={discount.value}
                  onChange={(e) =>
                    handleDiscountChange(index, 'value', e.target.value)
                  }
                />
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

        {/* ค่าใช้จ่ายเพิ่มเติม */}
        <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
          <h2 className="text-lg font-medium mb-4">
            เลือกประเภทค่าใช้จ่ายเพิ่มเติม
          </h2>
          <div className="space-y-3">
            <select className="border p-2 rounded-md">
              <option value="">ไม่มี</option>
              {charges.length > 0 ? (
                charges.map((charge) => (
                  <option key={charge.ID} value={charge.ID}>
                    {charge.Name} - {charge.DefaultAmount} บาท
                  </option>
                ))
              ) : (
                <option value="">ไม่มีข้อมูลค่าใช้จ่าย</option>
              )}
            </select>
          </div>
        </div>

        {/* การเลือกช่องทางการชำระ */}
        <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
          <h2 className="text-lg font-medium mb-4">Payment Method</h2>
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
            <button
              onClick={() => setSelectedPayment('cash')}
              className={`w-full flex items-center p-4 rounded-xl border ${
                selectedPayment === 'cash'
                  ? 'border-black bg-gray-50'
                  : 'border-gray-200'
              }`}
            >
              <Wallet className="w-5 h-5 mr-3" />
              <span>ชำระด้วยเงินสด</span>
            </button>
          </div>
        </div>

        {selectedPayment === 'credit' && (
          <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
            <h2 className="text-lg font-medium mb-4">Card Details</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-gray-600 mb-2">
                  Card Number
                </label>
                <input
                  type="text"
                  name="number"
                  value={cardDetails.number}
                  onChange={handleCardInput}
                  placeholder="0000 0000 0000 0000"
                  maxLength="19"
                  className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:border-black"
                />
                {cardErrors.number && (
                  <p className="text-red-500 text-sm mt-1">
                    {cardErrors.number}
                  </p>
                )}
              </div>
              <div className="flex gap-4">
                <div className="flex-1">
                  <label className="block text-sm text-gray-600 mb-2">
                    Expiry Date
                  </label>
                  <input
                    type="text"
                    name="expiry"
                    value={cardDetails.expiry}
                    onChange={handleCardInput}
                    placeholder="MM/YY"
                    maxLength="5"
                    className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:border-black"
                  />
                  {cardErrors.expiry && (
                    <p className="text-red-500 text-sm mt-1">
                      {cardErrors.expiry}
                    </p>
                  )}
                </div>
                <div className="flex-1">
                  <label className="block text-sm text-gray-600 mb-2">
                    CVC
                  </label>
                  <input
                    type="text"
                    name="cvc"
                    value={cardDetails.cvc}
                    onChange={handleCardInput}
                    placeholder="000"
                    maxLength="4"
                    className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:outline-none focus:border-black"
                  />
                  {cardErrors.cvc && (
                    <p className="text-red-500 text-sm mt-1">
                      {cardErrors.cvc}
                    </p>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}

        {selectedPayment === 'qr' && (
          <div className="bg-white rounded-2xl p-6 mb-6 shadow-sm text-center">
            <img
              src="/api/placeholder/300/300"
              alt="QR Code"
              className="mx-auto mb-4 rounded-xl"
            />
            <p className="text-gray-600">
              Scan to pay {calculateTotal().toLocaleString()} ฿
            </p>
          </div>
        )}

        
      </main>

      <div className="flex justify-center">
        <div className="fixed bottom-0 w-full lg:w-10/12  bg-white p-6 border-t border-gray-200 shadow-xl">
          <div className="flex justify-between items-center mb-4 px-5">
            <div>
              <p className="text-gray-600 ">Total Amount</p>
              <p className="text-2xl font-medium">
                {calculateTotal().toLocaleString()} ฿
              </p>
            </div>
            <button
              onClick={handlePayment}
              disabled={isProcessing}
              className={`px-8 py-4 rounded-xl ${
                isProcessing
                  ? 'bg-gray-300 text-gray-500'
                  : 'bg-black text-gold'
              }`}
            >
              {isProcessing ? 'Processing...' : 'Proceed to Pay'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default PaymentTables
