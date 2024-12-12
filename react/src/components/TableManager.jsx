import React, { useState } from "react";
import TableDetail from "./TableDetail";

// ข้อมูลตัวอย่าง
const initialTables = [
  { id: 1, number: 1, seats: 4, status: "ว่าง", time: null, customerName: null, customerCount: null },
  { id: 2, number: 2, seats: 2, status: "ไม่ว่าง", time: "18:30", customerName: "คุณสมชาย", customerCount: 2 },
  { id: 3, number: 3, seats: 6, status: "จอง", time: "19:00", customerName: "คุณสมหญิง", customerCount: 5 },
  { id: 4, number: 4, seats: 4, status: "ว่าง", time: null, customerName: null, customerCount: null },
  { id: 5, number: 5, seats: 8, status: "ไม่ว่าง", time: "18:45", customerName: "คุณสมศรี", customerCount: 7 },
  { id: 6, number: 6, seats: 4, status: "ว่าง", time: null, customerName: null, customerCount: null },
  { id: 7, number: 7, seats: 4, status: "ว่าง", time: null, customerName: null, customerCount: null },
  { id: 8, number: 8, seats: 4, status: "ว่าง", time: null, customerName: null, customerCount: null },
  { id: 9, number: 9, seats: 4, status: "ว่าง", time: null, customerName: null, customerCount: null },
  { id: 10, number: 10, seats: 4, status: "ว่าง", time: null, customerName: null, customerCount: null },
  { id: 11, number: 11, seats: 4, status: "ว่าง", time: null, customerName: null, customerCount: null },
];

// การจัดการโต๊ะทั้งหมด
const TableManager = () => {
  const [tables, setTables] = useState(initialTables);
  const [searchQuery, setSearchQuery] = useState("");
  const [filterStatus, setFilterStatus] = useState("");
  const [selectedTable, setSelectedTable] = useState(null);
  const [tablesToCombine, setTablesToCombine] = useState([]); // สำหรับโต๊ะที่จะรวม
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isCombineDialogOpen, setIsCombineDialogOpen] = useState(false); // Dialog สำหรับรวมโต๊ะ

  const [formData, setFormData] = useState({
    customerName: "",
    customerCount: "",
    time: "",
    status: "",
  });

  
