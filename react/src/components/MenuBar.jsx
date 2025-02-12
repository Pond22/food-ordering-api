import React, { useState, useEffect } from 'react'
import { Modal } from 'flowbite-react'
import axios from 'axios'
import {
  Banknote,
  Bell,
  ShoppingCart,
  SquareMenu,
  User,
  PlusIcon,
  Minus,
  Languages,
} from 'lucide-react'
import useCartStore from '../hooks/cart-store'
import MenuList from './MenuList2'

export default function MenuBar({ tableID, uuid }) {
  const [openCallModal, setOpenCallModal] = useState(false)
  const [openCartModal, setOpenCartModal] = useState(false)
  const [language, setLanguage] = useState('th') // ภาษาเริ่มต้นคือไทย
  const {
    cart,
    promotions,
    increaseQuantity,
    decreaseQuantity,
    increaseQuantityPromo,
    decreaseQuantityPromo,
  } = useCartStore()

  // console.log(tableID, '+', uuid)

  const calculateTotal = () => {
    let total = cart.reduce(
      (total, item) =>
        total +
        item.menuItem.Price * item.quantity +
        item.options.reduce(
          (optionTotal, option) => optionTotal + option.price,
          0
        ), // รวมราคาของ option
      0
    )

    // คำนวณราคาโปรโมชัน
    total += promotions.reduce(
      (promoTotal, promo) =>
        promoTotal + promo.promotion.Price * promo.quantity,
      0
    )

    return total
  }

  useEffect(() => {
    calculateTotal()
  }, [cart])

  const totalQuantity =
    cart.reduce((sum, item) => sum + item.quantity, 0) +
    promotions.reduce((promoSum, promo) => promoSum + promo.quantity, 0) // Adding promo quantity

  // Create the order data with items and promotions
  const confirmOrder = async () => {
    console.log('Table ID:', tableID)
    console.log('UUID:', uuid)

    // สร้าง orderData สำหรับรายการอาหารปกติ
    const orderData = cart
      .map((item) => {
        if (item.quantity > 0) {
          return {
            menu_item_id: item.menuItem.ID,
            quantity: item.quantity,
            notes: item.notes || '',
            options: item.options.map((option) => ({
              menu_option_id: option.menu_option_id,
            })),
          }
        }
        return null
      })
      .filter((item) => item !== null)

    // สร้าง promotionData สำหรับโปรโมชัน
    const promotionData = promotions.map((promo) => ({
      promotion_id: promo.promotion_id,
      menu_item_ids: promo.selectedOptions.map((opt) => opt.id),
    }))

    // ตรวจสอบว่ามีรายการสั่งอาหารหรือโปรโมชันหรือไม่
    if (orderData.length === 0 && promotionData.length === 0) {
      alert(
        language === 'th'
          ? 'กรุณาเลือกรายการอาหารหรือโปรโมชัน'
          : language === 'en'
          ? 'Please select food items or promotions'
          : '请选择食品或促销项目'
      )
      return
    }

    // สร้าง payload ตามโครงสร้างที่ API ต้องการ
    const requestPayload = {
      table_id: parseInt(tableID),
      uuid: uuid,
      items: orderData,
      use_promo: promotionData,
    }

    console.log('Request Payload:', requestPayload)

    try {
      const response = await axios.post(
        'http://localhost:8080/api/orders',
        requestPayload
      )

      if (response.status === 200) {
        console.log('Order successfully placed!', response.data)
        useCartStore.getState().clearAllCart()
        setOpenCartModal(false)

        alert(
          language === 'th'
            ? 'สั่งอาหารสำเร็จ!'
            : language === 'en'
            ? 'Order placed successfully!'
            : '订单已成功提交！'
        )
      } else {
        console.error('Failed to place order', response.data)
        alert(
          language === 'th'
            ? 'ไม่สามารถสั่งอาหารได้: ' +
                (response.data.error || 'ข้อผิดพลาดที่ไม่ทราบสาเหตุ')
            : language === 'en'
            ? 'Failed to place order: ' +
              (response.data.error || 'Unknown error')
            : '下单失败：' + (response.data.error || '未知错误')
        )
      }
    } catch (error) {
      if (error.response) {
        console.error('Error placing order:', error.response.data)
        alert(
          language === 'th'
            ? 'ข้อผิดพลาด: ' +
                (error.response.data.error || 'ข้อผิดพลาดที่ไม่ทราบสาเหตุ')
            : language === 'en'
            ? 'Error: ' + (error.response.data.error || 'Unknown error')
            : '错误：' + (error.response.data.error || '未知错误')
        )
      } else if (error.request) {
        console.error('No response from server:', error.request)
        alert(
          language === 'th'
            ? 'ไม่ได้รับการตอบสนองจากเซิร์ฟเวอร์ กรุณาลองใหม่อีกครั้ง'
            : language === 'en'
            ? 'No response from server, please try again later.'
            : '服务器无响应，请稍后重试。'
        )
      } else {
        console.error('Unexpected error:', error.message)
        alert(
          language === 'th'
            ? 'เกิดข้อผิดพลาดที่ไม่คาดคิด กรุณาลองใหม่อีกครั้ง'
            : language === 'en'
            ? 'An unexpected error occurred. Please try again.'
            : '发生意外错误。请重试。'
        )
      }
    }
  }

  return (
    <div>
      <nav className="bg-[#1C2B41] shadow-sm fixed top-0 w-full z-50">
        <div className="max-w-screen-xl flex flex-wrap items-center justify-between mx-auto px-4 py-2">
          <div className="flex flex-col gap-1 justify-center items-start text-white">
            <span className="text-xl font-semibold whitespace-nowrap">
              EasyOrder
            </span>
            <span className="text-base font-semibold whitespace-nowrap">
              โต๊ะที่ {tableID}
            </span>
          </div>
          <div className="block w-auto">
            <ul className="font-medium flex gap-4">
              <li>
                {/* ปุ่มเลือกภาษา */}
                <button
                  type="button"
                  className="flex items-center bg-gray-200 border border-gray-300 rounded-md px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-300"
                  onClick={() =>
                    setLanguage((prev) =>
                      prev === 'th' ? 'en' : prev === 'en' ? 'ch' : 'th'
                    )
                  }
                >
                  <Languages className="mr-2" />
                  {language === 'th'
                    ? 'ไทย'
                    : language === 'en'
                    ? 'English'
                    : '中文'}
                </button>
              </li>
              <li>
                <button
                  type="button"
                  onClick={() => setOpenCallModal(true)}
                  className="block py-2 text-white"
                >
                  <Bell />
                </button>
                <Modal
                  show={openCallModal}
                  size="md"
                  onClose={() => setOpenCallModal(false)}
                  popup
                >
                  <Modal.Header className="bg-gray-200" />
                  <Modal.Body className="bg-gray-200 rounded-md">
                    <div className="flex justify-center items-center gap-4">
                      <button
                        type="button"
                        className="p-4 bg-white gap-1 hover:bg-blue-500 transition-colors shadow-sm rounded-md text-black hover:text-white flex flex-col items-center justify-center"
                      >
                        <User className="size-7" />
                        <span>Call Staff</span>
                      </button>
                      <button
                        type="button"
                        className="p-4 bg-white gap-1 hover:bg-blue-500 transition-colors shadow-sm rounded-md text-black hover:text-white flex flex-col items-center justify-center"
                      >
                        <Banknote className="size-7" />
                        <span>Payment</span>
                      </button>
                    </div>
                  </Modal.Body>
                </Modal>
              </li>
              <li>
                <button type="button" className="block py-2 text-white">
                  <SquareMenu />
                </button>
              </li>
              <li>
                <button
                  type="button"
                  onClick={() => setOpenCartModal((prev) => !prev)}
                  className="relative my-2 inline-flex items-center text-sm font-medium text-center text-white"
                >
                  <ShoppingCart />
                  <span className="sr-only">Cart</span>
                  <div className="absolute inline-flex items-center justify-center w-5 h-5 text-xs font-bold text-white bg-red-500 rounded-full -top-2 -end-2">
                    {totalQuantity}
                  </div>
                </button>
                {openCartModal && (
                  <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50 p-4">
                    <div className="bg-white px-2 pt-2 pb-4 border border-gray/50 rounded-xl w-full max-w-xl min-h-full max-h-screen flex flex-col">
                      <div className="flex top-0 justify-between items-center">
                        <h4 className="text-xl font-semibold text-black">
                          {language === 'th'
                            ? 'รายการอาหารที่สั่ง'
                            : language === 'en'
                            ? 'Order list'
                            : '訂購的食物清單'}
                        </h4>
                        <button
                          onClick={() => setOpenCartModal((prev) => !prev)}
                          className="text-white/70 hover:text-white z-10 bg-black/30 rounded-full p-2"
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
                      </div>
                      <hr className="mt-4" />
                      <div className="flex-1 overflow-y-auto py-2">
                        {cart.map((item, i) => (
                          <div
                            key={i}
                            className="flex justify-between items-center gap-2 p-2 border-b last:border-none"
                          >
                            <div className="flex flex-col gap-1">
                              <h6 className="text-lg font-medium">
                                {item.menuItem.ID}
                                {item.menuItem.Name}
                              </h6>
                              <span className="text-sm text-gray-600">
                                หมายเหตุ : {item.notes || '-'}
                              </span>
                              <div className="text-sm text-gray-500 mt-1">
                                {item.options.map((option, index) => (
                                  <div key={index}>
                                    <span>{option.name}</span> (+{option.price}
                                    ฿)
                                    <span className="text-xs text-gray-400">
                                      (ID: {option.menu_option_id})
                                    </span>
                                  </div>
                                ))}
                              </div>
                            </div>
                            <div className="flex flex-col items-center gap-1">
                              <span className="text-base font-medium">
                                {(item.menuItem.Price +
                                  item.options.reduce(
                                    (optionTotal, option) =>
                                      optionTotal + option.price,
                                    0
                                  )) *
                                  item.quantity}{' '}
                                บาท
                              </span>
                              <div className="flex justify-between items-center gap-3">
                                <button
                                  className="bg-gray-200 hover:bg-gray-300 disabled:opacity-70 disabled:hover:bg-gray-200 p-2 rounded-md text-black"
                                  onClick={() =>
                                    decreaseQuantity(
                                      item.menu_item_id,
                                      item.notes,
                                      item.options
                                    )
                                  }
                                >
                                  <Minus className="size-2" />
                                </button>
                                <span className="text-base font-semibold">
                                  {item.quantity}
                                </span>
                                <button
                                  className="bg-gray-200 hover:bg-gray-300 p-2 rounded-md text-black"
                                  onClick={() =>
                                    increaseQuantity(
                                      item.menu_item_id,
                                      item.notes,
                                      item.options
                                    )
                                  }
                                >
                                  <PlusIcon className="size-2" />
                                </button>
                              </div>
                            </div>
                          </div>
                        ))}

                        {/* แสดงโปรโมชัน */}
                        {promotions.map((promo, i) => (
                          <div
                            key={i}
                            className="flex justify-between items-center gap-2 p-2 border-b last:border-none"
                          >
                            <div className="flex flex-col gap-1">
                              <h6 className="text-lg font-medium">
                                {language === 'th'
                                  ? promo.promotion.Name
                                  : language === 'en'
                                  ? promo.promotion.NameEn
                                  : promo.promotion.NameCh}
                              </h6>
                              <span className="text-sm text-gray-600">
                                {language === 'th'
                                  ? promo.promotion.Description
                                  : language === 'en'
                                  ? promo.promotion.DescriptionEn
                                  : promo.promotion.DescriptionCh}
                              </span>
                              {promo.selectedOptions &&
                                promo.selectedOptions.length > 0 && (
                                  <div className="mt-2">
                                    <h6 className="font-medium text-gray-800">
                                      {language === 'th'
                                        ? 'ตัวเลือกที่เลือก:'
                                        : language === 'en'
                                        ? 'Selected options:'
                                        : '已选择的选项:'}
                                    </h6>
                                    <ul className="list-disc pl-5 text-sm text-gray-500">
                                      {promo.selectedOptions.map(
                                        (option, index) => (
                                          <li key={index}>
                                            {language === 'th'
                                              ? option.name
                                              : language === 'en'
                                              ? option.nameEn
                                              : option.nameCh}
                                            (+{option.price}{' '}
                                            {language === 'th'
                                              ? 'บาท'
                                              : language === 'en'
                                              ? 'THB'
                                              : '泰铢'}
                                            )
                                          </li>
                                        )
                                      )}
                                    </ul>
                                  </div>
                                )}
                            </div>
                            <div>
                              {/* แสดงราคาโปรโมชัน */}
                              {promo.promotion.Price * promo.quantity}{' '}
                              {language === 'th'
                                ? 'บาท'
                                : language === 'en'
                                ? 'THB'
                                : '泰铢'}
                              <div className="flex justify-between items-center gap-3">
                                <button
                                  className="bg-gray-200 hover:bg-gray-300 disabled:opacity-70 disabled:hover:bg-gray-200 p-2 rounded-md text-black"
                                  onClick={() =>
                                    decreaseQuantityPromo(
                                      promo.promotion_id,
                                      1,
                                      promo.selectedOptions
                                    )
                                  }
                                >
                                  <Minus className="size-2" />
                                </button>
                                <span className="text-sm text-gray-600">
                                  {promo.quantity}
                                </span>
                                <button
                                  className="bg-gray-200 hover:bg-gray-300 p-2 rounded-md text-black"
                                  onClick={() =>
                                    increaseQuantityPromo(
                                      promo.promotion_id,
                                      1,
                                      promo.selectedOptions
                                    )
                                  }
                                >
                                  <PlusIcon className="size-2" />
                                </button>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                      <div className="flex flex-col gap-4 px-2 pt-4 border-t border-gray-200">
                        <div className="flex justify-between items-center">
                          <span className="text-xl font-bold">
                            {language === 'th'
                              ? 'รวมทั้งหมด'
                              : language === 'en'
                              ? 'Total'
                              : '总计'}
                          </span>
                          <span className="text-xl font-bold">
                            {calculateTotal()}{' '}
                            {language === 'th'
                              ? 'บาท'
                              : language === 'en'
                              ? 'THB'
                              : '泰铢'}
                          </span>
                        </div>
                        <button
                          className="px-4 py-3 w-full rounded-md text-white font-semibold bg-[#4bcc37] hover:bg-[#4aac3b]"
                          onClick={confirmOrder}
                        >
                          {language === 'th'
                            ? 'ยืนยันการสั่งอาหาร'
                            : language === 'en'
                            ? 'Confirm Order'
                            : '确认订单'}
                        </button>
                      </div>
                    </div>
                  </div>
                )}
              </li>
            </ul>
          </div>
        </div>
      </nav>
      <MenuList language={language} />
    </div>
  )
}
