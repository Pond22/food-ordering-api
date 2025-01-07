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
  const [searchTerm, setSearchTerm] = useState(""); // คำค้นหา
  const [deletedCategories, setDeletedCategories] = useState([]); // รายการหมวดหมู่ที่ถูกลบ
  const [activeTab, setActiveTab] = useState('addCat');// state สลับการแสดงตาราง
  const [showRestoreModal, setShowRestoreModal] = useState(false); // Modal สำหรับการยืนยันการกู้คืนหมวดหมู่
  const [categoryToRestore, setCategoryToRestore] = useState(null); // หมวดหมู่ที่ต้องการกู้คืน
  const [showDeleteModal, setShowDeleteModal] = useState(false);
const [categoryToDelete, setCategoryToDelete] = useState(null);
const [deleteType, setDeleteType] = useState(null); // false สำหรับลบแค่หมวดหมู่, true สำหรับลบพร้อมเมนู

  const API_BASE_URL = "http://127.0.0.1:8080/api/categories"; // URL ของ API

  // ฟังก์ชันเพื่อดึงข้อมูลหมวดหมู่ที่ถูกลบ
  const fetchDeletedCategories = async () => {
    try {
      const token = localStorage.getItem("token"); // ดึง token จาก localStorage
  
      if (!token) {
        setErrorMessage("โปรดล็อกอินก่อนใช้งาน");
        return; // ถ้าไม่มี token แสดงข้อความและหยุดการทำงาน
      }
  
      const response = await axios.get("http://127.0.0.1:8080/api/categories/get_delete_categories", {
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`, // ใช้ token ใน header
        },
      });
  
      setDeletedCategories(response.data); // เซ็ตข้อมูลหมวดหมู่ที่ถูกลบ
    } catch (error) {
      console.error(error);
      setErrorMessage("ไม่สามารถดึงข้อมูลหมวดหมู่ที่ถูกลบได้");
    }
  };
  useEffect(() => {
    fetchCategories();
    fetchDeletedCategories(); // เรียกดึงข้อมูลหมวดหมู่ที่ถูกลบ
  }, []);

  // ฟังก์ชันเพื่อดึงข้อมูลหมวดหมู่ทั้งหมด
  const fetchCategories = async () => {
    try {
      const token = localStorage.getItem("token"); // ดึง token จาก localStorage

      if (!token) {
        setErrorMessage("โปรดล็อกอินก่อนใช้งาน");
        return; // ถ้าไม่มี token แสดงข้อความและหยุดการทำงาน
      }

      const response = await axios.get(API_BASE_URL, {
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`, // ใช้ token ใน header
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
      const token = localStorage.getItem("token"); // ดึง token จาก localStorage

      if (!token) {
        setErrorMessage("โปรดล็อกอินก่อนใช้งาน");
        return; // ถ้าไม่มี token แสดงข้อความและหยุดการทำงาน
      }

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
            Authorization: `Bearer ${token}`, // ใช้ token ใน header
          },
        }
      );

      if (response.status === 200) {
        setSuccessMessage(`หมวดหมู่ "${response.data.nameEn}" ถูกเพิ่มเรียบร้อยแล้ว!`);
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
 // ฟังก์ชันสำหรับการลบหมวดหมู่
const handleDeleteCategory = async () => {
  if (!categoryToDelete) return;

  setLoading(true);

  try {
    const token = localStorage.getItem("token");
    const endpoint = deleteType
      ? `/delete_category_with_menu/${categoryToDelete.ID}`  // ลบพร้อมเมนู
      : `/delete_category/${categoryToDelete.ID}`;  // ลบแค่หมวดหมู่

    const response = await axios.delete(
      `${API_BASE_URL}${endpoint}`,
      {
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      }
    );

    if (response.status === 200) {
      setSuccessMessage(`หมวดหมู่ "${categoryToDelete.Name}" ถูกลบเรียบร้อยแล้ว!`);
      fetchDeletedCategories(); // รีเฟรชข้อมูลหมวดหมู่ที่ถูกลบ
      setShowDeleteModal(false); // ปิด Modal การลบ
    } else {
      setErrorMessage("เกิดข้อผิดพลาดในการลบหมวดหมู่");
    }
  } catch (error) {
    console.error(error);
    setErrorMessage("ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้");
  } finally {
    setLoading(false);
  }
};

  // ฟังก์ชันสำหรับการกรองหมวดหมู่ตามคำค้นหา
  const filteredCategories = categories.filter((category) => {
    const lowercasedSearchTerm = searchTerm.toLowerCase();
    return (
      category.ID.toString().includes(lowercasedSearchTerm) || // ค้นหาตาม ID
      category.Name.toLowerCase().includes(lowercasedSearchTerm) || // ค้นหาตามชื่อภาษาไทย
      category.NameCh.toLowerCase().includes(lowercasedSearchTerm) || // ค้นหาตามชื่อภาษาจีน
      category.NameEn.toLowerCase().includes(lowercasedSearchTerm) // ค้นหาตามชื่อภาษาอังกฤษ
    );
  });

// ฟังก์ชันสำหรับการกู้คืนหมวดหมู่
const handleRestoreCategory = async () => {
  if (!categoryToRestore) return;  // ตรวจสอบว่า categoryToRestore มีค่าหรือไม่

  setLoading(true);  // ตั้งค่า loading ให้เป็น true เมื่อเริ่มทำงาน

  try {
    const token = localStorage.getItem("token");  // รับ token จาก localStorage
    const response = await axios.post(
      `${API_BASE_URL}/restore_categories/${categoryToRestore.ID}`,  // URL ของ API สำหรับการกู้คืนหมวดหมู่
      {},  // ส่งข้อมูลเปล่าไปกับคำขอ
      {
        headers: {
          "Content-Type": "application/json",  // ระบุประเภทของข้อมูล
          Authorization: `Bearer ${token}`,  // ส่ง token ใน header
        },
      }
    );

    if (response.status === 200) {
      setSuccessMessage(`หมวดหมู่ "${response.data.category.Name}" กู้คืนเรียบร้อยแล้ว!`);  // แสดงข้อความสำเร็จ
      fetchDeletedCategories();  // รีเฟรชข้อมูลหมวดหมู่ที่ถูกลบ
      setShowRestoreModal(false);  // ปิด Modal การกู้คืน
    } else {
      setErrorMessage("เกิดข้อผิดพลาดในการกู้คืนหมวดหมู่");  // แสดงข้อความข้อผิดพลาด
    }
  } catch (error) {
    console.error(error);  // แสดงข้อผิดพลาดใน console
    setErrorMessage("ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้");  // แสดงข้อความหากไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้
  } finally {
    setLoading(false);  // ตั้งค่า loading เป็น false เมื่อเสร็จสิ้น
  }
};


  const [searchQuery, setSearchQuery] = useState("");

  // ฟังก์ชันกรองหมวดหมู่ตามคำค้นหา
  const filteredCategoriesDel = deletedCategories.filter((category) => {
    const searchTerm = searchQuery.toLowerCase();
    return (
      category.ID.toString().includes(searchTerm) || // ค้นหาตาม ID
      category.Name.toLowerCase().includes(searchTerm) || // ค้นหาตามชื่อ (ภาษาไทย)
      category.NameCh.toLowerCase().includes(searchTerm) || // ค้นหาตามชื่อ (ภาษาจีน)
      category.NameEn.toLowerCase().includes(searchTerm) // ค้นหาตามชื่อ (ภาษาอังกฤษ)
      
    );
  });

  return (
    <div className=" mx-auto p-6 bg-gray-50 flex items-center justify-center lg:ml-60">

      <div className="w-full bg-gray-800 rounded-lg shadow-lg">
        <div className="ml-12 text-2xl font-bold text-white mb-6 text-left m-5">
          จัดการหมวดหมู่ของเมนู
        </div>
         {/* ปุ่มเปลี่ยนตาราง */}
      <button
          className={`px-4 py-2 rounded-t-lg ${activeTab === 'addCat' ? 'bg-blue-500 text-white' : 'bg-gray-200 text-gray-700 hover:bg-gray-300'}`}
          onClick={() => setActiveTab('addCat')}
        >
          ข้อมูลหมวดหมู่ที่มีอยู่
        </button>
        <button
          className={`px-4 py-2 rounded-t-lg ${activeTab === 'viewDel' ? 'bg-blue-500 text-white' : 'bg-gray-200 text-gray-700 hover:bg-gray-300'}`}
          onClick={() => setActiveTab('viewDel')}
        >
          ข้อมูลหมวดหมู่ที่ถูกลบ
        </button>
         {/* สิ้นสุดปุ่มเปลี่ยนตาราง */}

         {activeTab === 'addCat' && (
        <div className="bg-gray-50 w-full p-8 rounded-lg">
          <div className="relative flex-grow flex justify-between">
            <input
              type="text"
              placeholder="ค้นหาเมนู"
              value={searchTerm} // ค่าของ input จะเชื่อมกับ state
              onChange={(e) => setSearchTerm(e.target.value)} // อัปเดตคำค้นหา
              className="w-2/3 p-1 pl-10 border rounded-md mb-3"
            />
            <Search className="absolute left-3 top-3 text-gray-500" size={20} />

            {/* แสดงป็อปอัพเมื่อกดปุ่มเพิ่ม */}
            <button
              onClick={() => setShowModal(true)}
              className="bg-blue-500 text-white font-medium py-2 px-4 rounded-lg hover:bg-blue-600 mb-4"
            >
              เพิ่มหมวดหมู่
            </button>
          </div>
          
          {/* ตารางแสดงข้อมูลหมวดหมู่ */}
          <table className="min-w-full bg-white table-auto border-collapse border border-gray-200">
            <thead className="border-b">
              <tr>
                <th className="p-4 text-center">ID</th>
                <th className="p-4 text-center">ชื่อ (ภาษาไทย)</th>
                <th className="p-4 text-center">ชื่อ (ภาษาจีน)</th>
                <th className="p-4 text-center">ชื่อ (ภาษาอังกฤษ)</th>
                <th className="p-4 text-center">จัดการ</th>
              </tr>
            </thead>
            <tbody>
              {filteredCategories.map((category) => (
                <tr key={category.ID} className="border-b hover:bg-gray-50">
                  <td className="p-4 text-center">{category.ID}</td>
                  <td className="p-4 text-center">{category.Name}</td>
                  <td className="p-4 text-center">{category.NameCh}</td>
                  <td className="p-4 text-center">{category.NameEn}</td>
                  <td className="p-4 text-center">
                  <button
        onClick={() => {
          setCategoryToDelete(category); // เซ็ตหมวดหมู่ที่ต้องการลบ
          setDeleteType(false); // ตั้งค่าให้ลบแค่หมวดหมู่
          setShowDeleteModal(true); // แสดง Modal การยืนยันการลบ
        }}
        className="text-red-600 hover:bg-red-100 p-2 rounded"
      >
        ลบแค่หมวดหมู่
      </button>
      <button
        onClick={() => {
          setCategoryToDelete(category); // เซ็ตหมวดหมู่ที่ต้องการลบ
          setDeleteType(true); // ตั้งค่าให้ลบพร้อมเมนู
          setShowDeleteModal(true); // แสดง Modal การยืนยันการลบ
        }}
        className="text-red-600 hover:bg-red-100 p-2 rounded ml-2"
      >
        ลบพร้อมเมนู
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
        </div>)}
      


        {activeTab === 'viewDel' && (
  <div className="bg-white w-full p-8 rounded-lg ">
    {/* ฟอร์มค้นหา */}
    <div className="mb-4">
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="ค้นหาหมวดหมู่..."
          className="w-2/3 px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        /> 
        <Search className="absolute left-3 top-3 text-gray-500" size={20} />
      </div>

    <h2 className="text-xl font-bold mb-4 text-gray-700">หมวดหมู่ที่ถูกลบ</h2>
    
    <table className="min-w-full table-auto border-collapse border border-gray-200">
      <thead>
        <tr>
          <th className="p-4 text-center border-b">ID</th>
          <th className="p-4 text-center border-b">ชื่อ (ภาษาไทย)</th>
          <th className="p-4 text-center border-b">ชื่อ (ภาษาจีน)</th>
          <th className="p-4 text-center border-b">ชื่อ (ภาษาอังกฤษ)</th>
          <th className="p-4 text-center border-b">จัดการ</th>
        </tr>
      </thead>
      <tbody>
        {deletedCategories.map && filteredCategoriesDel.map ((category) => (
          <tr key={category.ID} className="border-b hover:bg-gray-50">
            <td className="p-4 text-center">{category.ID}</td>
            <td className="p-4 text-center">{category.Name || "N/A"}</td>
            <td className="p-4 text-center">{category.NameCh || "N/A"}</td>
            <td className="p-4 text-center">{category.NameEn || "N/A"}</td>
            <td className="p-4 text-center">
              <button
                onClick={() => {
                  setCategoryToRestore(category); // เซ็ตหมวดหมู่ที่ต้องการกู้คืน
                  setShowRestoreModal(true); // แสดง Modal ยืนยัน
                }}
                className="text-blue-600 hover:bg-blue-100 p-2 rounded"
              >
                กู้คืน
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
</div>)}

</div>

 {/* Modal สำหรับการยืนยันการกู้คืนหมวดหมู่ */}
 {showRestoreModal && categoryToRestore && (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
        <div className="bg-white p-8 rounded-lg shadow-lg w-96">
          <div className="text-right">
            <button
              onClick={() => setShowRestoreModal(false)}
              className="top-2 right-2 text-gray-700 text-xl"
            >
              <X />
            </button>
          </div>
          <h2 className="text-xl font-bold mb-4 text-center">ยืนยันการกู้คืน</h2>
          <p className="text-center mb-4">
            คุณต้องการกู้คืนหมวดหมู่ "{categoryToRestore?.Name}" และเมนูทั้งหมดในหมวดหมู่นี้หรือไม่?
          </p>
          <div className="flex justify-between">
            <button
              onClick={() => setShowRestoreModal(false)}
              className="bg-gray-400 text-white py-2 px-4 rounded"
            >
              ยกเลิก
            </button>
            <button
              onClick={handleRestoreCategory} // ฟังก์ชันการกู้คืนหมวดหมู่
              className="bg-blue-500 text-white py-2 px-4 rounded"
              disabled={loading}
            >
              {loading ? "กำลังกู้คืน..." : "ยืนยันการกู้คืน"}
            </button>
          </div>
        </div>
      </div>
      )}

 {/* Modal สำหรับการยืนยันการลบ */}
{showDeleteModal && categoryToDelete && (
  <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
    <div className="bg-white p-8 rounded-lg shadow-lg w-96">
      <div className="text-right">
        <button
          onClick={() => setShowDeleteModal(false)}
          className="top-2 right-2 text-gray-700 text-xl"
        >
          <X />
        </button>
      </div>
      <h2 className="text-xl font-bold mb-4 text-center">ยืนยันการลบ</h2>
      <p className="text-center mb-4">
        คุณต้องการลบหมวดหมู่ "{categoryToDelete.Name}" {deleteType ? "พร้อมเมนูทั้งหมด" : "หรือไม่?"}
      </p>
      <div className="flex justify-between">
        <button
          onClick={() => setShowDeleteModal(false)}
          className="bg-gray-400 text-white py-2 px-4 rounded"
        >
          ยกเลิก
        </button>
        <button
          onClick={handleDeleteCategory} // ฟังก์ชันการลบ
          className="bg-red-500 text-white py-2 px-4 rounded"
          disabled={loading}
        >
          {loading ? "กำลังลบ..." : "ยืนยันการลบ"}
        </button>
      </div>
    </div>
  </div>
)}

      {/* ป็อปอัพสำหรับการเพิ่มหมวดหมู่ */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-8 rounded-lg shadow-lg w-96">
            <div className="text-right">{/* ปิดป็อปอัพ */}
            <button
              onClick={() => setShowModal(false)}
              className=" top-2 right-2 text-gray-700 text-xl"
            >
             <X />
            </button></div>
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

            
          </div>
        </div>
      )}
    </div>
  );
};

export default AddCategory;