// ฟังก์ชันรวมโต๊ะ
  const handleCombineTables = () => {
    if (tablesToCombine.length > 1) {
      const mainTable = tablesToCombine[0]; // โต๊ะหลัก
      const combinedCustomerCount = tablesToCombine.reduce((sum, table) => sum + (table.customerCount || 0), 0);
      const combinedTime = mainTable.time; // ใช้เวลาของโต๊ะหลัก
      const customerName = mainTable.customerName || "รวมโต๊ะ";
  
      const updatedTables = tables.map((table) => {
        if (tablesToCombine.some((t) => t.id === table.id)) {
          return {
            ...table,
            status: `รวมอยู่กับโต๊ะ ${mainTable.number}`, // แสดงสถานะรวม
            customerCount: combinedCustomerCount,
            customerName: customerName,
            time: combinedTime,
          };
        }
        return table;
      });
  
      setTables(updatedTables);
      setIsCombineDialogOpen(false); // ปิด dialog
      setTablesToCombine([]); // รีเซ็ตโต๊ะที่จะรวม
    } else {
      alert("กรุณาเลือกโต๊ะอย่างน้อย 2 โต๊ะ");
    }
  };

  // เปิด dialog การรวมโต๊ะ
  const handleTableCombine = (table) => {
    setTablesToCombine([table]); // กำหนดโต๊ะหลัก
    setIsCombineDialogOpen(true); // เปิด dialog
  };

  // ฟังก์ชันค้นหาและกรองโต๊ะ
  const filteredTables = tables.filter((table) => {
    const matchesSearch = table.number.toString().includes(searchQuery);
    const matchesStatus = filterStatus ? table.status === filterStatus : true;
    return matchesSearch && matchesStatus;
  });

  const handleSearchChange = (e) => setSearchQuery(e.target.value);
  const handleFilterChange = (e) => setFilterStatus(e.target.value);

  // ฟังก์ชันการจัดการโต๊ะ
  const handleTableAction = (table) => {
    setSelectedTable(table);
    setFormData({
      customerName: table.customerName || "",
      customerCount: table.customerCount || "",
      time: table.time || "",
      status: table.status || "ว่าง", // กำหนดสถานะเริ่มต้นเป็น "ว่าง"
    });
    setIsDialogOpen(true);
  };
  
  // ฟังก์ชันการยีนยันสถานะของโต๊ะ
  const handleFormSubmit = () => {
    const updatedTables = tables.map((table) =>
      table.id === selectedTable.id
        ? {
            ...table,
            ...formData,
            time: formData.status === "ใช้งาน" ? new Date().toLocaleTimeString() : formData.time,
          }
        : table
    );
  
    setTables(updatedTables);
    setIsDialogOpen(false);
  };

  // ฟังก์ชันการเคลียร์โต๊ะ
  const handleClearTable = (tableId) => {
    const updatedTables = tables.map((table) =>
      table.id === tableId
        ? {
            ...table,
            status: "ว่าง",
            customerName: null,
            customerCount: null,
            time: null,
          }
        : table
    );
    setTables(updatedTables);
  };

  return (
    <div className="้h-screen overflow-auto p-4">
      {/* ส่วนค้นหาและกรอง */}
      <div className="flex flex-col md:flex-row md:items-center mb-4">
        <input
          type="text"
          placeholder="ค้นหาโต๊ะ..."
          value={searchQuery}
          onChange={handleSearchChange}
          className="border p-2 rounded w-full md:w-1/3 mb-2 md:mb-0 md:mr-4"
        />
        <select
          value={filterStatus}
          onChange={handleFilterChange}
          className="border p-2 rounded w-full md:w-1/4"
        >
          <option value="">สถานะทั้งหมด</option>
          <option value="ว่าง">ว่าง</option>
          <option value="จอง">จอง</option>
          <option value="ไม่ว่าง">ไม่ว่าง</option>
        </select>
      </div>

      {/* ตารางแสดงข้อมูลโต๊ะ */}
      <TableDetail
        tables={filteredTables}
        onTableAction={handleTableAction}
        onClearTable={handleClearTable}
        onTableCombine={handleTableCombine} // ส่งฟังก์ชันรวมโต๊ะ
      />

     {/* Dialog สำหรับจัดการโต๊ะ */}
{isDialogOpen && (
  <div
    className="fixed top-0 left-0 w-full h-full flex justify-center items-center"
    style={{ backgroundColor: "rgba(0, 0, 0, 0.5)" }}
  >
    <div className="bg-white p-10 border rounded-xl w-96 shadow-md">
      <h2 className="text-lg font-bold mb-4">จัดการโต๊ะ</h2>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          handleFormSubmit();
        }}
      >
        <div className="mb-4">
          <label>สถานะ:</label>
          <select
            className="w-full p-2 border rounded"
            value={formData.status}
            onChange={(e) => setFormData({ ...formData, status: e.target.value })}
          >
            <option value="ว่าง">ว่าง</option>
            <option value="จอง">จอง</option>
            <option value="ใช้งาน">ใช้งาน</option>
          </select>
        </div>

        {/* แสดงฟิลด์เวลาเฉพาะเมื่อสถานะคือ "จอง" */}
        {formData.status === "จอง" && (
          <div className="mb-4">
            <label>เวลาในการจอง:</label>
            <input
              type="time"
              className="w-full p-2 border rounded"
              value={formData.time}
              onChange={(e) => setFormData({ ...formData, time: e.target.value })}
            />
          </div>
        )}

        {/* แสดงเวลาเช็คอินเมื่อสถานะคือ "ใช้งาน" */}
        {formData.status === "ใช้งาน" && (
          <div className="mb-4">
            <label>เช็คอินเวลา:</label>
            <p className="p-2 border rounded bg-gray-100">
              {new Date().toLocaleTimeString()}
            </p>
          </div>
        )}
        {/*สิ้นสุด แสดงเวลาเช็คอินเมื่อสถานะคือ "ใช้งาน" */}

        <div className="mb-4">
          <label>ชื่อลูกค้า:</label>
          <input
            type="text"
            className="w-full p-2 border rounded"
            value={formData.customerName}
            onChange={(e) => setFormData({ ...formData, customerName: e.target.value })}
          />
        </div>
        <div className="mb-4">
          <label>จำนวนลูกค้า:</label>
          <input
            type="number"
            className="w-full p-2 border rounded"
            value={formData.customerCount}
            onChange={(e) => setFormData({ ...formData, customerCount: e.target.value })}
          />
        </div>

        <div className="flex justify-end space-x-2">
          <button
            type="submit"
            className={`py-2 px-4 rounded ${
              formData.status === "ว่าง"
                ? "bg-gray-400 cursor-not-allowed"
                : "bg-blue-500 hover:bg-blue-700 text-white"
            }`}
            disabled={formData.status === "ว่าง"}
          >
            บันทึก
          </button>
          <button
            type="button"
            className="bg-red-500 text-white py-2 px-4 rounded hover:bg-red-700"
            onClick={() => setIsDialogOpen(false)}
          >
            ยกเลิก
          </button>
        </div>
      </form>
    </div>
  </div>
)}

  {/* Dialog สำหรับรวมโต๊ะ */}
{isCombineDialogOpen && (
  <div className="fixed top-0 left-0 w-full h-full flex justify-center items-center" style={{ backgroundColor: "rgba(0, 0, 0, 0.5)" }}>
    <div className="bg-white p-10 border rounded-xl w-96 shadow-md">
      <h2 className="text-lg font-bold mb-4">รวมโต๊ะ</h2>
      <p>คุณต้องการรวมโต๊ะที่เลือกไว้ใช่หรือไม่?</p>

      <div className="mb-4">
        <label htmlFor="tablesToCombine" className="block font-medium mb-2">เลือกโต๊ะที่จะรวม:</label>
        <select
          id="tablesToCombine"
          multiple
          value={tablesToCombine.map((table) => table.id)}
          onChange={(e) => {
            const selectedTableIds = Array.from(e.target.selectedOptions, (option) => parseInt(option.value));
            const selectedTables = tables.filter((table) => selectedTableIds.includes(table.id));
            setTablesToCombine([...tablesToCombine.slice(0, 1), ...selectedTables]); // เก็บโต๊ะหลักไว้
          }}
          className="w-full p-2 border rounded"
        >
          {tables
            .filter((table) => table.id !== tablesToCombine[0]?.id) // ไม่แสดงโต๊ะหลัก
            .map((table) => (
              <option key={table.id} value={table.id}>
                โต๊ะ {table.number} - {table.status}
              </option>
            ))}
        </select>
      </div>
      
      <div className="flex justify-end space-x-2">
        <button
          className="bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-700"
          onClick={handleCombineTables}
        >
          ยืนยัน
        </button>
        <button
          className="bg-red-500 text-white py-2 px-4 rounded hover:bg-red-700"
          onClick={() => setIsCombineDialogOpen(false)}
        >
          ยกเลิก
        </button>
      </div>
    </div>
  </div>
)}

    </div>
  );
};

export default TableManager;
