import React, { useState } from 'react'
import { Plus, Minus, PlusIcon } from 'lucide-react'
import useCartStore from '../hooks/cart-store'
import { Carousel } from 'flowbite-react'
const PromotionItem = ({ item, promotion }) => {
  const [isPopupOpen, setIsPopupOpen] = useState(false) // Popup state
  const [note, setNote] = useState('') // Note for the kitchen
  const [selectedOptions, setSelectedOptions] = useState({}) // Store selected options for each group
  const [quantity, setQuantity] = useState(1)
  const [selectedPromotion, setSelectedPromotion] = useState(null)
  const { addToCart } = useCartStore()

  const handleIncrease = () => setQuantity((prev) => prev + 1) // เพิ่มจำนวน
  const handleDecrease = () => {
    setQuantity((prev) => (prev > 1 ? prev - 1 : 1))
  }

  // Toggle popup visibility
  const togglePopup = () => setIsPopupOpen((prev) => !prev)

  // Handle adding item to cart
  const handleAddToCart = () => {
    if (quantity > 0) {
      const finalItem = {
        ...displayItem, // รวมข้อมูลจาก item หรือ promotion
        Price: displayItem.Price || 0, // กำหนดค่า Price เริ่มต้นถ้า undefined
      }

      // ตรวจสอบว่า finalItem มีข้อมูลที่จำเป็นครบถ้วนหรือไม่
      addToCart(finalItem, quantity, note, selectedOptions)
      setNote('') // Clear the note
      setSelectedOptions({}) // Clear selected options
      setQuantity(1) // Reset quantity
      setIsPopupOpen(false) // Close the popup after adding to cart
    }
    if (selectedPromotion && quantity > 0) {
      const promotionItem = {
        ID: selectedPromotion.ID,
        Name: selectedPromotion.Name,
        Price: selectedPromotion.Price,
        Image: selectedPromotion.Image,
        quantity: quantity,
        note: note,
        isPromotion: true,
      }
      addToCart(promotionItem)
      handleClosePopup()
    }
  }

  // Handle option selection change
  const handleOptionChange = (groupName, optionName, price) => {
    setSelectedOptions((prev) => ({
      ...prev,
      [groupName]: { optionName, price },
    }))
  }

  // กำหนดให้แสดงข้อมูลจาก item หรือ promotion
  const displayItem = item || promotion

  return (
    <div >
      {/* แสดงรูปภาพเมนูหรือโปรโมชัน */}
      <div className="my-4 h-48 sm:h-64 xl:h-80 2xl:h-96 border border-black">
        <Carousel leftControl=" " rightControl=" ">
          {displayItem ? (
            <div key={displayItem.ID} className="relative">
              <div className="bg-white border border-gold rounded-md shadow-md">
                {/* แสดงรูปภาพเมนูหรือโปรโมชัน */}
                {displayItem.Image ? (
                  <img
                    src={`data:image/png;base64,${displayItem.Image}`}
                    alt={displayItem.Name}
                    className="w-full h-full object-cover"
                    onClick={togglePopup} // เมื่อคลิกจะแสดง Popup
                  />
                ) : (
                  <div className="placeholder-image">No Image Available</div>
                )}
              </div>

              <div className="p-2 flex flex-col gap-4">
                <h3
                  className="text-lg text-gold font-semibold"
                  onClick={togglePopup}
                >
                  {displayItem.Name}
                </h3>
                <div className="flex justify-between items-center">
                  <p className="text-base text-red-700 font-semibold">
                    {displayItem.Price} THB
                  </p>
                  <button
                    onClick={togglePopup}
                    className="bg-green-400 hover:bg-green-500 rounded-full p-1 text-white"
                  >
                    <Plus className="size-5" />
                  </button>
                </div>
              </div>

              {/* <div className="absolute bottom-0 left-0 right-0 bg-black/50 text-white p-4">
              <h2 className="text-lg font-semibold">{displayItem.Name}</h2>
              <p className="text-sm">{displayItem.Description}</p>
            </div> */}
            </div>
          ) : (
            <p className="text-center text-lg">ไม่มีข้อมูลในขณะนี้</p>
          )}
        </Carousel>
      </div>

      {isPopupOpen && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50 p-4">
          <div className="bg-white border border-gray/50 rounded-xl w-full max-w-xl overflow-y-auto max-h-screen">
            <div className="relative">
              <button
                onClick={togglePopup}
                className="absolute right-4 top-4 text-white/70 hover:text-white z-10 
                bg-black/50 rounded-full p-2"
              >
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>

              {displayItem.Image ? (
                <img
                  src={`data:image/png;base64,${displayItem.Image}`}
                  alt={displayItem.Name || 'Item'}
                  className="w-full aspect-video object-cover"
                />
              ) : (
                <div className="w-full aspect-video bg-zinc-900 flex items-center justify-center">
                  <span className="text-gold/50 text-lg">
                    No Image Available
                  </span>
                </div>
              )}
            </div>

            <div className="p-6">
              <div className="flex justify-between">
                <h3 className="text-black text-2xl font-medium mb-2">
                  {displayItem.Name || 'Untitled Item'}
                </h3>
                <p className="text-black text-xl font-semibold mb-6">
                  {displayItem.Price || 0}&nbsp;บาท
                </p>
              </div>
              <p className="text-gray-400 mb-4">
                {displayItem.Description || 'No description available'}
              </p>

              {/* แสดงตัวเลือกต่างๆ หากมี */}
              {displayItem.OptionGroups &&
                displayItem.OptionGroups.length > 0 && (
                  <div className="mb-6">
                    {displayItem.OptionGroups.map((group) => (
                      <div key={group.ID}>
                        <h4 className="text-black text-lg font-medium">
                          {group.Name || 'Option Group'}
                        </h4>
                        <div className="flex flex-col gap-2 mt-2">
                          {group.Options.map((option) => (
                            <div
                              key={option.ID}
                              className="flex justify-start items-center gap-3 text-black/70 text-base"
                            >
                              <input
                                type="radio"
                                name={group.Name}
                                value={option.Name}
                                onChange={() =>
                                  handleOptionChange(
                                    group.Name,
                                    option.Name,
                                    option.Price
                                  )
                                }
                                className=""
                              />
                              {option.Name} (+{option.Price}฿)
                            </div>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                )}

              <div className="mb-4">
                <label className="block text-black/50 text-sm font-medium mb-2">
                  Special Instructions
                </label>
                <textarea
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  placeholder="Add your special request here..."
                  className="w-full px-4 py-3 bg-gray-100 border border-gray/30 rounded-lg 
                   placeholder-gray-500"
                  rows="3"
                />
              </div>

              {/* ปุ่มเพิ่ม-ลด จำนวน */}
              <div className="my-4 flex justify-center items-center gap-6">
                <button
                  className="bg-gray-200 hover:bg-gray-300 disabled:opacity-70 disabled:hover:bg-gray-200 p-2 rounded-md text-black"
                  onClick={handleDecrease}
                  disabled={quantity == 1}
                >
                  <Minus className="size-4" />
                </button>
                <span className="text-lg text-black font-semibold">
                  {quantity}
                </span>
                <button
                  className="bg-gray-200 hover:bg-gray-300 p-2 rounded-md text-black"
                  onClick={handleIncrease}
                >
                  <PlusIcon className="size-4" />
                </button>
              </div>

              {/* เปลี่ยนปุ่ม Close เป็น Add to Cart */}
              <button
                onClick={handleAddToCart}
                className="bg-blue-500 hover:bg-blue-600 text-white py-3 w-full rounded-lg px-4 font-medium transition-colors"
              >
                Add to Cart
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default PromotionItem
