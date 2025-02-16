// MenuTable.js
import React, { useState, useEffect } from "react";
import { Image, Plus, Search, Filter, X, User, Edit, Trash2 } from 'lucide-react';


const API_BASE_URL = `${import.meta.env.VITE_APP_API_URL}/api/menu"`; // URL ของ API หลัก

const MenuTable = ({ menuItemId }) => {
  const [menuOptions, setMenuOptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchMenuOptions = async () => {
      if (!menuItemId) return; // ถ้าไม่มี menuItemId ไม่ต้องทำอะไร
      setLoading(true);
      try {
        const token = localStorage.getItem('token')
        const response = await axios.get(`${import.meta.env.VITE_APP_API_URL}/api/menu/options/${menuItemId}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        setMenuOptions(response.data);
      } catch (error) {
        console.error("ไม่สามารถดึงข้อมูลตัวเลือกของเมนูได้:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchMenuOptions();
  }, [menuItemId]); // ค้นหาข้อมูลเมนูเมื่อ menuItemId เปลี่ยน

  if (loading) {
    return <div>กำลังโหลด...</div>;
  }

  if (error) {
    return <div>Error: {error}</div>;
  }

  return (
    <div>
      <h2>Option Groups for Menu Item {menuItemId}</h2>
      <table className="table-auto w-full border-collapse">
        <thead>
          <tr>
            <th className="border-b px-4 py-2">Group ID</th>
            <th className="border-b px-4 py-2">Group Name</th>
            <th className="border-b px-4 py-2">Max Selections</th>
            <th className="border-b px-4 py-2">Required</th>
            <th className="border-b px-4 py-2">Option ID</th>
            <th className="border-b px-4 py-2">Option Name</th>
            <th className="border-b px-4 py-2">Price (THB)</th>
          </tr>
        </thead>
        <tbody>
          {options.map((group) => (
            group.Options.map((option) => (
              <tr key={option.ID}>
                <td className="border-b px-4 py-2">{group.ID}</td>
                <td className="border-b px-4 py-2">{group.Name}</td>
                <td className="border-b px-4 py-2">{group.MaxSelections}</td>
                <td className="border-b px-4 py-2">{group.IsRequired ? 'Yes' : 'No'}</td>
                <td className="border-b px-4 py-2">{option.ID}</td>
                <td className="border-b px-4 py-2">{option.Name}</td>
                <td className="border-b px-4 py-2">{option.Price}</td>
              </tr>
            ))
          ))}
        </tbody>
      </table>
    </div>
  );
};
  
export default MenuTable;
