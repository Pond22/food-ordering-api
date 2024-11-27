import React, { useState, useEffect } from "react";
import axios from "axios";

const MenuManagement = () => {
  const [menus, setMenus] = useState([]); // เก็บข้อมูลเมนูทั้งหมด
  const [showAddMenuModal, setShowAddMenuModal] = useState(false); // state สำหรับแสดง/ซ่อน modal เพิ่มเมนู
  const [showAddCategoryModal, setShowAddCategoryModal] = useState(false); // state สำหรับแสดง/ซ่อน modal เพิ่มหมวดหมู่

  // Fetch ข้อมูลเมนูจากฐานข้อมูลเมื่อโหลดหน้า
  useEffect(() => {
    fetchMenus();
  }, []);

  const fetchMenus = async () => {
    try {
      const response = await axios.get("http://127.0.0.1:8080/getmenu?action=getAll");
      if (response.status === 200) {
        console.log(response.data)
        setMenus(response.data);
      }
    } catch (error) {
      console.error("Error fetching menus:", error);
      alert("ไม่สามารถดึงข้อมูลเมนูได้");
    }
  };

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Header */}
      <div className="flex justify-between items-center bg-white shadow p-4">
        <h1 className="text-xl font-bold text-gray-800">จัดการเมนูอาหาร</h1>
        <div>
          <button
            className="bg-blue-500 text-white px-4 py-2 rounded-lg mr-2 hover:bg-blue-600"
            onClick={() => setShowAddMenuModal(true)}
          >
            เพิ่มเมนูอาหาร
          </button>
          <button
            className="bg-green-500 text-white px-4 py-2 rounded-lg hover:bg-green-600"
            onClick={() => setShowAddCategoryModal(true)}
          >
            เพิ่มหมวดหมู่
          </button>
        </div>
      </div>

      {/* ตารางแสดงรายการเมนู */}
      <div className="p-4">
        <table className="w-full bg-white rounded-lg shadow">
          <thead>
            <tr className="bg-gray-200 text-left">
              <th className="py-2 px-4">ชื่อเมนู</th>
              <th className="py-2 px-4">หมวดหมู่</th>
              <th className="py-2 px-4">ราคา</th>
              <th className="py-2 px-4">คำอธิบาย</th>
            </tr>
          </thead>
          <tbody>
            {menus.map((menu) => (
              <tr key={menu.id} className="border-t">
                <td className="py-2 px-4">{menu.Name}</td>
                <td className="py-2 px-4">{menu.CategoryID}</td>
                <td className="py-2 px-4">{menu.Price}</td>
                <td className="py-2 px-4">{menu.Description}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Modal สำหรับเพิ่มเมนูอาหาร */}
      {showAddMenuModal && (
        <AddMenuModal onClose={() => setShowAddMenuModal(false)} onMenuAdded={fetchMenus} />
      )}

      {/* Modal สำหรับเพิ่มหมวดหมู่ */}
      {showAddCategoryModal && (
        <AddCategoryModal onClose={() => setShowAddCategoryModal(false)} />
      )}
    </div>
  );
};

