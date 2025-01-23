import React, { useState, useEffect } from 'react'
import axios from 'axios'
import styles from '../styles/TableDetail.module.css'
import { X, CalendarDays, Edit, Trash2, Plus, Split } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import PaymentTables from './PaymentTables'

// Helper component for displaying table info
const TableInfo = ({ label, value }) => (
  <div className="flex justify-between mb-2">
    <span>{label}:</span>
    <span>{value}</span>
  </div>
)

const TableManager = () => {
  const [tables, setTables] = useState([])
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [newTableName, setNewTableName] = useState('')
  const [newTableCapacity, setNewTableCapacity] = useState(0)
  const [newTableStatus, setNewTableStatus] = useState('ว่าง')
  const [message, setMessage] = useState('')
  const [isNameConflict, setIsNameConflict] = useState(false)
  const [selectedTables, setSelectedTables] = useState([])
  const [isMergeDialogOpen, setIsMergeDialogOpen] = useState(false)
  const [primaryTable, setPrimaryTable] = useState(null)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [tableToDelete, setTableToDelete] = useState(null)
  const [showSplitPopup, setShowSplitPopup] = useState(false) // สำหรับแสดง Popup การแยกโต๊ะ

  // Reservation Dialog States
  const [isReservationDialogOpen, setIsReservationDialogOpen] = useState(false)
  const [customerName, setCustomerName] = useState('')
  const [customerCount, setCustomerCount] = useState(0)
  const [reservationTime, setReservationTime] = useState('')
  const [selectedTable, setSelectedTable] = useState(null)
  const [selectedGroupId, setSelectedGroupId] = useState(null)

   const navigate = useNavigate()
   const handleCheckBillClick = () => {
     // เมื่อกดปุ่มจะนำทางไปยังหน้า PaymentTables
     navigate('/payment-tables')
   }

  // Fetch tables from API and sort by ID
  useEffect(() => {
    const fetchTables = async () => {
      try {
        const response = await axios.get('http://localhost:8080/api/tables')
        // Sorting tables by ID in ascending order
        const sortedTables = response.data.sort((a, b) => a.ID - b.ID)
        setTables(sortedTables) // Store sorted tables into state
      } catch (error) {
        console.error('Error fetching tables:', error)
      }
    }

    fetchTables() // Fetch tables when the component mounts

    // Optionally, you can reconnect the WebSocket here to listen for table updates
    const connectWebSocket = () => {
      try {
        const socket = new WebSocket('ws://localhost:8080/ws/tables')

        socket.onopen = () => {
          console.log('WebSocket connected successfully!')
        }

        socket.onerror = (error) => {
          console.error('WebSocket error:', error)
        }

        socket.onmessage = (event) => {
          const message = JSON.parse(event.data)
          if (message.type === 'table_update') {
            setTables(message.data)
          }
        }

        socket.onclose = () => {
          console.log('WebSocket connection closed. Attempting reconnect...')
          setTimeout(connectWebSocket, 5000)
        }

        return socket
      } catch (error) {
        console.error('Failed to connect to WebSocket:', error)
        setTimeout(connectWebSocket, 5000)
      }
    }

    const socket = connectWebSocket()

    return () => {
      socket.close()
    }
  }, []) // Empty dependency array to run only on mount

  // Check if table name is duplicated
  const checkTableNameConflict = async () => {
    try {
      const response = await axios.get('http://localhost:8080/api/tables')
      const data = response.data
      return data.some((table) => table.Name === newTableName)
    } catch (error) {
      console.error('Error fetching tables:', error)
      return false
    }
  }

  // Table rendering
  const renderTableDetails = (table) => {
    // ตรวจสอบว่าโต๊ะมี GroupID หรือไม่
    if (table.GroupID) {
      // หากมี GroupID, แสดงรายละเอียดเกี่ยวกับโต๊ะในกลุ่ม
      const groupedTables = tables.filter((t) => t.GroupID === table.GroupID) // โต๊ะในกลุ่มเดียวกัน
      return (
        <div className="mt-4">
          <h4 className="text-base text-black/50 font-semibold">
            รหัสกลุ่ม : {table.GroupID}
          </h4>
          {groupedTables.map((groupedTable) => (
            <TableInfo
              key={groupedTable.ID}
              label={`โต๊ะ ${groupedTable.Name}`}
              value={`ความจุ: ${groupedTable.Capacity}`}
            />
          ))}
        </div>
      )
    }
    return null // ถ้าไม่มี GroupID จะไม่แสดงอะไร
  }

  const handleTableAction = async (action, table) => {
    if (!table || !table.ID) {
      alert('ข้อมูลโต๊ะไม่ถูกต้อง')
      return
    }

    const endpoint =
      action === 'reserve'
        ? `http://localhost:8080/api/table/reservedTable/${table.ID}`
        : action === 'unreserve'
        ? `http://localhost:8080/api/table/unreservedTable/${table.ID}`
        : null

    if (!endpoint) {
      console.warn('Unsupported action:', action)
      return
    }

    try {
      const response = await axios.post(endpoint, {})
      alert(response.data.message || 'ดำเนินการสำเร็จ')
      setTables((prevTables) =>
        prevTables.map((t) =>
          t.ID === table.ID
            ? { ...t, Status: action === 'reserve' ? 'reserved' : 'available' }
            : t
        )
      )
    } catch (error) {
      console.error(`Error processing ${action}:`, error)
      alert('เกิดข้อผิดพลาดในการดำเนินการ')
    }
  }

  const handleReservationSubmit = () => {
    const currentTime = new Date()
    const reservationDate = new Date(reservationTime)

    if (reservationDate < currentTime) {
      alert('เวลาจองต้องไม่ใช่เวลาในอดีต')
      return
    }

    if (selectedTable) {
      const updatedTables = tables.map((table) =>
        table.ID === selectedTable.ID
          ? {
              ...table,
              Status: 'reserved',
              customerName,
              customerCount,
              reservationTime,
            }
          : table
      )

      setTables(updatedTables)
      setIsReservationDialogOpen(false)
      setCustomerName('')
      setCustomerCount(0)
      setReservationTime('')
    }
  }

  // Create a new table
  const createTable = async () => {
    const isConflict = await checkTableNameConflict()
    if (isConflict) {
      setIsNameConflict(true)
      setMessage('ชื่อโต๊ะซ้ำกับโต๊ะที่มีอยู่แล้ว')
      return
    }

    const tableData = {
      capacity: parseInt(newTableCapacity),
      name: newTableName,
      status: newTableStatus,
    }

    try {
      const response = await axios.post(
        'http://localhost:8080/api/table',
        tableData,
        {
          headers: {
            'Content-Type': 'application/json',
            accept: 'application/json',
          },
        }
      )

      const data = response.data

      if (response.status === 200) {
        setMessage(`โต๊ะ "${data.Name}" ถูกสร้างเรียบร้อยแล้ว!`)
        setTables((prevTables) => [...prevTables, data])
        setIsDialogOpen(false)
      } else {
        setMessage(`เกิดข้อผิดพลาด: ${data.error || 'ข้อมูลไม่ถูกต้อง'}`)
      }
    } catch (error) {
      setMessage('เกิดข้อผิดพลาดในการเชื่อมต่อ API')
      console.error(error)
    }
  }

  const handleDeleteTable = async () => {
    if (tableToDelete && tableToDelete.ID) {
      try {
        const response = await axios.delete(
          `http://localhost:8080/api/table/${tableToDelete.ID}`
        )
        if (response.status === 200) {
          alert('โต๊ะถูกลบเรียบร้อยแล้ว!')
          setTables((prevTables) =>
            prevTables.filter((table) => table.ID !== tableToDelete.ID)
          )
        } else {
          alert('ไม่สามารถลบโต๊ะได้')
        }
      } catch (error) {
        console.error('Error deleting table:', error)
        alert('เกิดข้อผิดพลาดในการลบโต๊ะ')
      }
    }
    setIsDeleteDialogOpen(false) // Close the dialog after deletion
  }

  const handlePrimaryTableSelection = (table) => {
    setPrimaryTable(table) // เมื่อผู้ใช้เลือกโต๊ะหลัก
  }

  const handleSelectTable = (tableId) => {
    setSelectedTables((prevSelected) =>
      prevSelected.includes(tableId)
        ? prevSelected.filter((id) => id !== tableId)
        : [...prevSelected, tableId]
    )
  }

  const mergeTables = async () => {
    if (selectedTables.length < 2) {
      alert('กรุณาเลือกโต๊ะอย่างน้อย 2 โต๊ะ')
      return
    }

    try {
      const response = await axios.post(
        'http://localhost:8080/api/table/mergeTable',
        {
          table_ids: selectedTables,
        }
      )

      if (response.status === 200) {
        const { group_id, main_table_id, message } = response.data
        alert(message || 'รวมโต๊ะสำเร็จ')

        // อัปเดตสถานะโต๊ะที่รวมแล้ว และเก็บค่า group_id
        setTables((prevTables) =>
          prevTables.map((table) =>
            selectedTables.includes(table.ID)
              ? {
                  ...table,
                  Status: 'merged',
                  GroupID: group_id,
                  MainTableID: main_table_id,
                }
              : table
          )
        )

        // เรียงโต๊ะตาม ID จากน้อยไปมาก
        setTables((prevTables) => {
          const sortedTables = prevTables.sort((a, b) => a.ID - b.ID)
          return sortedTables
        })

        // เคลียร์โต๊ะที่เลือก
        setSelectedTables([])
      } else {
        alert('ไม่สามารถรวมโต๊ะได้')
      }
    } catch (error) {
      console.error('Error merging tables:', error)
      alert('เกิดข้อผิดพลาดในการรวมโต๊ะ')
    }
  }

  const handleOpenPopupSpilt = (groupId) => {
    setSelectedGroupId(groupId)
    setShowSplitPopup(true)
  }
  const handleClosePopupSpilt = () => {
    setShowSplitPopup(false)
    setSelectedGroupId(null) // Reset selectedGroupId when closing popup
  }

  const handleSplitTable = async (groupId) => {
    try {
      const response = await axios.post(
        `http://localhost:8080/api/table/splitTable/${groupId}`
      )

      if (response.status === 200) {
        alert(response.data.message || 'แยกโต๊ะสำเร็จ')

        // อัปเดตสถานะของโต๊ะกลับไปเป็นสถานะเดิม
        setTables((prevTables) =>
          prevTables.map((table) =>
            table.GroupID === groupId
              ? {
                  ...table,
                  Status: 'available',
                  GroupID: null,
                  MainTableID: null,
                }
              : table
          )
        )
        // ปิด Popup หลังจากแยกโต๊ะสำเร็จ
        handleClosePopupSpilt()
      } else {
        alert('ไม่สามารถแยกโต๊ะได้')
      }
    } catch (error) {
      console.error('เกิดข้อผิดพลาดในการแยกโต๊ะ:', error)
      alert('เกิดข้อผิดพลาดในการแยกโต๊ะ')
    }
  }

  // ฟังก์ชันเปิด/ปิดโต๊ะ
  const handleToggleTable = async (table, isOpen) => {
    if (!table || !table.ID) {
      alert('ข้อมูลโต๊ะไม่ถูกต้อง')
      return
    }

    const endpoint = `http://localhost:8080/api/table/setstatus/${table.ID}`

    try {
      // ใช้ PUT request แทน POST
      const response = await axios.put(endpoint, {})

      // แสดงข้อความจาก API หรือข้อความที่กำหนดเอง
      alert(
        response.data.message || (isOpen ? 'เปิดโต๊ะสำเร็จ' : 'ปิดโต๊ะสำเร็จ')
      )

      // อัปเดตสถานะของโต๊ะใน state
      setTables((prevTables) =>
        prevTables.map((t) =>
          t.ID === table.ID
            ? { ...t, Status: isOpen ? 'available' : 'unavailable' }
            : t
        )
      )
    } catch (error) {
      console.error('Error toggling table:', error)
      alert('เกิดข้อผิดพลาดในการดำเนินการ')
    }
  }
  // Table rendering
  const renderTableActionButtons = (table) => {
    const status = table.Status || 'available'
    const tableStyles =
      status === 'reserved'
        ? styles.reservedTable
        : status === 'merged'
        ? styles.mergedTable
        : styles.availableTable

    return (
      <div className={tableStyles}>
        <div className="flex ">
          {status === 'available' && (
            <>
              <div className="flex">
                <div className="relative group">
                  <button
                    className="bg-white text-black p-2 border rounded-full hover:bg-gray-100"
                    onClick={() => handleTableAction('reserve', table)}
                  >
                    <CalendarDays></CalendarDays>
                  </button>
                  <div className="absolute left-1/2 -translate-x-1/2 w-auto bottom-full mb-2 hidden group-hover:block bg-gray-400 text-white text-sm py-1 px-2 rounded-md shadow-lg">
                    จอง
                  </div>
                </div>

                <div className="relative group">
                  <button
                    className="bg-white border text-black p-2 rounded-full hover:bg-gray-100"
                    onClick={() => handleToggleTable(table, true)} // เปิดโต๊ะ
                  >
                    <Plus></Plus>
                  </button>
                  <div className="absolute left-1/2 -translate-x-1/2 w-16 bottom-full mb-2 hidden group-hover:block bg-gray-400 text-white text-sm py-1 px-2 rounded-md shadow-lg">
                    เปิดโต๊ะ
                  </div>
                </div>

                <button
                  className={styles.trash}
                  onClick={() => {
                    setTableToDelete(table)
                    setIsDeleteDialogOpen(true) // เปิด dialog ลบโต๊ะ
                  }}
                >
                  <Trash2 className="size-6"></Trash2>
                </button>
              </div>
            </>
          )}
          {status === 'reserved' && (
            <>
              <button
                className="bg-gray-500 text-white px-4 py-2 rounded"
                onClick={() => handleTableAction('unreserve', table)}
              >
                ยกเลิกการจอง
              </button>
            </>
          )}
          {status === 'unavailable' && (
            <>
              <button
                className="bg-yellow-500 text-white px-4 py-2 rounded"
                onClick={() => handleToggleTable(table, false)} // ปิดโต๊ะ
              >
                ปิดโต๊ะ
              </button>
              {/* <button
              onClick={()=><PaymentTables></PaymentTables>}>เช็คบิล</button> */}
            </>
          )}
          {status === 'occupied' && (
            <>
              <button onClick={handleCheckBillClick}>เช็คบิล</button>
            </>
          )}
          {table.GroupID && (
            <div className="relative group">
              <button
                className="bg-orange-300 p-2 rounded-full text-white"
                onClick={() => handleOpenPopupSpilt(table.GroupID)}
              >
                <Split />
                <div className="absolute left-1/2 -translate-x-1/2 w-16 bottom-full mb-2 hidden group-hover:block bg-gray-400 text-white text-sm py-1 px-2 rounded-md shadow-lg">
                  แยกโต๊ะ
                </div>
              </button>
            </div>
          )}
        </div>
      </div>
    )
  }
  return (
    <div className="h-screen overflow-auto px-4 py-2 ">
      <div className="py-5  bg-gray-800 p-4 mb-2 rounded">
        <h1 className="text-white">จัดการโต๊ะและชำระเงิน</h1>
      </div>
      <button
        onClick={() => setIsDialogOpen(true)}
        className={styles.btnAction}
      >
        สร้างโต๊ะใหม่
      </button>

      <button
        onClick={() => setIsMergeDialogOpen(true)}
        className={styles.btnAction}
      >
        รวมกลุ่มโต๊ะ
      </button>
      {/* Render tables */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {tables
          .sort((a, b) => a.ID - b.ID) // จัดเรียงโต๊ะตาม ID (ascending)
          .map((table) => (
            <div
              key={table.ID}
              className={`border p-4 rounded-lg shadow-lg hover:bg-gray-100${
                table.GroupID ? 'border-orange-500 bg-orange-100' : ''
              } 
              ${
                table.Status === 'unavailable'
                  ? 'bg-red-200 border-orange-500'
                  : ''
              }
              ${
                table.Status === 'reserved'
                  ? 'bg-yellow-100 border-orange-500'
                  : ''
              }`}
            >
              <div className="flex justify-between">
                <h3 className="text-xl font-bold mb-2">{table.Name}</h3>
                {renderTableActionButtons(table)}
              </div>
              {/* Additional table actions can go here */}
              <TableInfo label="Capacity" value={table.Capacity} />
              <TableInfo label="Status" value={table.Status} />
              {/* แสดงรายละเอียดของกลุ่มโต๊ะหากโต๊ะมี GroupID */}
              {renderTableDetails(table)}
            </div>
          ))}
      </div>

      {/* Reservation Dialog */}
      {isReservationDialogOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-lg w-96">
            <h3 className="text-xl mb-4">จองโต๊ะ</h3>
            <input
              type="text"
              value={customerName}
              onChange={(e) => setCustomerName(e.target.value)}
              className="border p-2 mb-4 w-full"
              placeholder="ชื่อผู้จอง"
            />
            <input
              type="number"
              value={customerCount}
              onChange={(e) => setCustomerCount(e.target.value)}
              className="border p-2 mb-4 w-full"
              placeholder="จำนวนคน"
            />
            <input
              type="datetime-local"
              value={reservationTime}
              onChange={(e) => setReservationTime(e.target.value)}
              className="border p-2 mb-4 w-full"
            />
            <div className="flex justify-between">
              <button
                onClick={() => handleTableAction('reserve', table)}
                className="bg-green-500 text-white px-4 py-2 rounded"
              >
                ยืนยันการจอง
              </button>
              <button
                onClick={() => setIsReservationDialogOpen(false)}
                className="bg-red-500 text-white px-4 py-2 rounded"
              >
                ยกเลิก
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Merge Dialog */}
      {isMergeDialogOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-lg w-96">
            <h3 className="text-xl font-bold mb-4">เลือกโต๊ะสำหรับรวม</h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4 p-4">
              {tables.map((table) => (
                <div
                  key={table.ID}
                  onClick={() => handleSelectTable(table.ID)} // คลิกที่ div เพื่อเลือก
                  className={`flex items-center justify-between p-4 rounded-lg shadow-lg cursor-pointer ${
                    selectedTables.includes(table.ID)
                      ? 'bg-blue-100'
                      : 'bg-white'
                  } hover:bg-blue-50 transition-all duration-300 ease-in-out`}
                >
                  <label className="text-lg font-medium text-gray-800">
                    {table.Name}
                  </label>
                </div>
              ))}
            </div>

            <div className="flex justify-between">
              <button
                onClick={() => {
                  mergeTables()
                  setSelectedTables([]) // เคลียร์การเลือกโต๊ะหลังจากรวมโต๊ะ
                  setIsMergeDialogOpen(false) // ปิด dialog
                }}
                className="bg-green-500 text-white px-4 py-2 rounded"
              >
                รวมโต๊ะ
              </button>
              <button
                onClick={() => {
                  setSelectedTables([]) // เคลียร์การเลือกโต๊ะเมื่อปิด dialog
                  setIsMergeDialogOpen(false) // ปิด dialog
                }}
                className="bg-red-500 text-white px-4 py-2 rounded"
              >
                ยกเลิก
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ถ้าคุณเลือกโต๊ะที่รวมแล้ว แสดงปุ่มยกเลิกการรวม */}
      {selectedTables.length > 0 && selectedTables[0].GroupID && (
        <button
          onClick={() => cancelMerge(selectedTables[0].GroupID)}
          className="bg-red-500 text-white px-4 py-2 rounded mt-4"
        >
          ยกเลิกการรวมโต๊ะ
        </button>
      )}

      {/* New Table Dialog */}
      {isDialogOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center">
          <div className="bg-white p-6 rounded-lg shadow-lg w-96">
            <h3 className="text-xl mb-4">สร้างโต๊ะใหม่</h3>
            <label>ชื่อโต๊ะ</label>
            <input
              type="text"
              value={newTableName}
              onChange={(e) => setNewTableName(e.target.value)}
              className="border p-2 mb-4 w-full"
              placeholder="ชื่อโต๊ะ"
            />
            <label>จำนวนที่นั้ง</label>
            <input
              type="number"
              value={newTableCapacity}
              onChange={(e) => setNewTableCapacity(e.target.value)}
              className="border p-2 mb-4 w-full"
              placeholder="ความจุ"
              onWheel={(e) => e.preventDefault()} // ป้องกันการเลื่อนเมาส์
            />
            {/* <select
              value={newTableStatus}
              onChange={(e) => setNewTableStatus(e.target.value)}
              className="border p-2 mb-4 w-full"
            >
              <option value="ว่าง">ว่าง</option>
              <option value="จอง">จอง</option>
              <option value="ไม่ว่าง">ไม่ว่าง</option>
            </select> */}
            {isNameConflict && (
              <div className="text-red-500 text-sm mb-4">{message}</div>
            )}
            <div className="flex justify-end">
              <button
                onClick={() => setIsDialogOpen(false)}
                className="bg-red-500 text-white px-4 py-2 rounded"
              >
                ยกเลิก
              </button>
              <button onClick={createTable} className={styles.btnAction}>
                สร้างโต๊ะ
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Table Confirmation Dialog */}
      {isDeleteDialogOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-lg w-96">
            <h3 className="text-xl mb-4">ยืนยันการลบโต๊ะ</h3>
            <p>คุณแน่ใจหรือไม่ว่าต้องการลบโต๊ะ "{tableToDelete.Name}"?</p>
            <div className="flex justify-between mt-4">
              <button
                onClick={handleDeleteTable}
                className="bg-red-500 text-white px-4 py-2 rounded"
              >
                ยืนยัน
              </button>
              <button
                onClick={() => setIsDeleteDialogOpen(false)}
                className="bg-gray-500 text-white px-4 py-2 rounded"
              >
                ยกเลิก
              </button>
            </div>
          </div>
        </div>
      )}

      {showSplitPopup && selectedGroupId && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center">
          <div className="bg-white p-6 rounded-md shadow-lg w-96">
            <h2 className="text-xl font-semibold mb-4">ยืนยันการแยกโต๊ะ</h2>
            <p>คุณต้องการแยกกลุ่มโต๊ะนี้ใช่หรือไม่?</p>
            <div className="flex justify-between mt-4">
              <button
                onClick={handleClosePopupSpilt}
                className="px-4 py-2 border border-gray-300 rounded bg-gray-100 text-gray-700"
              >
                ปิด
              </button>
              <button
                onClick={() => handleSplitTable(selectedGroupId)}
                className="px-4 py-2 border border-green-500 rounded bg-green-500 text-white"
              >
                แยกโต๊ะ
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default TableManager
