import React, { useState } from 'react'
import styles from '../styles/MenuItem.module.css'
import { Minus, Plus, PlusIcon, } from 'lucide-react'
import useCartStore from '../hooks/cart-store'

const MenuItem = ({ item, language }) => {
  const [isPopupOpen, setIsPopupOpen] = useState(false) // Popup state
  const [note, setNote] = useState('') // Note for the kitchen
  const [selectedOptions, setSelectedOptions] = useState({})
  
  const [quantity, setQuantity] = useState(1)

  const { addToCart, getMenuItemOrderedQuantity } = useCartStore()

  const handleIncrease = () => setQuantity((prev) => prev + 1) // เพิ่มจำนวน
  const handleDecrease = () => {
    setQuantity((prev) => (prev > 1 ? prev - 1 : 1))
  }

  // Toggle popup visibility
  const togglePopup = () => setIsPopupOpen((prev) => !prev)

  const validateRequiredOptions = () => {
    if (!item.OptionGroups) return true

    const missingRequiredGroups = item.OptionGroups.filter(
      (group) => group.IsRequired
    ).filter((group) => !selectedOptions[group.ID])

    if (missingRequiredGroups.length > 0) {
      const missingGroupNames = missingRequiredGroups
        .map((group) =>
          language === 'th'
            ? group.Name
            : language === 'en'
            ? group.NameEn
            : group.NameCh
        )
        .join(', ')

      alert(
        language === 'th'
          ? `กรุณาเลือกตัวเลือกที่จำเป็นในกลุ่ม: ${missingGroupNames}`
          : language === 'en'
          ? `Please select required options in groups: ${missingGroupNames}`
          : `请选择必需的选项组：${missingGroupNames}`
      )
      return false
    }
    return true
  }

  // Handle adding item to cart
  const handleAddToCart = () => {
    if (quantity > 0) {
      // ตรวจสอบ required options ก่อน
      if (!validateRequiredOptions()) {
        return
      }

      const currentOrdered = getMenuItemOrderedQuantity(item.ID)
      const maxAvailable = item.MaxQuantity || Infinity

      if (currentOrdered + quantity > maxAvailable) {
        alert(
          language === 'th'
            ? `ไม่สามารถสั่งเพิ่มได้ เนื่องจากเกินจำนวนที่กำหนด (สูงสุด ${maxAvailable} รายการ)`
            : language === 'en'
            ? `Cannot order more. Exceeds maximum limit (${maxAvailable} items)`
            : `无法订购更多。超过最大限制（${maxAvailable}项）`
        )
        return
      }

      addToCart(item, quantity, note, Object.values(selectedOptions))
      setNote('')
      setSelectedOptions({})
      setQuantity(1)
      setIsPopupOpen(false)
    }
  }

  // Handle option selection change
  const handleOptionChange = (groupID, option) => {
    setSelectedOptions((prev) => ({
      ...prev,
      [groupID]: {
        menu_option_id: option.ID,
        name: option.Name, // เพิ่มชื่อ option
        price: option.Price, // เพิ่มราคาของ option
      },
    }))
  }

  return (
    <div
      className={
        'bg-white rounded-md shadow-xl hover:shadow-xl transition-shadow duration-300'
      }
    >
      {/* Menu Item Display */}
      {item.Image ? (
        <img
          src={`data:image/png;base64,${item.Image}`}
          alt={item.Name}
          className={
            'h-28 sm:h-36 w-full rounded-tl-md rounded-tr-md object-cover'
          }
          onClick={togglePopup} // When clicked, toggle the popup
        />
      ) : (
        <div className="h-28">No Image Available</div>
      )}

      <div className={'p-2 flex flex-col gap-4 bg-blackpremiem rounded-b-md'}>
        <h3
          className="text-md text-goldlight font-semibold"
          onClick={togglePopup}
        >
          {/* แสดงชื่อเมนูตามภาษาที่เลือก */}
          {language === 'th'
            ? item.Name
            : language === 'en'
            ? item.NameEn
            : item.NameCh}
        </h3>

        <div className={'flex justify-between items-center'}>
          <p className="text-base text-goldpre font-semibold">{item.Price} ฿</p>
          <button
            onClick={togglePopup}
            className=" bg-red-500 hover:bg-gold rounded-full p-1 text-white"
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

              {item.Image ? (
                <img
                  src={`data:image/png;base64,${item.Image}`}
                  alt={item.Name || 'Menu item'}
                  className="w-full aspect-video object-cover"
                />
              ) : (
                <div className="w-full aspect-video bg-zinc-900 flex items-center justify-center">
                  <span className="text-goldpre text-lg">
                    No Image Available
                  </span>
                </div>
              )}
            </div>

            <div className="p-6">
              <div className="flex justify-between">
                <h3 className="text-black text-2xl font-medium mb-2">
                  {/* แสดงชื่อเมนูใน Popup ตามภาษาที่เลือก */}
                  {language === 'th'
                    ? item.Name
                    : language === 'en'
                    ? item.NameEn
                    : item.NameCh}
                </h3>
                <p className="text-black text-xl font-semibold mb-6">
                  {item.Price || 0}&nbsp;บาท
                </p>
              </div>
              <p className="text-gray-400 mb-4">
                {/* แสดงคำอธิบายเมนูใน Popup ตามภาษาที่เลือก */}
                {language === 'th'
                  ? item.Description
                  : language === 'en'
                  ? item.DescriptionEn
                  : item.DescriptionCh}
              </p>

              {/* Displaying Option Groups if they exist */}
              {item.OptionGroups && item.OptionGroups.length > 0 && (
                <div className="mb-6">
                  {item.OptionGroups.map((group) => (
                    <div key={group.ID} className="mb-4">
                      <h4 className="text-black text-lg font-medium flex items-center gap-2">
                        {language === 'th'
                          ? group.Name
                          : language === 'en'
                          ? group.NameEn
                          : group.NameCh}
                        {group.IsRequired && (
                          <span className="text-red-500 text-sm">
                            {language === 'th'
                              ? '(จำเป็น)'
                              : language === 'en'
                              ? '(Required)'
                              : '(必需)'}
                          </span>
                        )}
                      </h4>
                      <div className="flex flex-col gap-2 mt-2">
                        {group.Options.map((option) => (
                          <div
                            key={option.ID}
                            className="flex justify-start items-center gap-3 text-black/70 text-base"
                          >
                            <input
                              type="radio"
                              name={`group-${group.ID}`}
                              value={option.ID}
                              checked={
                                selectedOptions[group.ID]?.menu_option_id ===
                                option.ID
                              }
                              onChange={() =>
                                handleOptionChange(group.ID, option)
                              }
                              className="cursor-pointer"
                            />
                            <label className="flex-1 cursor-pointer">
                              {language === 'th'
                                ? option.Name
                                : language === 'en'
                                ? option.NameEn
                                : option.NameCh}{' '}
                              {option.Price > 0 && `(+${option.Price}฿)`}
                            </label>
                          </div>
                        ))}
                      </div>
                      {group.IsRequired && !selectedOptions[group.ID] && (
                        <p className="text-red-500 text-sm mt-1">
                          {language === 'th'
                            ? 'กรุณาเลือกตัวเลือก'
                            : language === 'en'
                            ? 'Please select an option'
                            : '请选择一个选项'}
                        </p>
                      )}
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

              {/* ปุ่มเพิ่ม-ลด */}
              <div className="my-4 flex justify-center items-center gap-6">
                <button
                  className="bg-gray-200 hover:bg-gray-300 disabled:opacity-70 disabled:hover:bg-gray-200 p-2 rounded-md text-black"
                  onClick={handleDecrease}
                  disabled={quantity == 1}
                >
                  <Minus className="size-4" />
                </button>
                <span className="text-lg font-semibold">{quantity}</span>
                <button
                  className="bg-gray-200 hover:bg-gray-300 p-2 rounded-md text-black"
                  onClick={handleIncrease}
                >
                  <PlusIcon className="size-4" />
                </button>
              </div>

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

export default MenuItem
