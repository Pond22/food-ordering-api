import react, { useState, useEffect, useRef } from 'react'
import {
  Bell,
  X,
  UserRound,
  CreditCard,
} from 'lucide-react'

const Notify = () => {
  const [showNotifications, setShowNotifications] = useState(false)
  const [notifications, setNotifications] = useState([
    {
      id: 1,
      type: 'service',
      table: 'A12',
      message: 'เรียกพนักงาน',
      time: '2 นาทีที่แล้ว',
      read: false,
    },
    {
      id: 2,
      type: 'bill',
      table: 'B05',
      message: 'ขอเช็คบิล',
      time: '5 นาทีที่แล้ว',
      read: false,
    },
  ])
  const handleDismiss = (id) => {
    setNotifications(notifications.filter((note) => note.id !== id))
  }

  const markAllAsRead = () => {
    setNotifications(notifications.map((note) => ({ ...note, read: true })))
  }

  return (
    <div className="fixed bottom-4 right-4 z-50">
      {/* Notification Icon */}
      <button
        onClick={() => setShowNotifications(!showNotifications)}
        className="relative bg-white p-3 rounded-full shadow-lg hover:bg-gray-50"
      >
        <Bell className="w-6 h-6 text-gray-700" />
        {notifications.some((note) => !note.read) && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs w-5 h-5 flex items-center justify-center rounded-full">
            {notifications.filter((note) => !note.read).length}
          </span>
        )}
      </button>

      {/* Notification Popup */}
      {showNotifications && (
        <div className="absolute bottom-16 right-0 w-80 bg-white rounded-lg shadow-xl border overflow-hidden">
          <div className="p-4 border-b bg-gray-50 flex justify-between items-center">
            <h3 className="font-semibold text-gray-700">การแจ้งเตือน</h3>
            <button
              onClick={markAllAsRead}
              className="text-sm text-blue-600 hover:text-blue-800"
            >
              อ่านทั้งหมด
            </button>
          </div>

          <div className="max-h-96 overflow-y-auto">
            {notifications.length === 0 ? (
              <div className="p-4 text-center text-gray-500">
                ไม่มีการแจ้งเตือน
              </div>
            ) : (
              <div className="divide-y">
                {notifications.map((notification) => (
                  <div
                    key={notification.id}
                    className={`p-4 hover:bg-gray-50 flex items-start justify-between ${
                      !notification.read ? 'bg-blue-50' : ''
                    }`}
                  >
                    <div className="flex items-start space-x-3">
                      <div
                        className={`rounded-full p-2 ${
                          notification.type === 'service'
                            ? 'bg-blue-100 text-blue-600'
                            : 'bg-green-100 text-green-600'
                        }`}
                      >
                        {notification.type === 'service' ? (
                          <UserRound className="w-5 h-5" />
                        ) : (
                          <CreditCard className="w-5 h-5" />
                        )}
                      </div>
                      <div>
                        <p className="font-medium text-gray-800">
                          โต๊ะ {notification.table}
                        </p>
                        <p className="text-sm text-gray-600">
                          {notification.message}
                        </p>
                        <p className="text-xs text-gray-400 mt-1">
                          {notification.time}
                        </p>
                      </div>
                    </div>
                    <button
                      onClick={() => handleDismiss(notification.id)}
                      className="text-gray-400 hover:text-gray-600"
                    >
                      <X className="w-5 h-5" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="p-3 bg-gray-50 border-t">
            <button
              onClick={() => setShowNotifications(false)}
              className="w-full py-2 text-sm text-center text-gray-600 hover:text-gray-800"
            >
              ปิด
            </button>
          </div>
        </div>
      )}
    </div>
  )
};
export default Notify;