import React, { useState, useEffect } from 'react'
import axios from 'axios'
import { Check, Clock } from 'lucide-react'

const API_BASE_URL = 'http://127.0.0.1:8080/api/orders'

const OrderConfirmation = () => {
  const [orders, setOrders] = useState([])
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [isCancelModalOpen, setIsCancelModalOpen] = useState(false)
  const [selectedDish, setSelectedDish] = useState(null)
  const [selectedOrderId, setSelectedOrderId] = useState(null)
  const [selectedDishIndex, setSelectedDishIndex] = useState(null)
  const [selectedItemForCancel, setSelectedItemForCancel] = useState(null)

  // ฟังก์ชันจัดเรียงออเดอร์
  const sortOrders = (ordersToSort) => {
    const sortedOrders = [...ordersToSort]
    return sortedOrders.sort((a, b) => {
      if (a.status === 'pending' && b.status === 'confirmed') return -1
      if (a.status === 'confirmed' && b.status === 'pending') return 1
      return a.orderTime - b.orderTime
    })
  }

  // ฟังก์ชันที่ใช้ดึงข้อมูลจาก API ด้วย axios
 const fetchOrders = () => {
  const token = localStorage.getItem('token')
  axios
    .get(`${API_BASE_URL}/active`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((response) => {
      const validOrders = response.data.filter((order) => {
        // ตรวจสอบว่า items เป็นอาร์เรย์ที่ไม่ใช่ null หรือ undefined
        const hasPendingItem =
          Array.isArray(order.items) &&
          order.items.some((item) => item.status === 'pending')

        return (
          Array.isArray(order.items) &&
          (order.status !== 'pending' || hasPendingItem)
        )
      })

      // กรองเมนูที่มีสถานะ 'cancelled'
      const filteredOrders = validOrders.map((order) => ({
        ...order,
        items: Array.isArray(order.items)
          ? order.items.filter((item) => item.status !== 'cancelled')
          : [],
      }))

      setOrders(sortOrders(filteredOrders))
    })
    .catch((error) => {
      console.error('Error fetching orders:', error)
    })
}


  useEffect(() => {
    fetchOrders()
  }, [])

  // ฟังก์ชันยืนยันการเสิร์ฟอาหาร
  const handleConfirmDish = async (orderId, dishIndex) => {
    try {
      const token = localStorage.getItem('token')
      if (!token) {
        setError('กรุณาเข้าสู่ระบบใหม่')
        return
      }

      const orderItemId = orders.find((order) => order.id === orderId)?.items[dishIndex]?.id
      if (!orderItemId) {
        setError('ไม่พบรายการอาหารที่ต้องการยืนยัน')
        return
      }

      setIsModalOpen(false) 

      await axios.post(
        `${API_BASE_URL}/items/serve/${orderItemId}`,
        {}, 
        { headers: { Authorization: `Bearer ${token}` } }
      )

      fetchOrders()
    } catch (error) {
      setError('เกิดข้อผิดพลาดในการยืนยันการเสิร์ฟ')
      console.error('Error confirming dish:', error)
    }
  }

  // ฟังก์ชันยกเลิกรายการอาหาร
  const handleCancelDish = (ItemId, tableId, uuid) => {
    const token = localStorage.getItem('token')
    axios
      .post(`${API_BASE_URL}/items/cancel`, {
        id: tableId,
        items: [
          {
            order_item_id: ItemId,
            quantity: 1,
          },
        ],
        order_uuid: uuid,
      }, {
        headers: {
          Authorization: `Bearer ${token}`,
          accept: 'application/json',
          'Content-Type': 'application/json',
        },
      })
      .then(() => {
        // รีเฟรชข้อมูลหลังจากยกเลิก
        fetchOrders()
        setIsCancelModalOpen(false) // ปิด Modal หลังยกเลิก
      })
      .catch((error) => {
        console.error('Error canceling item:', error)
      })
  }

  return (
    <div className="min-h-screen bg-gray-100 p-4">
      <div className="max-w-md mx-auto space-y-4 ">
        <h1 className="text-2xl text-white font-bold text-center mb-6 bg-gray-900 py-4 rounded-lg">
          ยืนยันการเสิร์ฟอาหาร
        </h1>

        {orders.map((order) => (
          <div
            key={order.id}
            className={`bg-white rounded-lg shadow-md p-4 ${
              order.status === 'confirmed' ? 'opacity-50' : ''
            }`}
          >
            <div className="flex justify-between items-center mb-2">
              <div className="flex items-center space-x-2">
                <h2 className="text-lg font-semibold">
                  โต๊ะที่ {order.table_id}
                </h2>
                <div className="flex items-center text-gray-500 text-sm">
                  <Clock size={16} className="mr-1" />
                  {new Date(order.orderTime).toLocaleTimeString('th-TH', {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </div>
              </div>
            </div>

            <div className="space-y-2">
              {order.items &&
                order.items.map((item, index) => (
                  <div key={item.id || index}>
                    {item.status !== 'cancelled' && (
                      <div>
                        <div className="flex justify-between border-t">
                          <div></div>
                          <button
                            className="bg-red-500 text-white py-1 px-1 rounded-md mt-2"
                            onClick={() => {
                              setSelectedItemForCancel({
                                id: item.id,
                                tableId: order.table_id,
                                uuid: order.uuid,
                              })
                              setIsCancelModalOpen(true)
                            }}
                          >
                            ยกเลิก
                          </button>
                        </div>
                        <div className="flex justify-between items-center mb-2">
                          <div>
                            <div className="flex justify-between">
                              <span
                                className={`${
                                  item.status === 'served'
                                    ? 'line-through text-gray-500'
                                    : 'text-gray-800'
                                }`}
                              >
                                {item.menu_item.name}
                              </span>
                            </div>
                            <div>
                              <span className="text-gray-500 text-sm">
                                หมายเหตุ: {item.notes}
                              </span>
                            </div>
                          </div>
                          <div></div>
                          <span className="text-gray-500 text-sm">
                            ฿{item.price}
                          </span>
                        </div>
                      </div>
                    )}

                    {item.status !== 'served' &&
                      item.status !== 'cancelled' && (
                        <div>
                          <button
                            className="text-white bg-blue-500 border-blue-500 border-2 rounded-lg px-3 py-2 text-sm hover:bg-blue-600 w-full"
                            onClick={() => {
                              setIsModalOpen(true)
                              setSelectedDish(item.menu_item.name)
                              setSelectedOrderId(order.id)
                              setSelectedDishIndex(index)
                            }}
                          >
                            ยืนยันการเสิร์ฟ
                          </button>
                        </div>
                      )}

                    {item.status === 'served' && (
                      <Check className="text-green-600" size={20} />
                    )}
                  </div>
                ))}
            </div>

            {order.status === 'served' && (
              <div className="text-green-600 text-sm mt-2 text-center">
                เสิร์ฟครบทุกเมนูแล้ว
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Modal สำหรับการยืนยันการเสิร์ฟ */}
      {isModalOpen && (
        <div
          className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center"
          onClick={() => setIsModalOpen(false)}
        >
          <div
            className="bg-white p-6 rounded-lg w-2/3"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-xl font-semibold mb-4">
              ยืนยันการเสิร์ฟ {selectedDish}
            </h3>
            <p>
              คุณแน่ใจหรือไม่ว่าได้เสิร์ฟ{' '}
              <h3 className="text-red-400">{selectedDish}</h3>ครบถ้วน?
            </p>
            <div className="mt-4 flex justify-end space-x-4">
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-gray-600 border-gray-600 border-2 rounded-sm px-4 py-2 hover:bg-gray-100"
              >
                ยกเลิก
              </button>
              <button
                onClick={() =>
                  handleConfirmDish(selectedOrderId, selectedDishIndex)
                }
                className="text-white bg-green-600 rounded-sm px-4 py-2 hover:bg-green-700"
              >
                ยืนยัน
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal สำหรับการยกเลิก */}
      {isCancelModalOpen && (
        <div
          className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center"
          onClick={() => setIsCancelModalOpen(false)}
        >
          <div
            className="bg-white p-6 rounded-lg w-2/3"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-xl font-semibold mb-4">ยืนยันการยกเลิก</h3>
            <p>คุณแน่ใจหรือไม่ที่จะยกเลิกรายการอาหารนี้?</p>
            <div className="mt-4 flex justify-end space-x-4">
              <button
                onClick={() => setIsCancelModalOpen(false)}
                className="text-gray-600 border-gray-600 border-2 rounded-sm px-4 py-2 hover:bg-gray-100"
              >
                ยกเลิก
              </button>
              <button
                onClick={() =>
                  handleCancelDish(
                    selectedItemForCancel.id,
                    selectedItemForCancel.tableId,
                    selectedItemForCancel.uuid
                  )
                }
                className="text-white bg-red-600 rounded-sm px-4 py-2 hover:bg-red-700"
              >
                ยืนยัน
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default OrderConfirmation
