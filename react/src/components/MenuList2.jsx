import React, { useState, useEffect, useRef } from 'react'
import { Carousel } from 'flowbite-react'
import MenuItem from './MenuItem'
import { Plus, Minus, PlusIcon, Languages } from 'lucide-react'
import useCartStore from '../hooks/cart-store'
// import PromotionItem from './PromotionItem'

export default function MenuList() {
  const [categories, setCategories] = useState([])
  const [menus, setMenus] = useState([])
  const [activeLink, setActiveLink] = useState(0)
  const [promotions, setPromotions] = useState([])
  const [quantity, setQuantity] = useState(1)
  const [note, setNote] = useState('')
  const [isPopupOpen, setIsPopupOpen] = useState(false)
  const [displayItem, setDisplayItem] = useState(null) // ใช้เก็บข้อมูลโปรโมชันที่เลือก
  const [language, setLanguage] = useState('th') // ภาษาเริ่มต้นคือไทย
  const scrollRef = useRef(null)
  const { addToCart } = useCartStore()

  const handleCategoryClick = (category) => {
    setActiveLink(category)
  }

  // ฟังก์ชันเพิ่มจำนวน
  const handleIncrease = () => setQuantity((prev) => prev + 1)

  // ฟังก์ชันลดจำนวน
  const handleDecrease = () => {
    setQuantity((prev) => (prev > 1 ? prev - 1 : 1))
  }

  // ฟังก์ชันเปิดปิด popup
  const togglePopup = () => setIsPopupOpen((prev) => !prev)

  // ฟังก์ชันคลิกโปรโมชันเพื่อแสดงข้อมูลใน popup
  const handlePromotionClick = (promotion) => {
    setDisplayItem(promotion)
    setIsPopupOpen(true)
  }

  // ฟังก์ชันเพิ่มสินค้าในตะกร้า
  const handleAddToCart = () => {
    if (displayItem) {
      const finalItem = {
        ...displayItem,
        Price: displayItem.Price || 0,
      }
      addToCart(finalItem, quantity, note)
      setNote('') // Clear note
      setQuantity(1) // Reset quantity
      setIsPopupOpen(false) // Close popup
    }
  }

  // ฟังก์ชันเลือกตัวเลือก
  const handleOptionChange = (groupName, optionName, price) => {
    setSelectedOptions((prev) => ({
      ...prev,
      [groupName]: { optionName, price },
    }))
  }

  useEffect(() => {
    const fetchCategoryData = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/categories')
        const data = await response.json()
        setCategories(data)
      } catch (error) {
        console.error('Error fetching data:', error)
      }
    }

    const fetchMenuData = async () => {
      try {
        const response = await fetch(
          'http://localhost:8080/api/menu/ActiveMenu'
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
          'http://localhost:8080/api/promotions/Active',
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

  // ฟังก์ชันเปลี่ยนภาษา
  const handleLanguageChange = (lang) => {
    setLanguage(lang)
  }

  // SVG pattern สำหรับลายทรายแบบสวนเซน
  const zenSandPattern = `
    <svg width="200" height="200" viewBox="0 0 200 200" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <!-- Pattern สำหรับเม็ดทราย -->
        <pattern id="sandGrain" x="0" y="0" width="4" height="4" patternUnits="userSpaceOnUse">
          <circle cx="2" cy="2" r="0.3" fill="rgba(255,255,255,0.15)"/>
        </pattern>
        
        <pattern id="zenSandPattern" x="0" y="0" width="200" height="200" patternUnits="userSpaceOnUse">
          <!-- ลายทรายพื้นฐาน -->
          <rect width="200" height="200" fill="url(#sandGrain)"/>
          
          <!-- ลายเส้นทรายหลัก -->
          <path d="M0,40 Q50,35 100,40 T200,40" stroke="rgba(255,255,255,0.2)" fill="none" stroke-width="0.8"/>
          <path d="M0,80 Q50,75 100,80 T200,80" stroke="rgba(255,255,255,0.2)" fill="none" stroke-width="0.8"/>
          <path d="M0,120 Q50,115 100,120 T200,120" stroke="rgba(255,255,255,0.2)" fill="none" stroke-width="0.8"/>
          <path d="M0,160 Q50,155 100,160 T200,160" stroke="rgba(255,255,255,0.2)" fill="none" stroke-width="0.8"/>
          
          <!-- ลายเส้นทรายรอง -->
          <path d="M0,20 Q50,22 100,20 T200,20" stroke="rgba(255,255,255,0.12)" fill="none" stroke-width="0.5"/>
          <path d="M0,60 Q50,62 100,60 T200,60" stroke="rgba(255,255,255,0.12)" fill="none" stroke-width="0.5"/>
          <path d="M0,100 Q50,102 100,100 T200,100" stroke="rgba(255,255,255,0.12)" fill="none" stroke-width="0.5"/>
          <path d="M0,140 Q50,142 100,140 T200,140" stroke="rgba(255,255,255,0.12)" fill="none" stroke-width="0.5"/>
          <path d="M0,180 Q50,182 100,180 T200,180" stroke="rgba(255,255,255,0.12)" fill="none" stroke-width="0.5"/>
          
          <!-- ลายวงกลมเซน -->
          <circle cx="100" cy="100" r="40" stroke="rgba(255,255,255,0.08)" fill="none" stroke-width="0.5"/>
          
          <!-- จุดเน้นตามแนวลายทราย -->
          <circle cx="50" cy="40" r="1" fill="rgba(255,255,255,0.2)"/>
          <circle cx="150" cy="80" r="1" fill="rgba(255,255,255,0.2)"/>
          <circle cx="100" cy="120" r="1" fill="rgba(255,255,255,0.2)"/>
        </pattern>
      </defs>
      <rect width="200" height="200" fill="url(#zenSandPattern)"/>
    </svg>
  `;

  const encodedPattern = `data:image/svg+xml,${encodeURIComponent(zenSandPattern)}`;

  return (
    <div
      className="mt-[4.5rem] bg-zinc-900  max-w-screen-xl mx-auto"
      style={{
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }}
    >
      {/* เลเยอร์พื้นหลักที่มีความละเอียด */}
      <div className=" inset-0 bg-gradient-to-b from-zinc-800 to-zinc-900 z-0 py-2 px-4 ">
        {/* เลเยอร์ลายทรายหลัก */}

        {/* เลเยอร์เม็ดทรายละเอียด */}
        <div
          className=" inset-0  z-0"
          style={{
            backgroundImage: `radial-gradient(circle at center, rgba(255,255,255,0.1) 0.5px, transparent 0.5px)`,
            backgroundSize: '8px 8px',
          }}
        >
          {/* เลเยอร์เงาละเอียด */}
          <div
            className=" inset-0 z-0"
            style={{
              backgroundImage: `linear-gradient(45deg, transparent 45%, rgba(255,255,255,0.03) 45%, rgba(255,255,255,0.03) 55%, transparent 55%)`,
              backgroundSize: '4px 4px',
            }}
          >
            {/* เลเยอร์เงาด้านบนและล่าง */}
            {/* <div className="absolute inset-0 bg-gradient-to-t from-black/50 via-transparent to-black/50 z-0"></div> */}

            {/* เลเยอร์ vignette */}
            {/* <div className="absolute inset-0 bg-radial-to-corner from-transparent to-black/40 z-0"></div> */}
            {/* ปุ่มเลือกภาษา */}
            <div>
              <button onClick={() => handleLanguageChange('th')}>
                <Languages /> ไทย
              </button>
              <button onClick={() => handleLanguageChange('en')}>
                <Languages /> English
              </button>
              <button onClick={() => handleLanguageChange('ch')}>
                <Languages /> 中文
              </button>
            </div>

            {/* CATEGORY */}
            <div
              className="mt-2 flex justify-start items-center gap-2 overflow-x-auto cursor-grab scrollbar-hidden"
              ref={scrollRef}
              style={{ whiteSpace: 'nowrap' }}
            >
              <button
                className={
                  activeLink === 0
                    ? 'bg-red-500 hover:bg-red-600 transition-colors px-4 py-1.5 rounded-md text-white'
                    : 'bg-gray-300 hover:bg-gray-400 transition-colors px-4 py-1.5 rounded-md text-black'
                }
                onClick={() => handleCategoryClick(0)}
              >
                {language === 'th'
                  ? 'ทั้งหมด'
                  : language === 'en'
                  ? 'All'
                  : '全部'}
              </button>

              {categories.map((item) => (
                <button
                  key={item.ID}
                  className={
                    activeLink === item.ID
                      ? 'bg-red-500 hover:bg-red-600 transition-colors px-4 py-1.5 rounded-md text-white'
                      : 'bg-gray-300 hover:bg-gray-400 transition-colors px-4 py-1.5 rounded-md text-black'
                  }
                  onClick={() => handleCategoryClick(item.ID)}
                >
                  {language === 'th'
                    ? item.Name
                    : language === 'en'
                    ? item.NameEn
                    : item.NameCh}
                </button>
              ))}
            </div>

            {/* PROMOTION CAROUSEL */}
            <div className="my-4 h-48 sm:h-64 xl:h-80 2xl:h-96 border border-black">
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

            {/* Popup แสดงรายละเอียดโปรโมชัน */}
            {isPopupOpen && displayItem && (
              <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50 p-4">
                <div className="bg-white border border-gray/50 rounded-xl w-full max-w-xl overflow-y-auto max-h-screen">
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

                    {/* แสดงรูปภาพโปรโมชัน */}
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

                  {/* แสดงรายละเอียดโปรโมชัน */}
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
                        {displayItem.Price || 0} THB
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

                    {displayItem.Items.map((item, index) => (
                      <div key={index}>
                        <h4 className="text-black text-lg font-medium">
                          {
                            language === 'th'
                              ? item.MenuItem.Name // ภาษาไทย
                              : language === 'en'
                              ? item.MenuItem.NameEn // ภาษาอังกฤษ
                              : language === 'ch'
                              ? item.MenuItem.NameCh // ภาษาจีน
                              : item.MenuItem.Name // หากไม่มีภาษาให้แสดงชื่อในภาษาเริ่มต้น
                          }
                        </h4>
                        {/* <div className="flex flex-col gap-2 mt-2">
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
                        </div> */}
                      </div>
                    ))}

                    {/* ปรับจำนวน */}
                    <div className="my-4 flex justify-center items-center gap-6">
                      <button
                        className="bg-gray-200 hover:bg-gray-300 disabled:opacity-70 disabled:hover:bg-gray-200 p-2 rounded-md text-black"
                        onClick={handleDecrease}
                        disabled={quantity === 1}
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
                    {/* ช่องรับหมายเหตุ */}
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

                    {/* ปุ่มเพิ่มไปยังตะกร้า */}
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

            {/* MENU */}
            <div className="my-4">
              {categories.map((category) => {
                const filteredMenus =
                  activeLink === 0
                    ? menus
                    : menus.filter((menu) => menu.Category.ID === activeLink)

                const categoryMenus = filteredMenus.filter(
                  (menu) => menu.Category.ID === category.ID
                )

                if (categoryMenus.length === 0) return null

                return (
                  <div
                    key={category.ID}
                    id={`category-${category.ID}`}
                    className="mb-6"
                  >
                    <h2 className="text-2xl font-semibold mb-2 bg-gradient-to-t from-yellow-400 to-yellow-600 text-transparent bg-clip-text">
                      {language === 'th'
                        ? category.Name
                        : language === 'en'
                        ? category.NameEn
                        : category.NameCh}
                    </h2>
                    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                      {categoryMenus.map((menu) => (
                        <MenuItem
                          key={menu.ID}
                          item={menu}
                          language={language}
                        />
                      ))}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
