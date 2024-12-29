import React, { useState } from "react";
import axios from "axios";
import styles from "../styles/TableDetail.module.css";

const TableInfo = ({ label, value }) => (
  <div className="flex justify-between mb-2">
    <span>{label}:</span>
    <span>{value}</span>
  </div>
);

const TableDetail = ({ tables = [], onTableAction, onClearTable }) => {
  const [showDetails, setShowDetails] = useState(null);
  const [message, setMessage] = useState("");
  const [showMergePopup, setShowMergePopup] = useState(false);
  const [showSplitPopup, setShowSplitPopup] = useState(false);  // สำหรับแสดง Popup การแยกโต๊ะ
  const [selectedTables, setSelectedTables] = useState([]);
  const [primaryTable, setPrimaryTable] = useState(null);
  const [groupId, setGroupId] = useState(null);  // เก็บ ID ของกลุ่มโต๊ะ

  // ฟังก์ชัน handleSplitTable
  const handleSplitTable = async () => {
    if (!groupId) {
      setMessage("ไม่พบข้อมูลกลุ่มโต๊ะที่ต้องการแยก");
      return;
    }

    try {
      // เรียก API split table
      const response = await axios.post(
        `http://localhost:8080/api/table/splitTable/${groupId}`,
        {},
        {
          headers: {
            "Content-Type": "application/json",
            accept: "application/json",
          },
        }
      );

      if (response.data.message === "Tables split successfully") {
        setMessage("โต๊ะถูกรแยกสำเร็จ");
        setShowSplitPopup(false);
        setGroupId(null);  // ล้างกลุ่มโต๊ะหลังจากแยก
      }
    } catch (error) {
      setMessage("เกิดข้อผิดพลาดในการแยกโต๊ะ");
    }
  };

  const handleShowDetails = (tableId) => {
    setShowDetails(showDetails === tableId ? null : tableId);
  };

  const handleClosePopup = () => {
    setShowMergePopup(false);
    setPrimaryTable(null);
    setSelectedTables([]);
    setShowSplitPopup(false);
  };

  const handleTableSelection = (table) => {
    setSelectedTables((prevSelectedTables) =>
      prevSelectedTables.some((selectedTable) => selectedTable.ID === table.ID)
        ? prevSelectedTables.filter((selectedTable) => selectedTable.ID !== table.ID)
        : [...prevSelectedTables, table]
    );
  };

  const mergeTables = async () => {
    try {
      const tableIds = [primaryTable.ID, ...selectedTables.map((table) => table.ID)];

      const response = await axios.post(
        "http://localhost:8080/api/table/mergeTable",
        { table_ids: tableIds },
        {
          headers: {
            "Content-Type": "application/json",
            accept: "application/json",
          },
        }
      );

      if (response.data.message === "Tables merged successfully") {
        setGroupId(response.data.group_id);  // เก็บ group_id ของโต๊ะที่รวม
        setMessage("โต๊ะถูกรวมสำเร็จ");
        setShowMergePopup(false);
      }
    } catch (error) {
      setMessage("เกิดข้อผิดพลาดในการรวมโต๊ะ");
    }
  };

  if (tables.length === 0) {
    return <div>ไม่มีข้อมูลโต๊ะ</div>;
  }

  return (
    <div className="p-4">
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {tables.map((table) => (
          <div
            key={table.ID}
            className={`${
              table.Status === "available"
                ? "bg-green-100"
                : table.Status === "reserved"
                ? "bg-yellow-100"
                : table.Status === "ไม่ว่าง"
                ? "bg-red-100"
                : "bg-blue-100"
            } border border-gray-300 rounded-lg p-4 shadow-md`}
          >
            <div className="flex justify-between items-center mb-2">
              <h3 className="m-0">โต๊ะ {table.Name}</h3>
              <div className="flex gap-2">
                {table.Status !== "available" && (
                  <button
                    onClick={() => handleSplitTable(table)}
                    className="px-2 py-2 border border-blue-500 rounded bg-white text-blue-500 hover:bg-blue-100"
                  >
                    แยกโต๊ะ
                  </button>
                )}
                <button
                  onClick={() => onClearTable(table)}
                  className="px-2 py-2 border border-blue-500 rounded bg-white text-blue-500 hover:bg-blue-100"
                >
                  จัดการ
                </button>
                {table.Status === "available" && (
                  <button
                    onClick={() => onTableAction("reserve", table)}
                    className="px-2 py-2 border border-orange-500 rounded bg-white text-orange-500 hover:bg-orange-100"
                  >
                    จองโต๊ะ
                  </button>
                )}
                {table.Status === "reserved" && (
                  <button
                    onClick={() => onTableAction("unreserve", table)}
                    className="px-2 py-2 border border-red-500 rounded bg-white text-red-500 hover:bg-red-100"
                  >
                    ยกเลิกการจอง
                  </button>
                )}
                {table.Status !== "available" && (
                  <button
                    onClick={() => handleShowDetails(table.ID)}
                    className="px-2 py-2 border border-red-500 rounded bg-white text-red-500 hover:bg-red-100"
                  >
                    ดูรายละเอียด
                  </button>
                )}
              </div>
            </div>
            <div>
              <TableInfo label="สถานะ" value={table.Status} />
              <TableInfo label="ที่นั่ง" value={`${table.Capacity} ที่นั่ง`} />
              {table.customerName && (
                <TableInfo label="ลูกค้า" value={table.customerName} />
              )}
              {table.customerCount && (
                <TableInfo label="จำนวน" value={`${table.customerCount} คน`} />
              )}
              {table.time && <TableInfo label="เวลา" value={table.time} />}
            </div>
          </div>
        ))}
      </div>
      {message && <div className="mt-4 text-center text-red-500">{message}</div>}
      {showMergePopup && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center">
          <div className="bg-white p-6 rounded-md shadow-lg w-96">
            <h2 className="text-xl font-semibold mb-4">เลือกโต๊ะที่จะรวม</h2>
            <p>เลือกโต๊ะที่ต้องการรวมกับ {primaryTable.Name}</p>
            <div className="mt-4">
              {tables
                .filter((table) => table.ID !== primaryTable.ID)
                .map((table) => (
                  <div key={table.ID} className="flex items-center mb-2">
                    <input
                      type="checkbox"
                      checked={selectedTables.some(
                        (selectedTable) => selectedTable.ID === table.ID
                      )}
                      onChange={() => handleTableSelection(table)}
                      className="mr-2"
                    />
                    <span>{table.Name}</span>
                  </div>
                ))}
            </div>
            <div className="flex justify-between mt-4">
              <button
                onClick={handleClosePopup}
                className="px-4 py-2 border border-gray-300 rounded bg-gray-100 text-gray-700"
              >
                ปิด
              </button>
              <button
                onClick={mergeTables}
                className="px-4 py-2 border border-green-500 rounded bg-green-500 text-white"
              >
                รวมโต๊ะ
              </button>
            </div>
          </div>
        </div>
      )}
      {showSplitPopup && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center">
          <div className="bg-white p-6 rounded-md shadow-lg w-96">
            <h2 className="text-xl font-semibold mb-4">ยืนยันการแยกโต๊ะ</h2>
            <p>คุณต้องการแยกกลุ่มโต๊ะนี้ใช่หรือไม่?</p>
            <div className="flex justify-between mt-4">
              <button
                onClick={handleClosePopup}
                className="px-4 py-2 border border-gray-300 rounded bg-gray-100 text-gray-700"
              >
                ปิด
              </button>
              <button
                onClick={handleSplitTable}
                className="px-4 py-2 border border-green-500 rounded bg-green-500 text-white"
              >
                แยกโต๊ะ
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default TableDetail;