// Modal สำหรับเพิ่มเมนูอาหาร
const AddMenuModal = ({ onClose, onMenuAdded }) => {
  const [menuName, setMenuName] = useState("");
  const [price, setPrice] = useState("");
  const [category, setCategory] = useState("");
  const [categories, setCategories] = useState([]);
  const [image, setImage] = useState(null);

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        const response = await axios.get("http://localhost:8080/getCategory");
        if (response.status === 200) {
          console.log(response.data)
          setCategories(response.data);
        }
      } catch (error) {
        console.error("Error fetching categories:", error);
      }
    };
    fetchCategories();
  }, []);

  const handleAddMenu = async (e) => {
    e.preventDefault();

    const formData = new FormData();
    formData.append("name", menuName);
    formData.append("price", price);
    formData.append("CategoryID", category);
    formData.append("description", "เมนูอาหาร");
    if (image) formData.append("image", image);

    try {
      const response = await axios.post("http://127.0.0.1:8080/add_menu", formData, {
        body: { "Content-Type": "multipart/form-data" },
      });
      if (response.status === 200) {
        alert("เพิ่มเมนูสำเร็จ!");
        onMenuAdded(); // อัปเดตตารางเมนู
        onClose(); // ปิด modal
      }
    } catch (error) {
      console.error("Error adding menu:", error);
      alert("เกิดข้อผิดพลาดในการเพิ่มเมนู");
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center">
      <div className="bg-white rounded-lg p-8 shadow-lg w-96">
        <h2 className="text-xl font-bold mb-4">เพิ่มเมนูอาหาร</h2>
        <form onSubmit={handleAddMenu} className="space-y-4">
          <input
            type="text"
            placeholder="ชื่อเมนู"
            value={menuName}
            onChange={(e) => setMenuName(e.target.value)}
            className="w-full px-4 py-2 border rounded-lg"
            required
          />
          <input
            type="number"
            placeholder="ราคา"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            className="w-full px-4 py-2 border rounded-lg"
            required
          />
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            className="w-full px-4 py-2 border rounded-lg"
            required
          >
            <option value="" disabled>
              เลือกหมวดหมู่
            </option>
            {categories.map((cat) => (
              <option key={cat.ID} value={cat.ID}>
                {cat.Name}
              </option>
            ))}
          </select>
          <input
            type="file"
            onChange={(e) => setImage(e.target.files[0])}
            className="w-full px-4 py-2 border rounded-lg"
          />
          <div className="flex justify-end space-x-2">
            <button type="button" onClick={onClose} className="px-4 py-2 bg-gray-300 rounded-lg">
              ยกเลิก
            </button>
            <button type="submit" className="px-4 py-2 bg-blue-500 text-white rounded-lg">
              เพิ่มเมนู
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};



// Modal สำหรับเพิ่มหมวดหมู่
const AddCategoryModal = ({ onClose }) => {
  const [categoryName, setCategoryName] = useState("");

  const handleAddCategory = async (e) => {
    e.preventDefault();

    try {
      const response = await axios.post("http://127.0.0.1:8080/add_category",{
        name: categoryName.trim(),
      });
      if (response.status === 200) {
        console.log(response.data)
        alert(`เพิ่มหมวดหมู่ "${response.data.Name}" สำเร็จ!`);
        onClose(); // ปิด modal
      }
      else {
      alert(`ไม่สามารถเพิ่มหมวดหมู่ได้: ${response.statusText}`);
    }
    } catch (error) {
      if (error.response) {
        // ข้อผิดพลาดจากฝั่งเซิร์ฟเวอร์
        alert(`เกิดข้อผิดพลาดจากเซิร์ฟเวอร์: ${error.response.data.message || "ไม่ทราบสาเหตุ"}`);
      } else if (error.request) {
        // ข้อผิดพลาดในการเชื่อมต่อ
        alert("ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้");
      } else {
        // ข้อผิดพลาดที่ไม่ทราบสาเหตุ
        alert(`ข้อผิดพลาด: ${error.message}`);
      }
      console.error("Error adding category:", error);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center">
      <div className="bg-white rounded-lg p-8 shadow-lg w-96">
        <h2 className="text-xl font-bold mb-4">เพิ่มหมวดหมู่</h2>
        <form onSubmit={handleAddCategory} className="space-y-4">
          <input
            type="text"
            placeholder="ชื่อหมวดหมู่"
            value={categoryName}
            onChange={(e) => setCategoryName(e.target.value)}
            className="w-full px-4 py-2 border rounded-lg"
            required
          />
          <div className="flex justify-end space-x-2">
            <button type="button" onClick={onClose} className="px-4 py-2 bg-gray-300 rounded-lg">
              ยกเลิก
            </button>
            <button type="submit" className="px-4 py-2 bg-green-500 text-white rounded-lg">
              เพิ่มหมวดหมู่
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default MenuManagement;
