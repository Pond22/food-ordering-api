import React, { useState, useEffect } from 'react'
import styles from '../styles/MenuItem.module.css'
import { Minus, Plus, PlusIcon, ThumbsUp } from 'lucide-react'
import useCartStore from '../hooks/cart-store'
import { useSearchParams } from 'react-router-dom'


const MenuItem = ({ item, language, isPremium }) => {
  const [isPopupOpen, setIsPopupOpen] = useState(false) // Popup state
  const [note, setNote] = useState('') // Note for the kitchen
  const [selectedOptions, setSelectedOptions] = useState({})
  const [quantity, setQuantity] = useState(1)
  const [searchParams] = useSearchParams() // เพิ่มการใช้ useSearchParams

  const tableId = parseInt(searchParams.get('tableID'), 10)
  const uuid = searchParams.get('uuid')

  const { addToCart, getMenuItemOrderedQuantity, setTableData } = useCartStore()

  // เพิ่ม useEffect เพื่อเซ็ต tableId และ uuid
  useEffect(() => {
    if (tableId && uuid) {
      setTableData(tableId, uuid)
    }
  }, [tableId, uuid, setTableData])

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
    if (!tableId || !uuid) {
      alert(
        language === 'th'
          ? 'ไม่พบข้อมูลโต๊ะ กรุณาลองใหม่อีกครั้ง'
          : language === 'en'
          ? 'Table information not found. Please try again.'
          : '未找到餐桌信息。请重试。'
      )
      return
    }

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
    <div className={`relative bg-[#F8F3F2] rounded-lg overflow-hidden border-b border-[#3D3038] hover:bg-[#312B37] transition-colors`}>
      {/* Premium Badge */}
      {isPremium && (
        <div className="absolute top-2 left-2 z-10">
          <div className="flex items-center gap-1 bg-[#3B5780]/90 px-2 py-1 rounded-sm text-xs font-medium text-white shadow-sm">
            <ThumbsUp className="w-3 h-3" />
            <span>
              {language === 'th' 
                ? 'แนะนำ' 
                : language === 'en' 
                ? 'Best' 
                : '推荐'}
            </span>
          </div>
        </div>
      )}

      {/* รูปอาหาร */}
      <div className="relative aspect-[4/3] bg-gradient-to-b from-[#3D3038] to-[#2A2530]">
        {item.Image ? (
          <img
            src={`data:image/png;base64,${item.Image}`}
            alt={item.Name}
            className="w-full h-full object-cover opacity-95"
            onClick={togglePopup}
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <span className="text-[#4D4048]">400 x 300</span>
          </div>
        )}
        <div className="absolute inset-0 bg-gradient-to-t from-[#2A2530] to-transparent opacity-40"></div>
      </div>

      {/* ข้อมูลเมนู */}
      <div className="p-3 relative">
        <h3 
          className="text-sm font-medium text-[#231F20] mb-1 line-clamp-2 drop-shadow-sm hover:text-[#F3D77F] transition-colors" 
          onClick={togglePopup}
        >
          {language === 'th'
            ? item.Name
            : language === 'en'
            ? item.NameEn
            : item.NameCh}
        </h3>

        <div className="mb-2">
          <p className="text-xs text-gray-600 line-clamp-1 group-hover:text-gray-700">
            {language === 'th'
              ? item.Description
              : language === 'en'
              ? item.DescriptionEn
              : item.DescriptionCh}
            {((language === 'th' && item.Description?.length > 50) ||
              (language === 'en' && item.DescriptionEn?.length > 50) ||
              (language === 'ch' && item.DescriptionCh?.length > 50)) && (
              <button
                onClick={(e) => {
                  e.stopPropagation(); // ป้องกันการ trigger togglePopup
                  togglePopup();
                }}
                className="ml-1 text-[#F8F5F2] hover:text-[#AA1818] font-medium inline-flex items-center"
              >
                ...read more
              </button>
            )}
          </p>
        </div>

        <div className="flex justify-between items-center">
          <p className="text-sm font-medium text-[#D9BC91] drop-shadow-sm">
            {item.Price} ฿
          </p>
          <button
            onClick={togglePopup}
            className="bg-[#3D3038] hover:bg-[#4D4048] rounded-full p-1.5 text-[#E6C65C] transition-colors shadow-lg hover:shadow-xl"
          >
            <Plus className="size-4" />
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
                  viewBox="0 0 24 24"ก
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
