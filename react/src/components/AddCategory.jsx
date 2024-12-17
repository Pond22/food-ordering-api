import React, { useState, useEffect } from "react";
import axios from "axios";
import { ShieldCheck, Plus, Search, Filter, X, User, Edit, Trash2 } from 'lucide-react';

const AddCategory = () => {
  const [categoryNameTH, setCategoryNameTH] = useState(""); // ชื่อหมวดหมู่ภาษาไทย
  const [categoryNameCH, setCategoryNameCH] = useState(""); // ชื่อหมวดหมู่ภาษาจีน
  const [categoryNameEN, setCategoryNameEN] = useState(""); // ชื่อหมวดหมู่ภาษาอังกฤษ
  const [categories, setCategories] = useState([]); // รายการหมวดหมู่ทั้งหมด
  const [successMessage, setSuccessMessage] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false); // ควบคุมการแสดง popup

  const API_BASE_URL = "http://127.0.0.1:8080/api/categories"; // URL ของ API

  // ฟังก์ชันเพื่อดึงข้อมูลหมวดหมู่ทั้งหมด
  const fetchCategories = async () => {
    try {
      const response = await axios.get(API_BASE_URL, {
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer YOUR_ACCESS_TOKEN`, // แทนที่ด้วย access token ที่ถูกต้อง
        },
      });
      setCategories(response.data); // เซ็ตข้อมูลหมวดหมู่ที่ได้รับจาก API
    } catch (error) {
      console.error(error);
      setErrorMessage("ไม่สามารถดึงข้อมูลหมวดหมู่ได้");
    }
  };

  // เรียกฟังก์ชัน fetchCategories เมื่อคอมโพเนนต์โหลดครั้งแรก
  useEffect(() => {
    fetchCategories();
  }, []);

  // ฟังก์ชันสำหรับการเพิ่มหมวดหมู่
  const handleCategorySubmit = async (e) => {
    e.preventDefault();

    if (!categoryNameTH || !categoryNameCH || !categoryNameEN) {
      setErrorMessage("กรุณากรอกชื่อหมวดหมู่ทุกภาษา");
      return;
    }

    setLoading(true);
    setSuccessMessage("");
    setErrorMessage("");

    try {
      const response = await axios.post(
        API_BASE_URL,
        {
          name: categoryNameTH,
          nameEn: categoryNameEN,
          nameCh: categoryNameCH,
        },
        {
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer YOUR_ACCESS_TOKEN`,
          },
        }
      );

      if (response.status === 200) {
        setSuccessMessage(`หมวดหมู่ "${response.data.name_en}" ถูกเพิ่มเรียบร้อยแล้ว!`);
        setCategoryNameTH("");
        setCategoryNameEN("");
        setCategoryNameCH("");
        fetchCategories(); // รีเฟรชข้อมูลหลังจากเพิ่มหมวดหมู่
        setShowModal(false); // ปิดป็อปอัพ
      }
    } catch (error) {
      if (error.response) {
        setErrorMessage("เกิดข้อผิดพลาดในการเพิ่มหมวดหมู่");
      } else {
        setErrorMessage("ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้");
      }
    } finally {
      setLoading(false);
    }
  };

  // ฟังก์ชันสำหรับการลบหมวดหมู่
  const handleDelete = async (id) => {
    try {
      await axios.delete(`${API_BASE_URL}/${id}`, {
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer YOUR_ACCESS_TOKEN`,
        },
      });
      fetchCategories(); // รีเฟรชข้อมูลหลังจากลบหมวดหมู่
    } catch (error) {
      console.error(error);
      setErrorMessage("ไม่สามารถลบหมวดหมู่ได้");
    }
  };

  return (
    <div className="w-full mx-auto p-6 bg-gray-50 flex items-center justify-center ">
      <div className="w-full bg-blue-600  rounded-lg shadow-lg ">
        
        <h1 className="text-2xl font-bold text-white mb-6 text-left m-5">
          หมวดหมู่ทั้งหมด
        </h1>
        <div className="bg-white w-full p-8 rounded-lg">
        
        <div className="relative flex-grow flex justify-between">
            <input
              type="text"
              placeholder="ค้นหาเมนู"
              // value={}
              // onChange={(e) => setSearchTerm(e.target.value)}
              className="w-2/3 p-1 pl-10 border rounded-md"
            />
            <Search className="absolute left-3 top-3 text-gray-500" size={20} />
          
        {/* แสดงป็อปอัพเมื่อกดปุ่มเพิ่ม */}
        <button
          onClick={() => setShowModal(true)}
          className="bg-blue-500 text-white font-medium py-2 px-4 rounded-lg hover:bg-blue-600 mb-4"
        >
          เพิ่มหมวดหมู่
        </button></div>

        {/* ตารางแสดงข้อมูลหมวดหมู่ */}
        <table className="min-w-full table-auto">
          <thead>
            <tr>
              <th className="p-4 text-lef">ID</th>
              <th className="p-4 text-lef">ชื่อ (ภาษาไทย)</th>
              <th className="p-4 text-lef">ชื่อ (ภาษาจีน)</th>
              <th className="p-4 text-lef">ชื่อ (ภาษาอังกฤษ)</th>
              <th className="p-4 text-lef">จัดการ</th>
            </tr>
          </thead>
          <tbody>
            {categories.map((category) => (
              <tr key={category.id} className="border-b hover:bg-gray-50">
                <td className="p-4 text-center ">{category.ID}</td>
                <td className="p-4 text-center ">{category.Name}</td>
                <td className="p-4 text-center ">{category.NameCh}</td>
                <td className="p-4 text-center ">{category.NameEn}</td>
                <td className="p-4 text-lef ">
                  <button
                    onClick={() => handleDelete(category.id)}
                    className="text-red-600 hover:bg-red-100 p-2 rounded "
                  >
                    <Trash2 size={20} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {/* ข้อความแสดงผลลัพธ์ */}
        {successMessage && (
          <div className="mt-4 text-green-600 font-medium text-center">{successMessage}</div>
        )}
        {errorMessage && (
          <div className="mt-4 text-red-600 font-medium text-center">{errorMessage}</div>
        )}
      </div>
      </div>
      {/* ป็อปอัพสำหรับการเพิ่มหมวดหมู่ */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-8 rounded-lg shadow-lg w-96">
            <h2 className="text-xl font-bold mb-4 text-center">เพิ่มหมวดหมู่ใหม่</h2>
            <form onSubmit={handleCategorySubmit} className="space-y-4">
              <div>
                <label htmlFor="categoryNameTH" className="block text-gray-700 font-medium mb-2">
                  ชื่อหมวดหมู่ (ภาษาไทย)
                </label>
                <input
                  id="categoryNameTH"
                  type="text"
                  value={categoryNameTH}
                  onChange={(e) => setCategoryNameTH(e.target.value)}
                  className="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="ใส่ชื่อหมวดหมู่ (ภาษาไทย)"
                  required
                />
              </div>

              <div>
                <label htmlFor="categoryNameCH" className="block text-gray-700 font-medium mb-2">
                  ชื่อหมวดหมู่ (ภาษาจีน)
                </label>
                <input
                  id="categoryNameCH"
                  type="text"
                  value={categoryNameCH}
                  onChange={(e) => setCategoryNameCH(e.target.value)}
                  className="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="ใส่ชื่อหมวดหมู่ (ภาษาจีน)"
                  required
                />
              </div>

              <div>
                <label htmlFor="categoryNameEN" className="block text-gray-700 font-medium mb-2">
                  ชื่อหมวดหมู่ (ภาษาอังกฤษ)
                </label>
                <input
                  id="categoryNameEN"
                  type="text"
                  value={categoryNameEN}
                  onChange={(e) => setCategoryNameEN(e.target.value)}
                  className="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="ใส่ชื่อหมวดหมู่ (ภาษาอังกฤษ)"
                  required
                />
              </div>

              <div>
                <button
                  type="submit"
                  className={`w-full ${loading ? 'bg-gray-400' : 'bg-blue-500'} text-white font-medium py-2 px-4 rounded-lg hover:bg-blue-600 transition duration-200`}
                  disabled={loading}
                >
                  {loading ? "กำลังเพิ่ม..." : "เพิ่มหมวดหมู่"}
                </button>
              </div>
            </form>

            {/* ปิดป็อปอัพ */}
            <button
              onClick={() => setShowModal(false)}
              className="absolute top-2 right-2 text-gray-700 text-xl"
            >
              ×
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default AddCategory;
