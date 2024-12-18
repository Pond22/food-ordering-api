import React, { useState, useEffect } from "react";
import axios from "axios";

const API_BASE_URL = "http://127.0.0.1:8080/api/menu"; // กำหนด URL ของ API

const MenuManagement = () => {
  const [menus, setMenus] = useState([]);
  const [showAddMenuModal, setShowAddMenuModal] = useState(false);
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [showUploadModal, setShowUploadModal] = useState(false); // state สำหรับแสดง popup upload รูป
  const [selectedImage, setSelectedImage] = useState(null); // state สำหรับเก็บรูปที่เลือก
  const [currentMenuId, setCurrentMenuId] = useState(null); // state สำหรับเก็บ ID เมนูที่ต้องการอัพเดทรูป

  // เมื่อ showUploadModal เปลี่ยนสถานะ จะรีเซ็ตค่า selectedImage และ currentMenuId
  useEffect(() => {
    if (!showUploadModal) {
      setSelectedImage(null); // เคลียร์ภาพเมื่อปิด popup
      setCurrentMenuId(null); // เคลียร์ ID ของเมนูเมื่อปิด popup
    }
  }, [showUploadModal]); // useEffect จะทำงานเมื่อ showUploadModal เปลี่ยน

  useEffect(() => {
    fetchMenus("getAll");
  }, []);

  const fetchMenus = async (action, id = null) => {
    try {
      let url = `${API_BASE_URL}?action=${action}`;
      if (action === "getByID" && id) {
        url += `&id=${id}`;
      }
  
      const response = await axios.get(url);
  
      if (response.status === 200) {
        // เรียงลำดับข้อมูลเมนูจากน้อยไปมากตาม ID
        const sortedMenus = response.data.sort((a, b) => a.ID - b.ID);
        setMenus(sortedMenus); // ตั้งค่า menus ให้เป็นข้อมูลที่เรียงลำดับแล้ว
      }
    } catch (error) {
      console.error("Error fetching menus:", error);
      alert("ไม่สามารถดึงข้อมูลเมนูได้");
    }
  };

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      setSelectedImage(URL.createObjectURL(file)); // แสดงตัวอย่างรูปภาพที่เลือก
    }
  };

  const handleUploadImage = async () => {
    if (!selectedImage || !currentMenuId) {
      alert("กรุณาเลือกภาพและเมนูที่ต้องการอัพเดท");
      return;
    }

    const formData = new FormData();
    const file = document.querySelector("#imageInput").files[0];
    formData.append("image", file);

    try {
      const response = await axios.put(
        `${API_BASE_URL}/image/${currentMenuId}`,
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        }
      );

      if (response.status === 200) {
        alert("อัพโหลดรูปภาพสำเร็จ");
        setShowUploadModal(false); // ปิด popup หลังจากอัพโหลดเสร็จ
        fetchMenus("getAll"); // ดึงข้อมูลเมนูใหม่
      } else {
        alert("เกิดข้อผิดพลาดในการอัพโหลดรูปภาพ");
      }
    } catch (error) {
      console.error("Error uploading image:", error);
      alert("ไม่สามารถอัพโหลดรูปภาพได้");
    }
  };

  return (
    <div className="bg-white max-h-full h-full">
      <div className="flex justify-between items-center bg-white shadow p-4">
        <h1 className="text-xl font-bold text-gray-800">จัดการเมนูอาหาร</h1>
        <div>
          <button
            className="bg-blue-500 text-white px-4 py-2 rounded-lg mr-2 hover:bg-blue-600"
            onClick={() => setShowAddMenuModal(true)}
          >
            เพิ่มเมนูอาหาร
          </button>
        </div>
      </div>

      <div className="p-4">
        <table className="w-full bg-white rounded-lg shadow-lg">
          <thead>
            <tr className="bg-gray-200 text-left">
              <th className="p-4">รหัสสินค้า</th>
              <th className="p-4">รูปสินค้า</th>
              <th className="p-4">ชื่อเมนู</th>
              <th className="p-4">คำอธิบาย</th>
              <th className="p-4">หมวดหมู่</th>
              <th className="p-4">ราคา</th>
              <th className="p-4">ตัวเลือก</th>
              <th className="p-4">อัพโหลดรูปภาพ</th> {/* เพิ่มคอลัมน์สำหรับปุ่มอัพโหลดรูปภาพ */}
            </tr>
          </thead>
          <tbody>
            {menus.map((menu) => (
              <tr key={menu.id} className="border-t">
                <td className="p-4">{menu.ID}</td>
                <td className="p-4">
                  {menu.Image && menu.Image.length > 0 ? (
                    <img
                      src={`data:image/png;base64,${menu.Image}`}
                      alt={menu.Name}
                      className="w-20 h-20 object-cover"
                    />
                  ) : (
                    "ไม่มีรูป"
                  )}
                </td>
                <td className="p-4">
                  <div>ไทย: {menu.Name}</div>
                  <div>อังกฤษ: {menu.NameEn}</div>
                  <div>จีน: {menu.NameCh}</div>
                </td>
                <td className="p-4">
                  <div>ไทย: {menu.Description}</div>
                  <div>อังกฤษ: {menu.DescriptionEn}</div>
                  <div>จีน: {menu.DescriptionCh}</div>
                </td>
                <td className="p-4">
                  {menu.CategoryID ? (
                    <>
                      <div>{menu.Category.Name}</div>
                    </>
                  ) : (
                    "ไม่มีหมวดหมู่"
                  )}
                </td>
                <td className="p-4">{menu.Price}</td>
                <td className="p-4">
                  {menu.OptionGroups?.length > 0 ? (
                    <ul className="list-disc ml-4">
                      {menu.OptionGroups.map((group) => (
                        <li key={group.ID}>
                          <strong>{group.Name}</strong>
                          <ul className="list-circle ml-4">
                            {group.options.map((option) => (
                              <li key={option.ID}>
                                <span>
                                  {option.Name} / {option.NameEn} / {option.NameCh}
                                </span>{" "}
                                - {option.Price} บาท
                              </li>
                            ))}
                          </ul>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    "ไม่มีตัวเลือก"
                  )}
                </td>
                <td className="p-4">
                  <button
                    className="bg-green-500 text-white px-4 py-2 rounded-lg hover:bg-green-600"
                    onClick={() => {
                      setCurrentMenuId(menu.ID); // เก็บ ID เมนูที่ต้องการอัพเดท
                      setShowUploadModal(true); // เปิด popup
                    }}
                  >
                    อัพโหลดรูปภาพ
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Modal สำหรับอัพโหลดรูปภาพ */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center z-50">
          <div className="bg-white p-6 rounded-lg w-96">
            <div className="flex justify-between items-center">
              <h2 className="text-lg font-bold">อัพโหลดรูปภาพ</h2>
              <button
                className="text-red-500"
                onClick={() => setShowUploadModal(false)} // ปิด modal
              >
                &times;
              </button>
            </div>
            <div className="mt-4">
              <input
                type="file"
                id="imageInput"
                onChange={handleFileChange}
                className="border p-2 w-full"
              />
            </div>
            {selectedImage && (
              <div className="mt-4">
                <img
                  src={selectedImage}
                  alt="Preview"
                  className="w-32 h-32 object-cover"
                />
              </div>
            )}
            <div className="mt-4 flex justify-between">
              <button
                className="bg-blue-500 text-white px-4 py-2 rounded-lg"
                onClick={handleUploadImage}
              >
                ยืนยัน
              </button>
              <button
                className="bg-gray-300 text-gray-700 px-4 py-2 rounded-lg"
                onClick={() => setShowUploadModal(false)} // ปิด modal
              >
                ยกเลิก
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal สำหรับเพิ่มเมนูอาหาร */}
      {showAddMenuModal && (
        <AddMenuModal
          onClose={() => setShowAddMenuModal(false)}
          onMenuAdded={() => fetchMenus("getAll")}
        />
      )}
    </div>
  );
};

// Modal สำหรับเพิ่มเมนูอาหาร
const AddMenuModal = ({ onClose, onMenuAdded }) => {
  const [menuName, setMenuName] = useState("");
  const [menuNameEn, setMenuNameEn] = useState("");
  const [menuNameCh, setMenuNameCh] = useState("");
  const [price, setPrice] = useState("");
  const [description, setDescription] = useState("");
  const [descriptionEn, setDescriptionEn] = useState("");
  const [descriptionCh, setDescriptionCh] = useState("");
  const [category, setCategory] = useState("");
  const [categories, setCategories] = useState([]);
  const [options, setOptions] = useState([]);
  const [image, setImage] = useState(null);  // เก็บไฟล์รูปภาพที่เลือก

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        const response = await axios.get("http://127.0.0.1:8080/api/categories");
        if (response.status === 200) {
          setCategories(response.data); // เก็บข้อมูลหมวดหมู่
        }
      } catch (error) {
        console.error("Error fetching categories:", error);
      }
    };
    fetchCategories();
  }, []);

  const handleAddOption = () => {
    setOptions([
      ...options,
      { name: "", name_en: "", name_ch: "", price: 0, is_required: true, MaxSelections: 0 },
    ]);
  };

  const handleOptionChange = (index, field, value) => {
    const newOptions = [...options];
    newOptions[index][field] = value;
    setOptions(newOptions);
  };

  const handleRemoveOption = (index) => {
    setOptions(options.filter((_, i) => i !== index));
  };

  const handleAddMenu = async (e) => {
    e.preventDefault();
  
    try {
      // ถ้ามีรูปภาพให้แปลงไฟล์เป็น Base64
      let base64Image = null;
      if (image) {
        base64Image = await new Promise((resolve) => {
          const reader = new FileReader();
          reader.onloadend = () => resolve(reader.result.split(",")[1]); // แปลงรูปภาพเป็น Base64
          reader.readAsDataURL(image);
        });
      }
  
      // ตรวจสอบว่า category ถูกเลือกหรือไม่
      if (!category) {
        alert("กรุณาเลือกหมวดหมู่ของเมนู");
        return;
      }
  
      // ตรวจสอบว่า options มีค่า
      const validOptions = options.map((group) => ({
        MaxSelections: group.MaxSelections || 0, // ค่าเริ่มต้นถ้าไม่มี
        is_required: group.is_required || false,
        name: group.name || "",
        name_en: group.nameEn || "",
        name_ch: group.nameCh || "",
        options: Array.isArray(group.options)
          ? group.options.map((option) => ({
              name: option.name || "",
              name_en: option.nameEn || "",
              name_ch: option.nameCh || "",
              price: option.price || 0,
            }))
          : [],
      }));
  
      // สร้างข้อมูลที่จะส่งไป
      const menuData = {
        menu_item: {
          name: menuName,
          name_en: menuNameEn,
          name_ch: menuNameCh,
          price: parseFloat(price) || 0, // แปลงเป็น float
          description,
          description_en: descriptionEn,
          description_ch: descriptionCh,
          category_id: parseInt(category, 10), // แปลงเป็น int
          image: base64Image ? [base64Image] : null, // ส่ง null ถ้าไม่มีรูปภาพ
        },
        option_groups: validOptions,
      };
  
      // Log ข้อมูลก่อนส่ง
      console.log("menuData being sent:", menuData);
  
      const response = await axios.post(API_BASE_URL, menuData, {
        headers: { "Content-Type": "application/json" },
      });
  
      if (response.status === 200) {
        alert("เพิ่มเมนูสำเร็จ");
        onMenuAdded(); // รีเฟรชข้อมูล
        onClose(); // ปิด Modal
      }
    } catch (error) {
      console.error("Error adding menu:", error);
      const errorMessage = error.response?.data || "เกิดข้อผิดพลาดในการเพิ่มเมนู";
      alert(errorMessage);
    }
  };

  return (
    <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center z-50">
      <div className="bg-white p-6 rounded-lg w-6/12 ml-12 h-5/6 overflow-y-auto">
        <h2 className="text-xl font-bold mb-4">เพิ่มเมนูอาหาร</h2>
        <form onSubmit={handleAddMenu}>
          {/* ชื่อเมนู */}
          <label className="block mb-2">
            ชื่อเมนู (ไทย)
            <input
              type="text"
              value={menuName}
              onChange={(e) => setMenuName(e.target.value)}
              className="w-full p-2 border rounded"
              required
            />
          </label>

          {/* ชื่อเมนู (อังกฤษ) */}
          <label className="block mb-2">
            ชื่อเมนู (อังกฤษ)
            <input
              type="text"
              value={menuNameEn}
              onChange={(e) => setMenuNameEn(e.target.value)}
              className="w-full p-2 border rounded"
            />
          </label>

          {/* ชื่อเมนู (จีน) */}
          <label className="block mb-2">
            ชื่อเมนู (จีน)
            <input
              type="text"
              value={menuNameCh}
              onChange={(e) => setMenuNameCh(e.target.value)}
              className="w-full p-2 border rounded"
            />
          </label>

          {/* ราคา */}
          <label className="block mb-2">ราคา</label>
          <input
            type="number"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            className="w-full p-2 border rounded"
          />

          {/* คำอธิบาย */}
          <label className="block mb-2">
            คำอธิบาย (ไทย)
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full p-2 border rounded"
            />
          </label>

          {/* คำอธิบาย (อังกฤษ) */}
          <label className="block mb-2">
            คำอธิบาย (อังกฤษ)
            <textarea
              value={descriptionEn}
              onChange={(e) => setDescriptionEn(e.target.value)}
              className="w-full p-2 border rounded"
            />
          </label>

          {/* คำอธิบาย (จีน) */}
          <label className="block mb-2">
            คำอธิบาย (จีน)
            <textarea
              value={descriptionCh}
              onChange={(e) => setDescriptionCh(e.target.value)}
              className="w-full p-2 border rounded"
            />
          </label>

          {/* หมวดหมู่ */}
          <label className="block mb-2">
            หมวดหมู่
            <select
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              className="w-full p-2 border rounded"
            >
              <option value="">เลือกหมวดหมู่</option>
              {categories.length > 0 ? (
                categories.map((category) => (
                  <option key={category.ID} value={category.ID}>
                    {category.Name}
                  </option>
                ))
              ) : (
                <option disabled>ไม่พบหมวดหมู่</option>
              )}
            </select>
          </label>

          {/* ตัวเลือก */}
          <div className="mb-4">
            <button
              type="button"
              onClick={handleAddOption}
              className="bg-green-500 text-white px-4 py-2 rounded mb-2"
            >
              เพิ่มตัวเลือก
            </button>
            {options.map((option, index) => (
              <div key={index} className="mb-4">
                <label className="block mb-2">ตัวเลือก {index + 1}</label>
                <input
                  type="text"
                  placeholder="ชื่อตัวเลือก"
                  value={option.name}
                  onChange={(e) =>
                    handleOptionChange(index, "name", e.target.value)
                  }
                  className="w-full p-2 border rounded mb-2"
                />
                <input
                  type="text"
                  placeholder="ชื่อ (อังกฤษ)"
                  value={option.nameEn}
                  onChange={(e) =>
                    handleOptionChange(index, "nameEn", e.target.value)
                  }
                  className="w-full p-2 border rounded mb-2"
                />
                <input
                  type="text"
                  placeholder="ชื่อ (จีน)"
                  value={option.nameCh}
                  onChange={(e) =>
                    handleOptionChange(index, "nameCh", e.target.value)
                  }
                  className="w-full p-2 border rounded mb-2"
                />
                <input
                  type="number"
                  placeholder="ราคา"
                  value={option.price}
                  onChange={(e) =>
                    handleOptionChange(index, "price", e.target.value)
                  }
                  className="w-full p-2 border rounded mb-2"
                />
                <button
                  type="button"
                  onClick={() => handleRemoveOption(index)}
                  className="text-red-500"
                >
                  ลบตัวเลือก
                </button>
              </div>
            ))}
          </div>

          {/* รูปภาพ */}
          {/* <div className="mb-4">
            <label className="block mb-2">เลือกรูปภาพ</label>
            <input
              type="file"
              accept="image/*"
              onChange={(e) => setImage(e.target.files[0])}
              className="p-2"
            />
          </div> */}

          {/* ปุ่มเพิ่มเมนู */}
          <button
            type="submit"
            className="w-full bg-blue-500 text-white py-2 rounded"
          >
            เพิ่มเมนู
          </button>
        </form>

        <button
          onClick={onClose}
          className="absolute top-2 right-2 text-gray-700"
        >
          ✖️
        </button>
      </div>
    </div>
  );
};


export default MenuManagement;
