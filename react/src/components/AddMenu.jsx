import React, { useState, useEffect } from "react";
import { Image, Plus, Search, Filter, X, User, Edit, Trash2 } from 'lucide-react';
import axios from "axios";
import MenuTable from "./MenuTable"; // นำเข้า MenuTable มาใช้

const API_BASE_URL = "http://127.0.0.1:8080/api/menu"; // กำหนด URL ของ API

const MenuManagement = () => {
  const [menus, setMenus] = useState([]);  // menus เริ่มต้นเป็นอาร์เรย์
  const [categories, setCategories] = useState([]);  // categories เริ่มต้นเป็นอาร์เรย์
  const [showAddMenuModal, setShowAddMenuModal] = useState(false);
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [showUploadModal, setShowUploadModal] = useState(false); // state สำหรับแสดง popup upload รูป
  const [selectedImage, setSelectedImage] = useState(null); // state สำหรับเก็บรูปที่เลือก
  const [currentMenuId, setCurrentMenuId] = useState(null); // state สำหรับเก็บ ID เมนูที่ต้องการอัพเดทรูป
  const [loading, setLoading] = useState(false); // state สำหรับโหลด
  const [errorMessage, setErrorMessage] = useState(""); // state สำหรับข้อผิดพลาด
  const [groupOptions, setGroupOptions] = useState([]);
  const [activeTab, setActiveTab] = useState('menu');// state สลับการแสดงตาราง
  const [menuOptions, setMenuOptions] = useState([]);
  const [searchTerm, setSearchTerm] = useState(""); // สำหรับค้นหาชื่อเมนู
  const [categoryFilter, setCategoryFilter] = useState(""); // สำหรับกรองหมวดหมู่
  const [optionGroupFilter, setOptionGroupFilter] = useState(""); // สำหรับกรองกลุ่มตัวเลือก
  const [filteredMenus, setFilteredMenus] = useState(menus); // รายการเมนูที่กรองแล้ว

  // ฟังก์ชันกรองข้อมูล
  useEffect(() => {
    let result = menus;

    // กรองตามคำค้นหา
    if (searchTerm) {
      result = result.filter((menu) =>
        [menu.Name, menu.NameEn, menu.NameCh, menu.Price.toString()].some((name) =>
          name.toLowerCase().includes(searchTerm.toLowerCase())
        )
      );
    }

    // กรองตามหมวดหมู่
    if (categoryFilter) {
      result = result.filter((menu) => menu.CategoryID === parseInt(categoryFilter));
    }

    // กรองตามกลุ่มตัวเลือก
    if (optionGroupFilter) {
      result = result.filter((menu) =>
        menu.OptionGroups.some((group) => group.ID === parseInt(optionGroupFilter))
      );
    }

    setFilteredMenus(result);
  }, [menus, searchTerm, categoryFilter, optionGroupFilter]);

  

  // เมื่อ showUploadModal เปลี่ยนสถานะ จะรีเซ็ตค่า selectedImage และ currentMenuId
  useEffect(() => {
    if (!showUploadModal) {
      setSelectedImage(null); // เคลียร์ภาพเมื่อปิด popup
      setCurrentMenuId(null); // เคลียร์ ID ของเมนูเมื่อปิด popup
    }
  }, [showUploadModal]);

  // ดึงข้อมูลหมวดหมู่
  useEffect(() => {
    fetchCategories();
    fetchMenus("getAll");
  }, []);
  // ฟังก์ชันเปิด/ปิด อาหารหมด
  const toggleMenuStatus = async (menuId) => {
    try {
      const response = await axios.put(`http://localhost:8080/api/menu/status/${menuId}`, null, {
        headers: {
          "Content-Type": "application/json",
        },
      });
  
      if (response.status === 200) {
        alert(`เปลี่ยนสถานะของเมนู ${response.data.Name} เรียบร้อย`);
        fetchMenus(); // โหลดข้อมูลใหม่
      } else {
        throw new Error("เกิดข้อผิดพลาดในการเปลี่ยนสถานะ");
      }
    } catch (error) {
      console.error("Error toggling menu status:", error);
      alert("ไม่สามารถเปลี่ยนสถานะได้: " + error.message);
    }
  };

  // ฟังก์ชันดึงข้อมูลหมวดหมู่ จาก api
  const fetchCategories = async () => {
    try {
      setLoading(true);
      const response = await axios.get("http://127.0.0.1:8080/api/categories");//การดึงข้อมูล categories จาก api เส้นอื่น
      setCategories(response.data);
    } catch (error) {
      setErrorMessage("ไม่สามารถดึงข้อมูลหมวดหมู่ได้");
    } finally {
      setLoading(false);
    }
  };// สิ้นสุดฟังก์ชันดึงข้อมูลหมวดหมู่ จาก api

  // ดึงข้อมูล options
  useEffect(() => {
    if (currentMenuId) {
      fetchGroupOptions(currentMenuId);
    }
  }, [currentMenuId]);

  // ฟังก์ชันดึงข้อมูล options จาก api
  const fetchGroupOptions = async (menuId) => {
    try {
      const response = await axios.get(`${API_BASE_URL}/option-groups/${menuId}`);
      if (response.status === 200) {
        setMenuOptions(response.data); // ตั้งค่าเมนู options จาก API
      }
    } catch (error) {
      console.error("Error fetching group options:", error);
    }
  };// สิ้นสุดฟังก์ชันดึงข้อมูล options จาก api

  // ฟังก์ชันช่วยในการหาชื่อหมวดหมู่จาก categories
  const getCategoryNameById = (categoryId) => {
    const category = categories.find((cat) => cat.ID === categoryId);
    return category ? category.Name : "ไม่พบหมวดหมู่";
  };// สิ้นสุดฟังก์ชันช่วยในการหาชื่อหมวดหมู่จาก categories

  // ฟังก์ชันดึงข้อมูลเมนู จาก api
  const fetchMenus = async (action, id = null) => {
    try {
      let url = `${API_BASE_URL}?action=${action}`;
      if (action === "getByID" && id) {
        url += `&id=${id}`;
      }

      const response = await axios.get(url);

      if (response.status === 200 && Array.isArray(response.data)) {
        const sortedMenus = response.data.sort((a, b) => a.ID - b.ID);
        setMenus(sortedMenus); // ตั้งค่า menus ให้เป็นข้อมูลที่เรียงลำดับแล้ว
      } else {
        console.error("Data is not an array", response.data);
      }
    } catch (error) {
      console.error("Error fetching menus:", error);
    }
  };// สิ้นสุดฟังก์ชันดึงข้อมูลเมนู จาก api

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


  // Object ของฟังก์ชันแก้ไขข้อมูลเมนู
  const [menuDetails, setMenuDetails] = useState({
    ID: '',
    name: '',
    nameEn: '',
    nameCh: '',
    price: 0,
    description: '',
    descriptionEn: '',
    descriptionCh: '',
    categoryId: 0,
    optionGroups: [{
      name: "",
      nameEn: "",
      nameCh: "",
      MaxSelections: 1,
      isRequired: false,
      options: [
        {
          name: "",
          nameEn: "",
          nameCh: "",
          price: 0,
        },
      ],
    },

    ], // ใช้ array ว่างเป็นค่าเริ่มต้น
  });

  const [showEditMenuModal, setShowEditMenuModal] = useState(false);

  const handleEditMenu = async () => {
    try {
      // ตรวจสอบว่า menuDetails.ID มีค่าอยู่หรือไม่
      if (!menuDetails.ID) {
        throw new Error("ไม่พบ ID ของเมนูในการอัปเดต");
      }

      // 1. อัปเดตข้อมูลเมนูหลัก
      const menuPayload = {
        category_id: menuDetails.categoryId, // หมวดหมู่เป็นตัวเลข
        description: menuDetails.description.trim(),
        description_ch: menuDetails.descriptionCh.trim(),
        description_en: menuDetails.descriptionEn.trim(),
        name: menuDetails.name.trim(),
        name_ch: menuDetails.nameCh.trim(),
        name_en: menuDetails.nameEn.trim(),
        price: Number(menuDetails.price), // แปลงราคาเป็นตัวเลข
        image: "", // ถ้าไม่มีภาพให้ส่งเป็นค่าว่าง
      };

      const menuResponse = await axios.put(
        `http://localhost:8080/api/menu/${menuDetails.ID}`, // ใช้ menuDetails.ID
        menuPayload, // ส่งข้อมูลที่เตรียมไว้
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );

      // ตรวจสอบผลลัพธ์ของการอัปเดตเมนู
      if (menuResponse.status !== 200) {
        throw new Error("ไม่สามารถอัปเดตข้อมูลเมนูได้");
      }

      // 2. อัปเดตข้อมูล Option Groups
      for (const group of menuDetails.optionGroups) {
        if (!group.ID) {
          throw new Error(`ไม่พบ ID ของ Option Group ${group.name}`);
        }

        const options = Array.isArray(group.options) ? group.options : [];

        const groupPayload = {
          MaxSelections: group.MaxSelections || 1, // จำนวนที่เลือกได้สูงสุด
          is_required: group.isRequired || false,  // การบังคับเลือก
          name: group.name.trim(),
          name_en: group.nameEn.trim(),
          name_ch: group.nameCh.trim(),
          options: options.map((option) => ({
            name: option.name.trim(),
            name_en: option.nameEn.trim(),
            name_ch: option.nameCh.trim(),
            price: Number(option.price),
          })),
        };

        const groupResponse = await axios.put(
          `http://localhost:8080/api/menu/option-groups/${group.ID}`, // ใช้ group.ID
          groupPayload, // ส่งข้อมูลที่เตรียมไว้
          {
            headers: {
              "Content-Type": "application/json",
            },
          }
        );

        if (groupResponse.status !== 200) {
          throw new Error(`ไม่สามารถอัปเดตข้อมูล Option Group ${group.ID} ได้`);
        }
      }

      // 3. อัปเดตข้อมูล Options ภายในเมนู (ถ้ามีการแก้ไข)
      for (const group of menuDetails.optionGroups) {
        const options = Array.isArray(group.options) ? group.options : [];

        for (const option of options) {
          if (option.ID) {
            try {
              const optionPayload = {
                name: option.name.trim(),
                name_en: option.nameEn.trim(),
                name_ch: option.nameCh.trim(),
                price: Number(option.price), // ราคาแปลงเป็นตัวเลข
              };

              // อัปเดตข้อมูล Option โดยใช้ option.ID
              const optionResponse = await axios.put(
                `http://localhost:8080/api/menu/options/${option.ID}`, // ใช้ option.ID
                optionPayload,
                {
                  headers: {
                    "Content-Type": "application/json",
                  },
                }
              );

              // ตรวจสอบผลลัพธ์
              if (optionResponse.status !== 200) {
                throw new Error(`ไม่สามารถอัปเดตข้อมูล Option ${option.ID} ได้`);
              }
            } catch (error) {
              console.error(`Error updating option ID ${option.ID}:`, error);
              throw new Error(`เกิดข้อผิดพลาดในการอัปเดต Option ID ${option.ID}`);
            }
          }
        }
      }

      // 4. แจ้งความสำเร็จ
      alert("อัปเดตข้อมูลเมนูสำเร็จ");
      fetchMenus("getAll"); // โหลดข้อมูลใหม่
      setShowEditMenuModal(false); // ปิด Modal
    } catch (error) {
      console.error("Error updating menu:", error);
      alert("เกิดข้อผิดพลาดในการอัปเดตข้อมูลเมนู: " + error.message);
    }
  };


  const [isModalOpen, setIsModalOpen] = useState(false); // สถานะการเปิด/ปิด Modal
  const [updatedData, setUpdatedData] = useState({
    name: "",
    name_ch: "",
    name_en: "",
    price: 0,
  }); // ข้อมูลที่ผู้ใช้กรอกเพื่ออัพเดท

  // ฟังก์ชันเปิด Modal และตั้งค่า Option ที่จะถูกแก้ไข
  const handleEditClick = (option) => {
    setSelectedOption(option);
    setUpdatedData({
      name: option.Name,
      name_ch: option.NameCh,
      name_en: option.NameEn,
      price: option.Price,
    });
    setIsModalOpen(true);
  };

  // ฟังก์ชันปิด Modal
  const handleCloseModal = () => {
    setIsModalOpen(false);
    setSelectedOption(null);
  };

  // ฟังก์ชันสำหรับการอัพเดท Option ผ่าน API
  const updateOption = async () => {
    try {
      const response = await axios.put(
        `http://localhost:8080/api/menu/${selectedOption.GroupID}/options/${selectedOption.ID}`,
        updatedData,
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
      if (response.status === 200) {
        alert("Option updated successfully!");
        // รีเฟรชข้อมูลในตาราง
        handleCloseModal();
      } else {
        alert("Failed to update Option");
      }
    } catch (error) {
      console.error("Error updating option:", error);
      alert("Error updating Option: " + (error.response?.data?.error || error.message));
    }
  };

  // ฟังก์ชันสำหรับการจัดการการเปลี่ยนแปลงในฟอร์ม
  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setUpdatedData((prevData) => ({
      ...prevData,
      [name]: value,
    }));
  };
  //ฟังก์ชันการลบเมนู
  const deleteMenu = async (id) => {
    const confirmDelete = window.confirm("คุณต้องการลบเมนูนี้หรือไม่?");
    if (!confirmDelete) return;
  
    try {
      const response = await axios.delete(`http://localhost:8080/api/menu/${id}`);
      if (response.status === 200) {
        alert("ลบเมนูสำเร็จ");
        // อัปเดตรายการเมนูหลังจากลบ
        setMenus((prevMenus) => prevMenus.filter((menu) => menu.ID !== id));
      } else {
        throw new Error("ลบเมนูไม่สำเร็จ");
      }
    } catch (error) {
      console.error("Error deleting menu:", error);
      alert("เกิดข้อผิดพลาดในการลบเมนู");
    }
  };


  return (
    <div className="bg-white max-h-full h-full lg:ml-60">
      <div className="flex justify-between items-center bg-white shadow p-4">
        <h1 className="text-xl ml-10 font-bold text-gray-800">จัดการเมนูอาหาร</h1>
        <div>
          <button
            className="bg-blue-500 text-white px-4 py-2 rounded-lg mr-2 hover:bg-blue-600"
            onClick={() => setShowAddMenuModal(true)}
          >
            เพิ่มเมนูอาหาร
          </button>
        </div>
      </div>

      <div className="flex mb-4">

      </div>
      <div className="p-4">
        
        <div className="flex items-center mb-4">
        {/* การค้นหาชื่อเมนู */}
        <input
          type="text"
          className="p-2 border border-gray-300 rounded-lg"
          placeholder="ค้นหาชื่อเมนู ราคา..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />

        {/* ตัวกรองหมวดหมู่ */}
        <select
          className="ml-4 p-2 border border-gray-300 rounded-lg"
          value={categoryFilter}
          onChange={(e) => setCategoryFilter(e.target.value)}
        >
          <option value="">เลือกหมวดหมู่</option>
          {categories.map((category) => (
            <option key={category.ID} value={category.ID}>
              {category.Name}
            </option>
          ))}
        </select>

        {/* ตัวกรองกลุ่มตัวเลือก */}
        <select
          className="ml-4 p-2 border border-gray-300 rounded-lg"
          value={optionGroupFilter}
          onChange={(e) => setOptionGroupFilter(e.target.value)}
        >
          <option value="">ตัวเลือกทั้งหมด</option>
          {menus.flatMap((menu) =>
            menu.OptionGroups.map((group) => (
              <option key={group.ID} value={group.ID}>
                {group.Name}
              </option>
            ))
          )}
        </select>
      </div>

      {/* ปุ่มเปลี่ยนตาราง */}
      <button
          className={`px-4 py-2 rounded-t-lg ${activeTab === 'menu' ? 'bg-blue-500 text-white' : 'bg-gray-200 text-gray-700 hover:bg-gray-300'}`}
          onClick={() => setActiveTab('menu')}
        >
          ข้อมูลสินค้า
        </button>
        <button
          className={`px-4 py-2 rounded-t-lg ${activeTab === 'options' ? 'bg-blue-500 text-white' : 'bg-gray-200 text-gray-700 hover:bg-gray-300'}`}
          onClick={() => setActiveTab('options')}
        >
          ข้อมูลตัวเลือกของสินค้า
        </button>
          {/* สิ้นสุดปุ่มเปลี่ยนตาราง */}

        {/* ตารางแสดงเมนู */}
        {activeTab === 'menu' && (
          
          <table className="w-full max-x-full bg-white rounded-3xl shadow-lg border border-gray-300">
        <thead>
          <tr className="bg-blue-500 text-left text-white">
            <th className="p-4">รหัสสินค้า</th>
            <th className="p-4">รูปสินค้า</th>
            <th className="p-4">ชื่อเมนู</th>
            <th className="p-4 w-3/12">คำอธิบาย</th>
            <th className="p-4">หมวดหมู่</th>
            <th className="p-4">ราคา (THB)</th>
            <th className="p-4 text-center">ตัวเลือก</th>
            <th className="p-4 text-center">สถานะ</th>
            <th className="p-4 text-center">Actions</th>
          </tr>
        </thead>
        <tbody>
          {Array.isArray(filteredMenus) && filteredMenus.length > 0 ? (
            filteredMenus.map((menu) => (
              <tr key={menu.ID} className="border-t">
                <td className="text-center">{menu.ID}</td>
                <td className="">
                  {menu.Image && menu.Image.length > 0 ? (
                    <img
                      src={`data:image/png;base64,${menu.Image}`}
                      alt={menu.Name}
                      className="w-40 h-25 object-cover"
                    />
                  ) : (
                    "ไม่มีรูป"
                  )}
                </td>
                <td className="p-2">
                  <div>TH: {menu.Name} </div>
                  <div>EN: {menu.NameEn}</div>
                  <div>CH: {menu.NameCh}</div>
                </td>
                <td className="p-2">
                  <div className="mb-2">TH: {menu.Description}</div>
                  <div className="mb-2">EN: {menu.DescriptionEn}</div>
                  <div>CH: {menu.DescriptionCh}</div>
                </td>
                <td className="p-2">
                  {menu.CategoryID ? (
                    <div>{getCategoryNameById(menu.CategoryID)}</div>
                  ) : (
                    "ไม่พบหมวดหมู่"
                  )}
                </td>
                <td className="p-2 text-center">{menu.Price}</td>
                <td className="p-2">
                  {/* แสดงตัวเลือก */}
                  {Array.isArray(menu.OptionGroups) && menu.OptionGroups.length > 0 ? (
                    <ul className="ml-4">
                      {menu.OptionGroups.map((group) => (
                        <li className="border-b"key={group.ID}>
                          <div>กลุ่ม : {group.ID}</div>
                          <span>TH : {group.Name} | EN : {group.NameEn} | CH : {group.NameCh}</span>
                          <ul className="list-circle ml-4">
                            {group.Options && group.Options.length > 0 ? (
                              group.Options.map((option) => (
                                <li key={option.ID}>
                                  <span>{option.Name} | {option.NameEn} | {option.NameCh} - {option.Price} THB</span> 
                                </li>
                              ))
                            ) : (
                              <span className="text-red-500 text-center">ไม่มีตัวเลือก</span>
                            )}
                          </ul>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <span className="text-red-500 text-center">ไม่มีตัวเลือก</span>
                  )}
                    </td>
                    <td className="text-center">
                    <button
  className={`p-2 px-4 rounded-lg ${menu.Is_available ? "bg-green-400  hover:bg-green-500" : "bg-red-400 hover:bg-red-500"} text-white`}
  onClick={() => toggleMenuStatus(menu.ID).then(updatedMenu => {
    // ใช้ข้อมูลเมนูที่อัปเดตเพื่อรีเฟรช UI
    if (updatedMenu) {
      const updatedMenus = menus.map(m => m.ID === updatedMenu.ID ? updatedMenu : m);
      setMenus(updatedMenus); // อัปเดต state
    }
  })}
>
  {menu.Is_available ? "พร้อม" : "หมด"}
</button>
                    </td>
                    <td className=" p-2 justify-center items-center">
                      <div className="flex">
                        <button
                          className="bg-yellow-500 text-white mx-2 px-4 py-2 rounded-lg hover:bg-yellow-600"
                          onClick={() => {
                            setCurrentMenuId(menu.ID); // ตั้งค่า ID ของเมนูที่ต้องการแก้ไข
                            setCurrentMenuId(menu.ID);
                            setMenuDetails({ //ข้อมูลของเมนูทั้งหมดที่ถูกส่งไปให้แก้ไข
                              ID: menu.ID || "",
                              name: menu.Name || "",
                              nameEn: menu.NameEn || "",
                              nameCh: menu.NameCh || "",
                              price: menu.Price || 0,
                              description: menu.Description || "",
                              descriptionEn: menu.DescriptionEn || "",
                              descriptionCh: menu.DescriptionCh || "",
                              categoryId: menu.CategoryID || "",
                              optionGroups: Array.isArray(menu.OptionGroups) ? menu.OptionGroups.map(group => ({
                                ID: group.ID || "",
                                name: group.Name || "",
                                nameEn: group.NameEn || "",
                                nameCh: group.NameCh || "",
                                MaxSelections: group.MaxSelections || 1,
                                isRequired: group.IsRequired || false,
                                options: Array.isArray(group.Options) ? group.Options.map(option => ({
                                  ID: option.ID || "",
                                  name: option.Name || "",
                                  nameEn: option.NameEn || "",
                                  nameCh: option.NameCh || "",
                                  price: option.Price || 0,
                                })) : [],
                              })) : [],
                            });
                            setShowEditMenuModal(true);
                          }}
                        >
                          <div className="flex"><Edit />edit</div>
                        </button>

                        <button
                          className="bg-green-500 text-white px-2 py-2 rounded-lg hover:bg-green-600 "
                          onClick={() => {
                            setCurrentMenuId(menu.ID);
                            setShowUploadModal(true);
                          }}
                        >

                          <div className="flex"><Image /> image</div>
                        </button>
                        <button
          className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600"
          onClick={() => deleteMenu(menu.ID)}
        >
          ลบ
        </button>
                      </div>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan="8" className="text-center p-4">ไม่พบเมนู</td>
                </tr>
              )}
            </tbody>
          </table>
        )}
        {/* สิ้นสุดตารางแสดงข้อมูลสินค้า */}

        {/* ตารางแสดง options */}
        {activeTab === 'options' && (
          <table className="w-full bg-white  rounded-lg border border-gray-300 shadow-lg">
            <thead>
              <tr className="bg-blue-500 text-left  text-white">
                <th className="border-b px-4 py-2">Group ID</th>
                <th className="border-b px-4 py-2">Group Name</th>
                <th className="border-b px-4 py-2">Max Selections</th>
                <th className="border-b px-4 py-2">Required</th>
                <th className="border-b px-4 py-2">Option ID</th>
                <th className="border-b px-4 py-2">Option Name</th>
                <th className="border-b px-4 py-2">Price (THB)</th>
                <th className="border-b px-4 py-2">Actions</th>
              </tr>
            </thead>
            <tbody>
              {menuOptions.length > 0 ? (
                menuOptions.map((group) => (
                  group.Options.map((option) => (
                    <tr key={option.ID}>
                      <td className="border-b px-4 py-2">{group.ID}</td>
                      <td className="border-b px-4 py-2">ไทย: {group.Name}<br />อังกฤษ: {group.NameEn}<br />จีน: {group.NameCh}</td>
                      <td className="border-b px-4 py-2">{group.MaxSelections}</td>
                      <td className="border-b px-4 py-2">{group.IsRequired ? 'Yes' : 'No'}</td>
                      <td className="border-b px-4 py-2">{option.ID}</td>
                      <td className="border-b px-4 py-2">ไทย: {option.Name}<br />อังกฤษ: {option.NameEn}<br />จีน: {option.NameCh}</td>
                      <td className="border-b px-4 py-2">{option.Price}</td>
                      <td className="border-b px-4 py-2">
                        <button onClick={() => handleEditClick(option)}>
                          <Edit />
                        </button>

                      </td>
                    </tr>
                  ))
                ))
              ) : (
                <tr>
                  <td colSpan="7" className="text-center p-4">ไม่พบตัวเลือก</td>
                </tr>
              )}
            </tbody>
          </table>

        )}
        {/* สิ้นสุดตารางแสดง options */}



      </div>

      {/* Popup สำหรับการแก้ไขข้อมูล Option */}
      {isModalOpen && (
        <div className="fixed inset-0 flex justify-center items-center bg-gray-800 bg-opacity-50">
          <div className="bg-white p-4 rounded-lg shadow-lg w-96">
            <h2 className="text-xl mb-4">Edit Option</h2>
            <form>
              <div className="mb-2">
                <label htmlFor="name" className="block">Option Name (TH)</label>
                <input
                  type="text"
                  id="name"
                  name="name"
                  value={updatedData.name}
                  onChange={handleInputChange}
                  className="w-full p-2 border rounded"
                />
              </div>
              <div className="mb-2">
                <label htmlFor="name_ch" className="block">Option Name (CH)</label>
                <input
                  type="text"
                  id="name_ch"
                  name="name_ch"
                  value={updatedData.name_ch}
                  onChange={handleInputChange}
                  className="w-full p-2 border rounded"
                />
              </div>
              <div className="mb-2">
                <label htmlFor="name_en" className="block">Option Name (EN)</label>
                <input
                  type="text"
                  id="name_en"
                  name="name_en"
                  value={updatedData.name_en}
                  onChange={handleInputChange}
                  className="w-full p-2 border rounded"
                />
              </div>
              <div className="mb-2">
                <label htmlFor="price" className="block">Price (THB)</label>
                <input
                  type="number"
                  id="price"
                  name="price"
                  value={updatedData.price}
                  onChange={handleInputChange}
                  className="w-full p-2 border rounded"
                />
              </div>
              <div className="flex justify-between mt-4">
                <button
                  type="button"
                  onClick={handleCloseModal}
                  className="bg-gray-500 text-white px-4 py-2 rounded"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={updateOption}
                  className="bg-blue-500 text-white px-4 py-2 rounded"
                >
                  Save Changes
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
      {/* สิ้นสุด Popup สำหรับการแก้ไขข้อมูล Option */}

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
                <X />
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
      {/* สิ้นสุด Modal สำหรับอัพโหลดรูปภาพ */}

      {/* Modal สำหรับเพิ่มเมนูอาหาร */}
      {showAddMenuModal && (
        <AddMenuModal
          onClose={() => setShowAddMenuModal(false)}
          onMenuAdded={() => fetchMenus("getAll")}
        />
      )}

      {/* Modal สำหรับแก้ไขเมนูอาหาร */}
      {showEditMenuModal && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-50 flex justify-center items-center z-50">
          <div className="bg-white p-6 rounded-lg w-6/12 ml-12 h-5/6 overflow-y-auto">
            <div className="flex justify-between items-center">
              <h2 className="text-2xl font-bold">แก้ไขข้อมูลเมนู</h2>
              <button
                className="text-red-500"
                onClick={() => setShowEditMenuModal(false)} // ปิด Modal
              >
                <X />
              </button>
            </div>

            {/* ฟอร์มแก้ไขเมนู */}
              
              <div className="mb-4 text-xl"><label>รหัสสินค้า : {menuDetails.ID}</label></div>
              <label className="text-lg">ชื่อเมนูอาหาร</label>
              <div className="flex px-4 border">
              <div className="my-5 mr-2">
              <label>ชื่อเมนู (ไทย)</label>
              <input
                type="text"
                value={menuDetails.name}
                onChange={(e) => setMenuDetails({ ...menuDetails, name: e.target.value })}
                className="border p-2 w-full mt-2"
              />
            </div>
            <div className="mt-4 mx-2">
              <label className="text-lg">ชื่อเมนู (อังกฤษ)</label>
              <input
                type="text"
                value={menuDetails.nameEn}
                onChange={(e) => setMenuDetails({ ...menuDetails, nameEn: e.target.value })}
                className="border p-2 w-full mt-2"
              />
            </div>
            <div className="mt-4 mx-2">
              <label className="text-lg">ชื่อเมนู (จีน)</label>
              <input
                type="text"
                value={menuDetails.nameCh}
                onChange={(e) => setMenuDetails({ ...menuDetails, nameCh: e.target.value })}
                className="border p-2 w-full mt-2"
              />
            </div>
            </div>
            <div className="mt-4">
              <label className="text-lg">ราคา</label>
              <input
                type="number"
                value={menuDetails.price}
                onChange={(e) => setMenuDetails({ ...menuDetails, price: e.target.value })}
                className="border p-2 w-full "
              />
            </div>
            <div className="mt-4 text-lg">คำอธิบายของอาหาร</div>
            <div className="border p-4">
            <div className="mt-4">
              <label>คำอธิบาย (ไทย)</label>
              <textarea
                value={menuDetails.description}
                onChange={(e) => setMenuDetails({ ...menuDetails, description: e.target.value })}
                className="border p-2 w-full mt-2"
              />
            </div>
            <div className="mt-4">
              <label>คำอธิบาย (อังกฤษ)</label>
              <textarea
                value={menuDetails.descriptionEn}
                onChange={(e) => setMenuDetails({ ...menuDetails, descriptionEn: e.target.value })}
                className="border p-2 w-full mt-2"
              />
            </div>
            <div className="mt-4">
              <label>คำอธิบาย (จีน)</label>
              <textarea
                value={menuDetails.descriptionCh}
                onChange={(e) => setMenuDetails({ ...menuDetails, descriptionCh: e.target.value })}
                className="border p-2 w-full mt-2"
              />
            </div>
            </div>

            {/* หมวดหมู่ */}
            <div className="mt-6">
              <label className="text-lg">หมวดหมู่</label>
              <select
                value={menuDetails.categoryId || ""}
                onChange={(e) =>
                  setMenuDetails({
                    ...menuDetails,
                    categoryId: e.target.value ? Number(e.target.value) : null,
                  })
                }
                className="border p-2 w-full mt-2"
              >
                <option value="">กรุณาเลือกหมวดหมู่</option>
                {categories.map((category) => (
                  <option key={category.ID} value={category.ID}>
                    {category.Name}
                  </option>
                ))}
              </select>
            </div>

            {/* Option Groups */}
            <div className="mt-6 mb-2 text-lg">ข้อมูลกลุ่มตัวเลือกและตัวเลือกของเมนูอาหาร</div>
            {menuDetails.optionGroups.length > 0 ? (
              menuDetails.optionGroups.map((group, groupIndex) => (
                <div key={groupIndex}>
                   
                  <div className="border p-2 pb-8 ">
                    <label>รหัสกลุ่มตัวเลือก : {group.ID}</label>
                  <div className="mt-4">ชื่อกลุ่มตัวเลือกภายในเมนู</div>
                  <div className="flex px-4 border">
                  <div className="mt-4 mb-4">
                    <label>ชื่อกลุ่มตัวเลือก (ไทย)</label>
                    <input
                      type="text"
                      value={group.name}
                      onChange={(e) => {
                        const updatedGroups = [...menuDetails.optionGroups];
                        updatedGroups[groupIndex].name = e.target.value;
                        setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
                      }}
                      className="border p-2 w-full mt-2"
                    />
                  </div>
                  <div className="mt-4 mx-2">
                    <label>ชื่อกลุ่มตัวเลือก (อังกฤษ)</label>
                    <input
                      type="text"
                      value={group.nameEn}
                      onChange={(e) => {
                        const updatedGroups = [...menuDetails.optionGroups];
                        updatedGroups[groupIndex].nameEn = e.target.value;
                        setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
                      }}
                      className="border p-2 w-full mt-2"
                    />
                  </div>
                  <div className="mt-4">
                    <label>ชื่อกลุ่มตัวเลือก (จีน)</label>
                    <input
                      type="text"
                      value={group.nameCh}
                      onChange={(e) => {
                        const updatedGroups = [...menuDetails.optionGroups];
                        updatedGroups[groupIndex].nameCh = e.target.value;
                        setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
                      }}
                      className="border p-2 w-full mt-2"
                    />
                  </div>
                  <div className="mt-4">
                    <label>จำนวนที่เลือกได้สูงสุด</label>
                    <input
  type="number"
  value={group.MaxSelections || ""}
  onChange={(e) => {
    const updatedGroups = [...menuDetails.optionGroups];
    updatedGroups[groupIndex].MaxSelections = Number(e.target.value); // แปลงเป็นตัวเลข
    setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
  }}
  className="border p-2 w-full mt-2"
/>
                  </div>
                  <div className="mt-4">
                    <label>ต้องเลือกหรือไม่</label>
                    <input
                      type="checkbox"
                      checked={group.isRequired}
                      onChange={(e) => {
                        setMenuDetails((prevDetails) => {
                          const updatedGroups = [...prevDetails.optionGroups];
                          updatedGroups[groupIndex] = {
                            ...updatedGroups[groupIndex],
                            isRequired: e.target.checked,
                          };
                          return { ...prevDetails, optionGroups: updatedGroups };
                        });
                      }}
                      className="mt-2 ml-1"
                    />
                  </div>
                  </div>
                 
                  
                  {/* ฟอร์มสำหรับแก้ไข Option ใน Option Group */}
                  {group.options.length > 0 ? (
                    group.options.map((option, optionIndex) => (
                      <div key={optionIndex}>
                        
                        <div className="mt-4">ชื่อตัวเลือกภายในเมนู</div>
                        <div className="flex px-4 border">
                        <div className="my-2 mx-2">
                          <label>ชื่อตัวเลือก (ไทย)</label>
                          <input
                            type="text"
                            value={option.name}
                            onChange={(e) => {
                              const updatedGroups = [...menuDetails.optionGroups];
                              updatedGroups[groupIndex].options[optionIndex].name = e.target.value;
                              setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
                            }}
                            className="border p-2 w-full mt-2"
                          />
                        </div>
                        <div className="mt-2 mr-2">
                          <label>ชื่อตัวเลือก (อังกฤษ)</label>
                          <input
                            type="text"
                            value={option.nameEn}
                            onChange={(e) => {
                              const updatedGroups = [...menuDetails.optionGroups];
                              updatedGroups[groupIndex].options[optionIndex].nameEn = e.target.value;
                              setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
                            }}
                            className="border p-2 w-full mt-2"
                          />
                        </div>
                        <div className="mt-2 mr-2">
                          <label>ชื่อตัวเลือก (จีน)</label>
                          <input
                            type="text"
                            value={option.nameCh}
                            onChange={(e) => {
                              const updatedGroups = [...menuDetails.optionGroups];
                              updatedGroups[groupIndex].options[optionIndex].nameCh = e.target.value;
                              setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
                            }}
                            className="border p-2 w-full mt-2"
                          />
                        </div>
                        <div className="mt-2">
                          <label>ราคา Option</label>
                          <input
                            type="number"
                            value={option.price}
                            onChange={(e) => {
                              const updatedGroups = [...menuDetails.optionGroups];
                              updatedGroups[groupIndex].options[optionIndex].price = e.target.value;
                              setMenuDetails({ ...menuDetails, optionGroups: updatedGroups });
                            }}
                            className="border p-2 w-16 mt-2"
                          />THB
                        </div>
                        </div>
                      </div>
                    ))
                  ) : (
                    <p>ไม่มีตัวเลือกในกลุ่มนี้</p>
                  )}
                 
                </div>
                </div>
              ))
            ) : (
              <p>ยังไม่มีกลุ่มตัวเลือก</p>
            )}

            <div className="mt-4 flex justify-end">
              <button
                className="mx-2 bg-gray-300 text-gray-700 px-4 py-2 rounded-lg"
                onClick={() => setShowEditMenuModal(false)} // ปิด Modal
              >
                ยกเลิก
              </button>
              <button
                className="bg-blue-500 text-white px-4 py-2 rounded-lg"
                onClick={handleEditMenu} // เรียกฟังก์ชันแก้ไขเมนู
              >
                ยืนยัน
              </button>
            </div>
          </div>
        </div>
      )}
      {/* สิ้นสุด Modal สำหรับแก้ไขเมนูอาหาร */}

    </div>
  );
};




{/* ---------------------------------------------------------------------------------------------------------- */ }
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
  const [options, setOptions] = useState([]); // เก็บข้อมูลตัวเลือก
  const [image, setImage] = useState(null); // เก็บไฟล์รูปภาพที่เลือก

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

  // ฟังก์ชันเพิ่มกลุ่มตัวเลือก
  const handleAddOptionGroup = () => {
    setOptions([
      ...options,
      {
        MaxSelections: 1,
        is_required: true,
        name: "",
        name_en: "",
        name_ch: "",
        options: [],
      },
    ]);
  };

  // ฟังก์ชันแก้ไขข้อมูลของกลุ่มตัวเลือก
  const handleOptionGroupChange = (groupIndex, field, value) => {
    const newOptions = [...options];
    newOptions[groupIndex][field] = value;
    setOptions(newOptions);
  };

  // ฟังก์ชันเพิ่มตัวเลือกในกลุ่ม
  const handleAddOption = (groupIndex) => {
    const updatedOptions = [...options];
    updatedOptions[groupIndex].options.push({
      name: "",
      name_en: "",
      name_ch: "",
      price: 0,
    });
    setOptions(updatedOptions);
  };

  // ฟังก์ชันแก้ไขข้อมูลของตัวเลือกในกลุ่ม
  const handleOptionChange = (groupIndex, optionIndex, field, value) => {
    const updatedOptions = [...options];
    updatedOptions[groupIndex].options[optionIndex][field] = value;
    setOptions(updatedOptions);
  };

  // ฟังก์ชันลบตัวเลือกในกลุ่ม
  const handleRemoveOption = (groupIndex, optionIndex) => {
    const updatedOptions = [...options];
    updatedOptions[groupIndex].options.splice(optionIndex, 1);
    setOptions(updatedOptions);
  };

  // ฟังก์ชันลบกลุ่มตัวเลือก
  const handleRemoveOptionGroup = (groupIndex) => {
    const updatedOptions = [...options];
    updatedOptions.splice(groupIndex, 1);
    setOptions(updatedOptions);
  };

  const handleAddMenu = async (e) => {
    e.preventDefault();

    // ตรวจสอบข้อมูลที่จำเป็นก่อนส่ง
    if (!menuName || !category || !price) {
      alert("กรุณากรอกข้อมูลที่จำเป็น");
      return;
    }

    // ตรวจสอบตัวเลือก
    const validOptionGroups = options.map((group) => ({
      MaxSelections: group.MaxSelections || 1,
      is_required: group.is_required ?? true, // ตรวจสอบว่ามีค่าหรือไม่ ถ้าไม่มีใช้ค่า default
      name: group.name || "",
      name_en: group.name_en || "",
      name_ch: group.name_ch || "",
      options: group.options.map((option) => ({
        name: option.name || "",
        name_en: option.name_en || "",
        name_ch: option.name_ch || "",
        price: parseFloat(option.price) || 0, // ราคาต้องเป็นตัวเลข
      })),
    }));

    const menuData = {
      menu_item: {
        name: menuName,
        name_en: menuNameEn,
        name_ch: menuNameCh,
        price: parseFloat(price) || 0,
        description: description || "",
        description_en: descriptionEn || "",
        description_ch: descriptionCh || "",
        category_id: parseInt(category, 10),
        image: [], // แก้ไขให้ image เป็น array ที่ API ยอมรับ
      },
      option_groups: validOptionGroups,
    };

    try {
      const response = await axios.post(
        "http://localhost:8080/api/menu",
        menuData,
        {
          headers: { "Content-Type": "application/json" },
        }
      );

      if (response.status === 200) {
        alert("เพิ่มเมนูสำเร็จ");
        onMenuAdded();
        onClose();
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
        <div className="flex justify-between">
          <h2 className="text-xl font-bold mb-4">เพิ่มเมนูอาหาร</h2>
          <button
            onClick={onClose}
            className="  right-2 text-red-500"
          >
            <X />
          </button>
        </div>
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
            onWheel={(e) => e.preventDefault()}  // ป้องกันการเลื่อนเมาส์
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

          {/* กลุ่มตัวเลือก */}
          <div className="mb-4">
            <button type="button" onClick={handleAddOptionGroup}>
              เพิ่มกลุ่มตัวเลือก
            </button>

            {options.map((group, groupIndex) => (
              <div key={groupIndex}>
                {/* ฟิลด์ของกลุ่มตัวเลือก */}
                <input
                  type="text"
                  placeholder="ชื่อกลุ่ม"
                  value={group.name}
                  onChange={(e) =>
                    handleOptionGroupChange(groupIndex, "name", e.target.value)
                  }
                />
                <input
                  type="text"
                  placeholder="ชื่อกลุ่ม (อังกฤษ)"
                  value={group.name_en}
                  onChange={(e) =>
                    handleOptionGroupChange(groupIndex, "name_en", e.target.value)
                  }
                />
                <input
                  type="text"
                  placeholder="ชื่อกลุ่ม (จีน)"
                  value={group.name_ch}
                  onChange={(e) =>
                    handleOptionGroupChange(groupIndex, "name_ch", e.target.value)
                  }
                />
                {/* ตัวเลือกภายในกลุ่ม */}
                <button type="button" onClick={() => handleAddOption(groupIndex)}>
                  เพิ่มตัวเลือกในกลุ่มนี้
                </button>
                {group.options.map((option, optionIndex) => (
                  <div key={optionIndex}>
                    <input
                      type="text"
                      placeholder="ชื่อตัวเลือก"
                      value={option.name}
                      onChange={(e) =>
                        handleOptionChange(groupIndex, optionIndex, "name", e.target.value)
                      }
                    />
                    <input
                      type="text"
                      placeholder="ชื่อ (อังกฤษ)"
                      value={option.name_en}
                      onChange={(e) =>
                        handleOptionChange(groupIndex, optionIndex, "name_en", e.target.value)
                      }
                    />
                    <input
                      type="text"
                      placeholder="ชื่อ (จีน)"
                      value={option.name_ch}
                      onChange={(e) =>
                        handleOptionChange(groupIndex, optionIndex, "name_ch", e.target.value)
                      }
                    />
                    <input
                      type="number"
                      placeholder="ราคา"
                      value={option.price}
                      onChange={(e) =>
                        handleOptionChange(groupIndex, optionIndex, "price", e.target.value)
                      }
                    />
                    <button
                      type="button"
                      onClick={() => handleRemoveOption(groupIndex, optionIndex)}
                    >
                      ลบตัวเลือก
                    </button>
                  </div>
                ))}
                <button
                  type="button"
                  onClick={() => handleRemoveOptionGroup(groupIndex)}
                >
                  ลบกลุ่มตัวเลือก
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
          <div className="flex justify-end">
            <button
              onClick={onClose}
              className="w-1/5 bg-gray-400 hover:bg-gray-500 text-white py-2 rounded mx-2"
            >
              ยกเลิก
            </button>
            <button
              type="submit"
              className="w-1/5 bg-blue-500 hover:bg-blue-700 text-white py-2 rounded"
            >
              เพิ่มเมนู
            </button>
          </div>
        </form>


      </div>
    </div>
  );
};


export default MenuManagement;
