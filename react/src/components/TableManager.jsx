import React, { useState, useEffect } from "react";
import axios from "axios";
import styles from "../styles/TableDetail.module.css";

// Helper component for displaying table info
const TableInfo = ({ label, value }) => (
  <div className="flex justify-between mb-2">
    <span>{label}:</span>
    <span>{value}</span>
  </div>
);

const TableManager = () => {
  const [tables, setTables] = useState([]);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newTableName, setNewTableName] = useState("");
  const [newTableCapacity, setNewTableCapacity] = useState(0);
  const [newTableStatus, setNewTableStatus] = useState("ว่าง");
  const [message, setMessage] = useState("");
  const [isNameConflict, setIsNameConflict] = useState(false);
  const [selectedTables, setSelectedTables] = useState([]);
  const [isMergeDialogOpen, setIsMergeDialogOpen] = useState(false);
  const [primaryTable, setPrimaryTable] = useState(null);

 // Fetch tables from API and sort by ID
useEffect(() => {
  const fetchTables = async () => {
    try {
      const response = await axios.get("http://localhost:8080/api/tables");
      // Sorting tables by ID in ascending order
      const sortedTables = response.data.sort((a, b) => a.ID - b.ID);
      setTables(sortedTables); // Store sorted tables into state
    } catch (error) {
      console.error("Error fetching tables:", error);
    }
  };

  fetchTables(); // Fetch tables when the component mounts

    // Optionally, you can reconnect the WebSocket here to listen for table updates
    const connectWebSocket = () => {
      try {
        const socket = new WebSocket("ws://localhost:8080/ws/tables");

        socket.onopen = () => {
          console.log("WebSocket connected successfully!");
        };

        socket.onerror = (error) => {
          console.error("WebSocket error:", error);
        };

        socket.onmessage = (event) => {
          const message = JSON.parse(event.data);
          if (message.type === "table_update") {
            setTables(message.data);
          }
        };

        socket.onclose = () => {
          console.log("WebSocket connection closed. Attempting reconnect...");
          setTimeout(connectWebSocket, 5000);
        };

        return socket;
      } catch (error) {
        console.error("Failed to connect to WebSocket:", error);
        setTimeout(connectWebSocket, 5000);
      }
    };

    const socket = connectWebSocket();

    return () => {
      socket.close();
    };
  }, []); // Empty dependency array to run only on mount

   // Check if table name is duplicated
   const checkTableNameConflict = async () => {
    try {
      const response = await axios.get("http://localhost:8080/api/tables");
      const data = response.data;
      return data.some((table) => table.Name === newTableName);
    } catch (error) {
      console.error("Error fetching tables:", error);
      return false;
    }
  };

  const handleTableAction = async (action, table) => {
    if (!table || !table.ID) {
      alert("ข้อมูลโต๊ะไม่ถูกต้อง");
      return;
    }

    const endpoint =
      action === "reserve"
        ? `http://localhost:8080/api/table/reservedTable/${table.ID}`
        : action === "unreserve"
        ? `http://localhost:8080/api/table/unreservedTable/${table.ID}`
        : null;

    if (!endpoint) {
      console.warn("Unsupported action:", action);
      return;
    }

    try {
      const response = await axios.post(endpoint, {});
      alert(response.data.message || "ดำเนินการสำเร็จ");
      setTables((prevTables) =>
        prevTables.map((t) =>
          t.ID === table.ID
            ? { ...t, Status: action === "reserve" ? "reserved" : "available" }
            : t
        )
      );
    } catch (error) {
      console.error(`Error processing ${action}:`, error);
      alert("เกิดข้อผิดพลาดในการดำเนินการ");
    }
  };

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

  // Create a new table
  const createTable = async () => {
    const isConflict = await checkTableNameConflict();
    if (isConflict) {
      setIsNameConflict(true);
      setMessage("ชื่อโต๊ะซ้ำกับโต๊ะที่มีอยู่แล้ว");
      return;
    }

    const tableData = {
      capacity: parseInt(newTableCapacity),
      name: newTableName,
      status: newTableStatus,
    };

    try {
      const response = await axios.post("http://localhost:8080/api/table", tableData, {
        headers: {
          "Content-Type": "application/json",
          accept: "application/json",
        },
      });

      const data = response.data;

      if (response.status === 200) {
        setMessage(`โต๊ะ "${data.Name}" ถูกสร้างเรียบร้อยแล้ว!`);
        setTables((prevTables) => [...prevTables, data]);
        setIsDialogOpen(false);
      } else {
        setMessage(`เกิดข้อผิดพลาด: ${data.error || "ข้อมูลไม่ถูกต้อง"}`);
      }
    } catch (error) {
      setMessage("เกิดข้อผิดพลาดในการเชื่อมต่อ API");
      console.error(error);
    }
  };

  const handlePrimaryTableSelection = (table) => {
    setPrimaryTable(table); // เมื่อผู้ใช้เลือกโต๊ะหลัก
  };

  const handleSelectTable = (tableId) => {
    setSelectedTables((prevSelected) =>
      prevSelected.includes(tableId)
        ? prevSelected.filter((id) => id !== tableId)
        : [...prevSelected, tableId]
    );
  };

  const mergeTables = async () => {
  if (selectedTables.length < 2) {
    alert("กรุณาเลือกโต๊ะอย่างน้อย 2 โต๊ะ");
    return;
  }

  try {
    const response = await axios.post("http://localhost:8080/api/table/mergeTable", {
      table_ids: selectedTables,
    });

    if (response.status === 200) {
      const { group_id, main_table_id, message } = response.data;
      alert(message || "รวมโต๊ะสำเร็จ");

      // อัปเดตสถานะโต๊ะที่รวมแล้ว และเก็บค่า group_id
      setTables((prevTables) =>
        prevTables.map((table) =>
          selectedTables.includes(table.ID)
            ? { 
                ...table, 
                Status: "merged", 
                GroupID: group_id, 
                MainTableID: main_table_id 
              }
            : table
        )
      );

      // เรียงโต๊ะตาม ID จากน้อยไปมาก
      setTables(prevTables => {
        const sortedTables = prevTables.sort((a, b) => a.ID - b.ID);
        return sortedTables;
      });

      // เคลียร์โต๊ะที่เลือก
      setSelectedTables([]); 
    } else {
      alert("ไม่สามารถรวมโต๊ะได้");
    }
  } catch (error) {
    console.error("Error merging tables:", error);
    alert("เกิดข้อผิดพลาดในการรวมโต๊ะ");
  }
};

  const handleMergeTables = async () => {
    try {
      if (!primaryTable || selectedTables.length === 0) {
        alert("กรุณาเลือกโต๊ะหลักและโต๊ะที่ต้องการรวม");
        return;
      }

      const response = await axios.post("http://localhost:8080/api/table/mergeTable", {
        primary_table: primaryTable.ID, // Send ID of primary table
        selected_table_ids: selectedTables.map((table) => table.ID), // Send IDs of selected tables
      });

      if (response.status === 200) {
        alert("รวมโต๊ะสำเร็จ");
        // Optionally, update tables state to reflect the merged status
      } else {
        alert("ไม่สามารถรวมโต๊ะได้");
      }
    } catch (error) {
      console.error("เกิดข้อผิดพลาดในการรวมโต๊ะ:", error);
      alert("เกิดข้อผิดพลาดในการรวมโต๊ะ");
    }
  };

  const handleSplitTable = async (groupId) => {
    try {
      const response = await axios.post(`http://localhost:8080/api/table/splitTable/${groupId}`);
  
      if (response.status === 200) {
        alert(response.data.message || "แยกโต๊ะสำเร็จ");
  
        // อัปเดตสถานะของโต๊ะกลับไปเป็นสถานะเดิม
        setTables((prevTables) =>
          prevTables.map((table) =>
            table.GroupID === groupId
              ? { ...table, Status: "available", GroupID: null, MainTableID: null }
              : table
          )
        );
      } else {
        alert("ไม่สามารถแยกโต๊ะได้");
      }
    } catch (error) {
      console.error("เกิดข้อผิดพลาดในการแยกโต๊ะ:", error);
      alert("เกิดข้อผิดพลาดในการแยกโต๊ะ");
    }
  };

  // Table rendering
  const renderTableActionButtons = (table) => {
    const status = table.Status || "available";
    const tableStyles = status === "reserved" ? styles.reservedTable : status === "merged" ? styles.mergedTable : styles.availableTable;

    return (
      <div className={tableStyles}>
        {status === "available" && (
          <button className="bg-blue-500 text-white px-4 py-2 rounded mb-2" onClick={() => handleTableAction("reserve", table)}>
            จอง
          </button>
        )}
        {status === "reserved" && (
          <button className="bg-gray-500 text-white px-4 py-2 rounded" onClick={() => handleTableAction("unreserve", table)}>
            ยกเลิกการจอง
          </button>
        )}
        {status === "available" && (
          <button className="bg-gray-500 text-white px-4 py-2 rounded" onClick={() => handleSplitTable(table.GroupID)}>
            แยกโต๊ะ
          </button>
        )}
      </div>
    );
  };

  return (
    <div className="h-screen overflow-auto p-4 lg:ml-60">
      {/* Render tables */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {tables.map((table) => (
          <div key={table.ID} className="border p-4 rounded-lg shadow-lg">
            <h3 className="text-xl font-bold mb-2">{table.Name}</h3>
            <TableInfo label="Capacity" value={table.Capacity} />
            <TableInfo label="Status" value={table.Status} />
            {renderTableActionButtons(table)}
            {/* Additional table actions can go here */}
            
          </div>
        ))}
      </div>

      <button onClick={() => setIsDialogOpen(true)} className="bg-blue-500 text-white px-4 py-2 rounded mt-4">
        สร้างโต๊ะใหม่
      </button>

      <button onClick={() => setIsMergeDialogOpen(true)} className="bg-blue-500 text-white px-4 py-2 rounded">
        รวมกลุ่มโต๊ะ
      </button>

      {/* Merge Dialog */}
      {isMergeDialogOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-lg w-96">
            <h3 className="text-xl font-bold mb-4">เลือกโต๊ะสำหรับรวม</h3>
            <div className="mb-4">
              {tables.map((table) => (
                <div key={table.ID} className="flex items-center mb-2">
                  <input
                    type="checkbox"
                    checked={selectedTables.includes(table.ID)}
                    onChange={() => handleSelectTable(table.ID)}
                    className="mr-2"
                  />
                  <label className="text-gray-800">{table.Name}</label>
                </div>
              ))}
            </div>
            <div className="flex justify-between">
              <button onClick={() => { mergeTables(); setIsMergeDialogOpen(false); }} className="bg-green-500 text-white px-4 py-2 rounded">
                รวมโต๊ะ
              </button>
              <button onClick={() => setIsMergeDialogOpen(false)} className="bg-red-500 text-white px-4 py-2 rounded">
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
            <input
              type="text"
              value={newTableName}
              onChange={(e) => setNewTableName(e.target.value)}
              className="border p-2 mb-4 w-full"
              placeholder="ชื่อโต๊ะ"
            />
            <input
              type="number"
              value={newTableCapacity}
              onChange={(e) => setNewTableCapacity(e.target.value)}
              className="border p-2 mb-4 w-full"
              placeholder="ความจุ"
            />
            <select
              value={newTableStatus}
              onChange={(e) => setNewTableStatus(e.target.value)}
              className="border p-2 mb-4 w-full"
            >
              <option value="ว่าง">ว่าง</option>
              <option value="จอง">จอง</option>
              <option value="ไม่ว่าง">ไม่ว่าง</option>
            </select>
            {isNameConflict && <div className="text-red-500 text-sm mb-4">{message}</div>}
            <div className="flex justify-between">
              <button onClick={createTable} className="bg-blue-500 text-white px-4 py-2 rounded">
                สร้างโต๊ะ
              </button>
              <button onClick={() => setIsDialogOpen(false)} className="bg-red-500 text-white px-4 py-2 rounded">
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
