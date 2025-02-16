import React, { useState, useEffect, useRef } from 'react'
import { Carousel } from 'flowbite-react'
import MenuItem from './MenuItem'
import { Plus, Minus, PlusIcon } from 'lucide-react'
import useCartStore from '../hooks/cart-store'
import { useSearchParams } from 'react-router-dom'

const API_BASE_URL = 'http://127.0.0.1:8080/api'

export default function MenuList({ language }) {
  const [categories, setCategories] = useState([])
  const [menus, setMenus] = useState([])
  const [activeLink, setActiveLink] = useState(0)
  const [promotions, setPromotions] = useState([])
  const [quantity, setQuantity] = useState(1)
  const [note, setNote] = useState('')
  const [isPopupOpen, setIsPopupOpen] = useState(false)
  const [displayItem, setDisplayItem] = useState(null)
  const [selectedOptions, setSelectedOptions] = useState([])
  const [searchParams] = useSearchParams()

  const tableId = parseInt(searchParams.get('tableID'), 10)
  const uuid = searchParams.get('uuid')

  const scrollRef = useRef(null)
  const { addPromotion, getPromotionOrderedQuantity, setTableData } = useCartStore()

  const handleCategoryClick = (category) => {
    setActiveLink(category)
  }

  const handleIncrease = () => setQuantity((prev) => prev + 1)
  const handleDecrease = () => {
    setQuantity((prev) => (prev > 1 ? prev - 1 : 1))
  }

  const togglePopup = () => setIsPopupOpen((prev) => !prev)

  const handlePromotionClick = (promotion) => {
    setDisplayItem(promotion)
    setIsPopupOpen(true)
  }

  useEffect(() => {
    if (tableId && uuid) {
      setTableData(tableId, uuid)
    }
  }, [tableId, uuid, setTableData])

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

    if (quantity > 0 && displayItem) {
      const currentOrdered = getPromotionOrderedQuantity(displayItem.ID)
      const maxAvailable = displayItem.Quantity || Infinity

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

      if (selectedOptions.length < displayItem.MinSelections) {
        alert(
          language === 'th'
            ? `กรุณาเลือกอย่างน้อย ${displayItem.MinSelections} รายการ`
            : language === 'en'
            ? `Please select at least ${displayItem.MinSelections} items`
            : `请至少选择${displayItem.MinSelections}个项目`
        )
        return
      }

      const promotion = {
        ...displayItem,
        Price: displayItem.Price || 0,
        selectedOptions: selectedOptions,
      }

      addPromotion(promotion, quantity)

      setNote('')
      setQuantity(1)
      setSelectedOptions([])
      setIsPopupOpen(false)
    }
  }

  const handleOptionChange = (menuItem) => {
    const isSelected = selectedOptions.some((opt) => opt.id === menuItem.ID)

    if (isSelected) {
      setSelectedOptions((prev) => prev.filter((opt) => opt.id !== menuItem.ID))
    } else {
      if (selectedOptions.length < displayItem.MaxSelections) {
        const currentItemCount = selectedOptions.filter(
          (opt) => opt.id === menuItem.ID
        ).length
        const itemInPromotion = displayItem.Items.find(
          (item) => item.MenuItem.ID === menuItem.ID
        )

        if (currentItemCount < itemInPromotion.Quantity) {
          setSelectedOptions((prev) => [
            ...prev,
            {
              id: menuItem.ID,
              name: menuItem.Name,
              price: menuItem.Price,
              quantity: 1,
            },
          ])
        } else {
          alert(
            `คุณไม่สามารถเลือกเมนู ${menuItem.Name} ได้มากกว่า ${itemInPromotion.Quantity} ครั้ง`
          )
        }
      } else {
        alert(
          `คุณสามารถเลือกได้สูงสุด ${displayItem.MaxSelections} รายการเท่านั้น`
        )
      }
    }
  }

  useEffect(() => {
    const fetchCategoryData = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/categories`)
        const data = await response.json()
        setCategories(data)
      } catch (error) {
        console.error('Error fetching data:', error)
      }
    }

    const fetchMenuData = async () => {
      try {
        const response = await fetch(
          `${API_BASE_URL}/menu/ActiveMenu`
        )
        const data = await response.json()
        setMenus(data)
      } catch (error) {
        console.error('Error fetching data:', error)
      }
    }

    const fetchPromotionData = async () => {
      try {
        const response = await fetch(
          `${API_BASE_URL}/promotions/Active`,
          {
            headers: { accept: 'application/json' },
          }
        )
        if (response.ok) {
          const data = await response.json()
          setPromotions(data)
        } else {
          console.error('Failed to fetch promotions')
        }
      } catch (error) {
        console.error('Error fetching promotions:', error)
      }
    }

    fetchCategoryData()
    fetchMenuData()
    fetchPromotionData()
  }, [])

  return (
    <div className="mt-[5.5rem] bg-gradient-to-br from-[#1c2025] via-[#34393f] to-[#6f757c] max-w-screen-xl mx-auto">
      <div className="bg-[#1c2025] border-b border-[#3D3038]/30">
        <div 
          className="flex items-center gap-1.5 overflow-x-auto scrollbar-hidden py-2 px-3"
          ref={scrollRef}
        >
          <button
            className={`px-3 py-1.5 rounded-full transition-all duration-300 ${
              activeLink === 0
                ? 'bg-[#3D3038] text-[#E6C65C] shadow-sm'
                : 'bg-[#F8F5F2] text-[#2D232A] hover:bg-[#F0EBE9]'
            }`}
            onClick={() => handleCategoryClick(0)}
          >
            <span className="text-xs font-medium">
              {language === 'th' ? 'ทั้งหมด' : language === 'en' ? 'All' : '全部'}
            </span>
          </button>

          {categories.map((item) => (
            <button
              key={item.ID}
              className={`px-3 py-1.5 rounded-full transition-all duration-300 ${
                activeLink === item.ID
                  ? 'bg-[#3D3038] text-[#E6C65C] shadow-sm'
                  : 'bg-[#F8F5F2] text-[#2D232A] hover:bg-[#F0EBE9]'
              }`}
              onClick={() => handleCategoryClick(item.ID)}
            >
              <span className="text-xs font-medium whitespace-nowrap">
                {language === 'th'
                  ? item.Name
                  : language === 'en'
                  ? item.NameEn
                  : item.NameCh}
              </span>
            </button>
          ))}
        </div>
        <div className="h-[1px] bg-gradient-to-r from-[#3D3038] via-[#3D3038]/50 to-transparent"></div>
      </div>

      <div className="overflow-y-auto">
        <div className="my-4 h-48 sm:h-64 xl:h-80 2xl:h-96 py-2 mx-2">
          <Carousel leftControl=" " rightControl=" ">
            {promotions.length > 0 ? (
              promotions.map((promotion) => (
                <div key={promotion.ID} className="relative">
                  <img
                    className="w-full h-full object-cover"
                    src={`data:image/png;base64,${promotion.Image}`}
                    alt={promotion.Name}
                    onClick={() => handlePromotionClick(promotion)}
                  />
                </div>
              ))
            ) : (
              <p className="text-center text-lg">
                {language === 'th'
                  ? 'ไม่มีโปรโมชั่นในขณะนี้'
                  : language === 'en'
                  ? 'No promotions at the moment'
                  : '暂无促销'}
              </p>
            )}
          </Carousel>
        </div>

        {menus.some(menu => menu.IsRecommended) && (
          <div className="mx-2 my-4">
            <h2 className="text-base font-medium text-white/90 mb-2 px-1">
              {language === 'th' 
                ? 'เมนูแนะนำ' 
                : language === 'en' 
                ? 'Recommended Menu' 
                : '推荐菜单'}
            </h2>
            
            <div className="grid grid-cols-2 gap-2">
              {menus
                .filter(menu => menu.IsRecommended)
                .map(menu => (
                  <MenuItem key={menu.ID} item={menu} language={language} isPremium />
                ))}
            </div>
          </div>
        )}

        <div className="mx-2 my-4">
          {categories.map((category) => {
            const filteredMenus = activeLink === 0
              ? menus.filter(menu => menu.CategoryID === category.ID && !menu.IsRecommended)
              : menus.filter(menu => menu.CategoryID === category.ID && menu.CategoryID === activeLink && !menu.IsRecommended);

            if (filteredMenus.length === 0) return null;

            return (
              <div key={category.ID} className="mb-6">
                <h2 className="text-base font-medium text-white/90 mb-2 px-1">
                  {language === 'th'
                    ? category.Name
                    : language === 'en'
                    ? category.NameEn
                    : category.NameCh}
                </h2>
                
                <div className="grid grid-cols-2 gap-2">
                  {filteredMenus.map(menu => (
                    <MenuItem key={menu.ID} item={menu} language={language} />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {isPopupOpen && displayItem && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl w-full max-w-xl overflow-y-auto max-h-screen">
            <div className="relative">
              <button
                onClick={togglePopup}
                className="absolute right-4 top-4 text-white/70 hover:text-white z-10 bg-black/50 rounded-full p-2"
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

              <img
                src={`data:image/png;base64,${displayItem.Image}`}
                alt={displayItem.Name}
                className="w-full aspect-video object-cover"
              />

              <div className="p-6">
                <h3 className="text-black text-2xl font-medium mb-2">
                  {displayItem.Name}
                </h3>
                <p className="text-black text-xl font-semibold mb-6">
                  {displayItem.Price} THB
                </p>

                <div className="mb-4 p-3 bg-gray-100 rounded-lg">
                  <p className="text-sm text-gray-600">
                    {language === 'th'
                      ? `ต้องเลือกอย่างน้อย ${displayItem.MinSelections} รายการ และไม่เกิน ${displayItem.MaxSelections} รายการ`
                      : language === 'en'
                      ? `Select minimum ${displayItem.MinSelections} and maximum ${displayItem.MaxSelections} items`
                      : `最少选择 ${displayItem.MinSelections} 项，最多选择 ${displayItem.MaxSelections} 项`}
                  </p>
                  <p className="text-sm text-gray-600 mt-1">
                    {language === 'th'
                      ? `เลือกแล้ว: ${selectedOptions.length} รายการ`
                      : language === 'en'
                      ? `Selected: ${selectedOptions.length} items`
                      : `已选择: ${selectedOptions.length} 项`}
                  </p>
                </div>

                {displayItem &&
                  displayItem.Items &&
                  Array.isArray(displayItem.Items) &&
                  displayItem.Items.length > 0 && (
                    <div className="mb-4">
                      {displayItem.Items.map((item) => (
                        <div key={item.ID} className="mb-2">
                          <div className="flex items-center justify-between">
                            <div>
                              <h4 className="text-black text-lg font-medium">
                                {item.MenuItem.Name}-{item.MenuItem.ID}
                              </h4>
                              <p className="text-sm text-gray-600">
                                {item.MenuItem.Price} THB
                              </p>
                            </div>
                            <div className="flex items-center space-x-2">
                              <input
                                type="checkbox"
                                id={`checkbox-${item.MenuItem.ID}`}
                                checked={selectedOptions.some(
                                  (opt) => opt.id === item.MenuItem.ID
                                )}
                                onChange={() =>
                                  handleOptionChange(item.MenuItem)
                                }
                                disabled={
                                  selectedOptions.length >=
                                    displayItem.MaxSelections &&
                                  !selectedOptions.some(
                                    (opt) => opt.id === item.MenuItem.ID
                                  )
                                }
                              />
                              <label
                                htmlFor={`checkbox-${item.MenuItem.ID}`}
                                className="text-sm"
                              >
                                {selectedOptions.some(
                                  (opt) => opt.id === item.MenuItem.ID
                                )
                                  ? 'Selected'
                                  : 'Select'}
                              </label>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}

                <div className="my-4 flex justify-center items-center gap-6">
                  <button
                    className="bg-gray-200 p-2 rounded-md"
                    onClick={handleDecrease}
                    disabled={quantity === 1}
                  >
                    <Minus className="size-4" />
                  </button>
                  <span className="text-lg">{quantity}</span>
                  <button
                    className="bg-gray-200 p-2 rounded-md"
                    onClick={handleIncrease}
                  >
                    <PlusIcon className="size-4" />
                  </button>
                </div>

                <textarea
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  placeholder="Add your special request here..."
                  className="w-full p-3 bg-gray-100 border border-gray/30 rounded-lg"
                  rows="3"
                />

                <button
                  onClick={handleAddToCart}
                  className="bg-blue-500 hover:bg-blue-600 text-white py-3 w-full rounded-lg font-medium"
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
        </div>
      )}
    </div>
  )
}
