import React, { useState, useEffect } from 'react'
import { useLocation } from 'react-router-dom'
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
  const location = useLocation()
  const { tableID, uuid } = location.state || {}

  const [billableItems, setBillableItems] = useState([]) // Store billable items
  const [charges, setCharges] = useState([]) // Store charge types
  const [discounts, setDiscounts] = useState([]) // Store discount types
  const [selectedDiscounts, setSelectedDiscounts] = useState([
    { discountID: '', value: '' },
  ])

  const [cardDetails, setCardDetails] = useState({
    number: '',
    expiry: '',
    cvc: '',
    name: '',
  })
  const [cardErrors, setCardErrors] = useState({})

  // Fetch billable items
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

  // Fetch discount types
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

  // Fetch charge types
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

  // Calculate total (with discount, tip, charges)
  const calculateTotal = () => {
    const tipAmount =
      (billableItems.reduce((sum, item) => sum + item.price, 0) *
        tipPercentage) /
      100
    const total =
      billableItems.reduce((sum, item) => sum + item.price, 0) +
      tipAmount -
      appliedDiscount
    return Math.max(0, total)
  }

  // Add discount
  const addDiscount = () => {
    setSelectedDiscounts([...selectedDiscounts, { discountID: '', value: '' }])
  }

  const handleDiscountChange = (index, field, value) => {
    const updatedDiscounts = [...selectedDiscounts]
    updatedDiscounts[index][field] = value
    setSelectedDiscounts(updatedDiscounts)
  }

  const handlePayment = async () => {
    if (!selectedPayment) {
      alert('Please select a payment method')
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
              <span>Subtotal</span>
              <span>{calculateTotal().toLocaleString()} ฿</span>
            </div>
          </div>
        </div>

        {/* Discount Section */}
        <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
          <h2 className="text-lg font-medium mb-4">Discounts</h2>
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
                <input
                  type="number"
                  className="border p-2 rounded-md"
                  placeholder="Amount"
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
              Add Discount
            </button>
          </div>
        </div>

        {/* Additional charges */}
        <div className="bg-white rounded-2xl p-6 mb-6 shadow-md">
          <h2 className="text-lg font-medium mb-4">Additional Charges</h2>
          <div className="space-y-3">
            <select className="border p-2 rounded-md">
              <option value="">None</option>
              {charges.length > 0 ? (
                charges.map((charge) => (
                  <option key={charge.ID} value={charge.ID}>
                    {charge.Name} - {charge.DefaultAmount} ฿
                  </option>
                ))
              ) : (
                <option value="">No additional charges available</option>
              )}
            </select>
          </div>
        </div>

        {/* Payment Method */}
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
