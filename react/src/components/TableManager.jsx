import React, { useState } from "react";
import axios from "axios";
import TableDetail from "./TableDetail";
import styles from "../styles/TableManager.module.css";

const initialTables = [
  { id: 1, number: 1, seats: 4, status: 'ว่าง', time: null, customerName: null, customerCount: null },
  { id: 2, number: 2, seats: 2, status: 'ไม่ว่าง', time: '18:30', customerName: 'คุณสมชาย', customerCount: 2 },
  { id: 3, number: 3, seats: 6, status: 'จอง', time: '19:00', customerName: 'คุณสมหญิง', customerCount: 5 },
  { id: 4, number: 4, seats: 4, status: 'ว่าง', time: null, customerName: null, customerCount: null },
  { id: 5, number: 5, seats: 8, status: 'ไม่ว่าง', time: '18:45', customerName: 'คุณสมศรี', customerCount: 7 },
  { id: 6, number: 6, seats: 4, status: 'ว่าง', time: null, customerName: null, customerCount: null }
];

const TableManager = () => {
  const [tables, setTables] = useState(initialTables);
  const [selectedTable, setSelectedTable] = useState(null);
  const [formData, setFormData] = useState({
    customerName: '',
    customerCount: '',
    time: '',
    status: '',
  });
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const handleTableAction = (table) => {
    setSelectedTable(table);
    setFormData({
      customerName: table.customerName || '',
      customerCount: table.customerCount || '',
      time: table.time || '',
      status: table.status,
    });
    setIsDialogOpen(true);
  };

  const handleFormSubmit = () => {
    const updatedTables = tables.map((table) =>
      table.id === selectedTable.id
        ? {
            ...table,
            customerName: formData.customerName,
            customerCount: parseInt(formData.customerCount, 10),
            time: formData.time,
            status: formData.status,
          }
        : table
    );
    setTables(updatedTables);
    setIsDialogOpen(false);
    clearForm();
  };

  const clearForm = () => {
    setFormData({
      customerName: '',
      customerCount: '',
      time: '',
      status: '',
    });
    setSelectedTable(null);
  };

  const handleClearTable = (tableId) => {
    const updatedTables = tables.map((table) =>
      table.id === tableId
        ? {
            ...table,
            status: 'ว่าง',
            customerName: null,
            customerCount: null,
            time: null,
          }
        : table
    );
    setTables(updatedTables);
  };

  return (
    <div>
      <TableDetail
        tables={tables}
        onTableAction={handleTableAction}
        onClearTable={handleClearTable}
      />
      {isDialogOpen && (
        <div className="fixed top-0 left-0 w-full h-full flex justify-center items-center"
          style={{
            backgroundColor: 'rgba(0, 0, 0, 0.5)',
          }}
        >
          <div className="bg-white p-10 border rounded-xl w-96 shadow-md">
            <h2>จัดการโต๊ะ</h2>
            <form onSubmit={(e) => {
              e.preventDefault();
              handleFormSubmit();
            }}>
              <div style={{ marginBottom: '12px' }}>
                <label>ชื่อลูกค้า:</label>
                <input className="w-full p-2 border border-gray-500 rounded"
                  type="text"
                  value={formData.customerName}
                  onChange={(e) => setFormData({ ...formData, customerName: e.target.value })}
                  
                />
              </div>
              <div className="mb-3" >
                <label>จำนวนลูกค้า:</label>
                <input className="w-full p-2 border border-gray-500 rounded"
                  type="number"
                  value={formData.customerCount}
                  onChange={(e) => setFormData({ ...formData, customerCount: e.target.value })}
                  
                />
              </div>
              <div style={{ marginBottom: '12px' }}>
                <label>เวลา:</label>
                <input className="w-full p-2 border border-gray-500 rounded"
                  type="time"
                  value={formData.time}
                  onChange={(e) => setFormData({ ...formData, time: e.target.value })}
                  
                />
              </div>
              <div style={{ marginBottom: '12px' }}>
                <label>สถานะ:</label>
                <select className="w-full p-4 border border-gray-500 rounded"
                  value={formData.status}
                  onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                  style={{ width: '100%', padding: '8px', marginBottom: '8px' }}
                >
                  <option value="ว่าง">ว่าง</option>
                  <option value="จอง">จอง</option>
                  <option value="ใช้งาน">ใช้งาน</option>
                </select>
              </div>
              <button className="bg-white mx-2 py-2 px-5  border border-gray-300 rounded-md hover:bg-slate-200" type="submit" >
                บันทึก
              </button>
              <button className="bg-red-500 mx-2 py-2 px-5 border rounded-md hover:bg-red-700 text-white"
                type="button"
                onClick={() => setIsDialogOpen(false)}
               
              >
                ยกเลิก
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default TableManager;
