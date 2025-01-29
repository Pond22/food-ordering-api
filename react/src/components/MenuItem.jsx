import React, { useState } from 'react'
import { Plus, Minus, PlusIcon } from 'lucide-react'
import useCartStore from '../hooks/cart-store'
import styles from '../styles/MenuItem.module.css'

const MenuItem = ({ item, promotion, language }) => {
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

  // ฟังก์ชันการเพิ่มสินค้าไปที่ตะกร้า
  const handleAddToCart = () => {
    if (quantity > 0) {
      const finalItem = {
        ...displayItem,
        Price: displayItem.Price || 0,
        selectedOptions, // ส่ง selectedOptions ไปที่ตะกร้า
      }
      console.log('Adding Item to Cart: ', finalItem)
      addToCart(finalItem, quantity, note, selectedOptions)

      setNote('') // รีเซ็ต note
      setSelectedOptions({}) // รีเซ็ต selectedOptions
      setQuantity(1) // รีเซ็ต quantity
      setIsPopupOpen(false) // ปิด Popup
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
        selectedOptions, // ส่ง selectedOptions ในรูปแบบเดียวกับที่ API ต้องการ
      }
      console.log('Adding Promotion to Cart: ', promotionItem)
      addToCart(promotionItem, quantity, note, selectedOptions) // ส่งโปรโมชันไปที่ addToCart
      handleClosePopup()
    }
  }

  // ฟังก์ชันที่ใช้สำหรับเพิ่มตัวเลือก
  const handleOptionChange = (groupName, optionID, price) => {
    setSelectedOptions((prev) => ({
      ...prev,
      [groupName]: { menuOptionID: optionID, price }, // ใช้ menuOptionID แทน optionName
    }))
    console.log('Selected Options Updated: ', { ...selectedOptions })
  }

  // กำหนดให้แสดงข้อมูลจาก item หรือ promotion
  const displayItem = item || promotion

  return (
    <div
      className={
        'bg-gradient-to-b from-black/80 via-black/50 to-white/10 border border-white/30 rounded-md shadow-xl hover:shadow-xl transition-shadow duration-300'
      }
    >
      {/* แสดงรูปภาพเมนูหรือโปรโมชัน */}
      {displayItem.Image ? (
        <img
          src={`data:image/png;base64,${displayItem.Image}`}
          alt={displayItem.Name}
          className={
            'h-28 sm:h-36 w-full rounded-tl-md rounded-tr-md object-cover'
          }
          onClick={togglePopup} // เมื่อคลิกจะแสดง Popup
        />
      ) : (
        <div className="placeholder-image h-28 sm:h-36 w-full">
          No Image Available
        </div>
      )}

      <div className={'p-2 px-4 flex flex-col gap-4'}>
        <h3
          className="text-lg bg-gradient-to-t from-yellow-400 to-yellow-600 text-transparent bg-clip-text"
          onClick={togglePopup}
        >
          {language === 'th'
            ? displayItem.Name
            : language === 'en'
            ? displayItem.NameEn
            : displayItem.NameCh}
        </h3>
        <div className={'flex justify-between items-center'}>
          <p className="text-base text-red-700 font-semibold">
            {displayItem.Price} THB
          </p>
          <button
            onClick={togglePopup}
            className=" bg-gold hover:bg-green-500 rounded-full p-1 text-white"
          >
            <Plus className="size-5" />
          </button>
        </div>
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
                  {language === 'th'
                    ? displayItem.Name
                    : language === 'en'
                    ? displayItem.NameEn
                    : displayItem.NameCh}
                </h3>
                <p className="text-black text-xl font-semibold mb-6">
                  {displayItem.Price || 0}&nbsp;บาท
                </p>
              </div>
              <p className="text-gray-400 mb-4">
                {language === 'th'
                  ? displayItem.Description
                  : language === 'en'
                  ? displayItem.DescriptionEn
                  : displayItem.DescriptionCh}
              </p>

              {/* แสดงตัวเลือกต่างๆ หากมี */}
              {displayItem.OptionGroups.map((group, groupIndex) => (
                <div key={`group-${groupIndex}-${group.ID}`}>
                  <h4 className="text-black text-lg font-medium">
                    {language === 'th'
                      ? group.Name
                      : language === 'en'
                      ? group.NameEn
                      : group.NameCh}
                  </h4>
                  <div className="flex flex-col gap-2 mt-2">
                    {group.Options.map((option, optionIndex) => (
                      <div
                        key={`option-${groupIndex}-${optionIndex}-${option.ID}`} // ใช้ key ที่รวม groupIndex และ optionIndex
                        className="flex justify-start items-center gap-3 text-black/70 text-base"
                      >
                        <input
                          type="radio"
                          name={group.Name}
                          value={option.Name}
                          onChange={() =>
                            handleOptionChange(
                              group.Name,
                              option.ID, // ใช้ option.ID แทน option.Name
                              option.Price
                            )
                          }
                          className=""
                        />
                        {language === 'th'
                          ? option.Name
                          : language === 'en'
                          ? option.NameEn
                          : option.NameCh}{' '}
                        (+{option.Price}฿)
                      </div>
                    ))}
                  </div>
                </div>
              ))}

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

              {/*  Add to Cart */}
              <button
                onClick={handleAddToCart}
                className="bg-blue-500 hover:bg-blue-600 text-white py-3 w-full rounded-lg px-4 font-medium transition-colors"
              >
                {language === 'th'
                  ? 'เพิ่มไปยังตะกร้า'
                  : language === 'en'
                  ? 'Add to Cart'
                  : '加入购物车'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default MenuItem
