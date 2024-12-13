import React, { useState, useEffect } from "react";
import axios from "axios";

const MenuManagement = () => {
  const [menus, setMenus] = useState([]); // เก็บข้อมูลเมนูทั้งหมด
  const [showAddMenuModal, setShowAddMenuModal] = useState(false); // state สำหรับแสดง/ซ่อน modal เพิ่มเมนู
  const [showAddCategoryModal, setShowAddCategoryModal] = useState(false); // state สำหรับแสดง/ซ่อน modal เพิ่มหมวดหมู่

  // Fetch ข้อมูลเมนูจาก API
  useEffect(() => {
    fetchMenus();
  }, []);

  const fetchMenus = async () => {
    try {
      const response = await axios.get("http://127.0.0.1:8080/getmenu?action=getAll");
      if (response.status === 200) {
        setMenus(response.data);
      }
    } catch (error) {
      console.error("Error fetching menus:", error);
      alert("ไม่สามารถดึงข้อมูลเมนูได้");
    }
  };

  return (
    <div className="bg-white max-h-full h-full">
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
            </tr>
          </thead>
          <tbody>
            {menus.map((menu) => (
              <tr key={menu.id} className="border-t">
                <td className="p-4">{menu.id}</td>
                <td className="p-4">
                  {menu.image && menu.image.length > 0 ? (
                    <img
                      src={`data:image/png;base64,${menu.image}`}
                      alt={menu.name}
                      className="w-20 h-20 object-cover"
                    />
                  ) : (
                    "ไม่มีรูป"
                  )}
                </td>
                <td className="p-4">
                  <div>ไทย: {menu.name}</div>
                  <div>อังกฤษ: {menu.nameEn}</div>
                  <div>จีน: {menu.nameCh}</div>
                </td>
                <td className="p-4">
                  <div>ไทย: {menu.description}</div>
                  <div>อังกฤษ: {menu.descriptionEn}</div>
                  <div>จีน: {menu.descriptionCh}</div>
                </td>
                <td className="p-4">{menu.category?.name || "ไม่มีหมวดหมู่"}</td>
                <td className="p-4">{menu.price}</td>
                <td className="p-4">
                  {menu.optionGroups?.length > 0 ? (
                    <ul className="list-disc ml-4">
                      {menu.optionGroups.map((group) => (
                        <li key={group.id}>
                          <strong>{group.name}</strong>
                          <ul className="list-circle ml-4">
                            {group.options.map((option) => (
                              <li key={option.id}>
                                <span>
                                  {option.name} / {option.nameEn} / {option.nameCh}
                                </span>{" "}
                                - {option.price} บาท
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
  const [menuNameEn, setMenuNameEn] = useState("");
  const [menuNameCh, setMenuNameCh] = useState("");
  const [price, setPrice] = useState("");
  const [description, setDescription] = useState("");
  const [descriptionEn, setDescriptionEn] = useState("");
  const [descriptionCh, setDescriptionCh] = useState("");
  const [category, setCategory] = useState("");
  const [categories, setCategories] = useState([]);
  const [options, setOptions] = useState([]);
  const [image, setImage] = useState(null);

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        const response = await axios.get("http://127.0.0.1:8080/getCategory");
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

  const handleAddOption = () => {
    setOptions([
      ...options,
      { name: "", nameEn: "", nameCh: "", price: 0 },
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

    const formData = new FormData();
    formData.append("name", menuName);
    formData.append("nameEn", menuNameEn);
    formData.append("nameCh", menuNameCh);
    formData.append("price", price);
    formData.append("description", description);
    formData.append("descriptionEn", descriptionEn);
    formData.append("descriptionCh", descriptionCh);
    formData.append("categoryID", category);
    formData.append("options", JSON.stringify(options));
    if (image) formData.append("image", image);

    try {
      const response = await axios.post("http://127.0.0.1:8080/add_menu", formData, {
        headers: { "Content-Type": "multipart/form-data" },
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
  <div className="bg-white rounded-lg p-8 shadow-lg w-7/12 h-[80vh] overflow-y-auto md:ml-32">
    <h2 className="text-xl font-bold mb-4">เพิ่มเมนูอาหาร</h2>
    <form onSubmit={handleAddMenu} className="space-y-4">
      {/* ชื่อเมนูสามภาษา */}
      <h2>ชื่อเมนู</h2>
      <div className="flex">
        <input
          type="text"
          placeholder="ชื่อเมนู (ภาษาไทย)"
          value={menuName}
          onChange={(e) => setMenuName(e.target.value)}
          className="w-full px-4 py-2 border rounded-lg"
          required
        />
        <input
          type="text"
          placeholder="ชื่อเมนู (ภาษาอังกฤษ)"
          value={menuNameEn}
          onChange={(e) => setMenuNameEn(e.target.value)}
          className="w-full px-4 py-2 border rounded-lg"
        />
        <input
          type="text"
          placeholder="ชื่อเมนู (ภาษาจีน)"
          value={menuNameCh}
          onChange={(e) => setMenuNameCh(e.target.value)}
          className="w-full px-4 py-2 border rounded-lg"
        />
      </div>

      {/* คำอธิบายสามภาษา */}
      <textarea
        placeholder="คำบรรยาย (ภาษาไทย)"
        value={description}
        onChange={(e) => setDescription(e.target.value)}
        className="w-full px-4 py-2 border rounded-lg"
      />
      <textarea
        placeholder="คำบรรยาย (ภาษาอังกฤษ)"
        value={descriptionEn}
        onChange={(e) => setDescriptionEn(e.target.value)}
        className="w-full px-4 py-2 border rounded-lg"
      />
      <textarea
        placeholder="คำบรรยาย (ภาษาจีน)"
        value={descriptionCh}
        onChange={(e) => setDescriptionCh(e.target.value)}
        className="w-full px-4 py-2 border rounded-lg"
      />

      {/* ราคาของเมนู */}
  <div>
    <label className="block font-bold mb-2">ราคาเมนู (บาท)</label>
    <input
      type="number"
      placeholder="ระบุราคาเมนู"
      value={price}
      onChange={(e) => setPrice(e.target.value)}
      className="w-full px-4 py-2 border rounded-lg"
      required
    />
  </div>

      {/* หมวดหมู่ */}
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

      {/* ตัวเลือก (Options) */}
      <div>
        <h3 className="font-bold">ตัวเลือก</h3>
        {options.map((option, index) => (
          <div key={index} className="space-y-2 border-b pb-2 mb-2">
            <input
              type="text"
              placeholder="ชื่อ (ภาษาไทย)"
              value={option.name}
              onChange={(e) =>
                handleOptionChange(index, "name", e.target.value)
              }
              className="w-full px-4 py-2 border rounded-lg"
            />
            <input
              type="text"
              placeholder="ชื่อ (ภาษาอังกฤษ)"
              value={option.nameEn}
              onChange={(e) =>
                handleOptionChange(index, "nameEn", e.target.value)
              }
              className="w-full px-4 py-2 border rounded-lg"
            />
            <input
              type="text"
              placeholder="ชื่อ (ภาษาจีน)"
              value={option.nameCh}
              onChange={(e) =>
                handleOptionChange(index, "nameCh", e.target.value)
              }
              className="w-full px-4 py-2 border rounded-lg"
            />
            <input
              type="number"
              placeholder="ราคา"
              value={option.price}
              onChange={(e) =>
                handleOptionChange(index, "price", e.target.value)
              }
              className="w-full px-4 py-2 border rounded-lg"
            />
            <button
              type="button"
              onClick={() => handleRemoveOption(index)}
              className="text-red-500"
            >
              ลบ
            </button>
          </div>
        ))}
        <button
          type="button"
          onClick={handleAddOption}
          className="px-4 py-2 bg-gray-300 rounded-lg"
        >
          เพิ่มตัวเลือก
        </button>
      </div>

      {/* รูปภาพ */}
      <input
        type="file"
        onChange={(e) => setImage(e.target.files[0])}
        className="w-full px-4 py-2 border rounded-lg"
      />

      {/* ปุ่มเพิ่มเมนู */}
      <div className="flex justify-end space-x-2">
        <button
          type="button"
          onClick={onClose}
          className="px-4 py-2 bg-gray-300 rounded-lg"
        >
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
  const [name, setName] = useState("");
  const [nameEn, setNameEn] = useState("");
  const [nameCh, setNameCh] = useState("");
  const [loading, setLoading] = useState(false);

  const handleAddCategory = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await axios.post("http://127.0.0.1:8080/add_category", {
        name: name.trim(),
        nameEn: nameEn.trim(),
        nameCh: nameCh.trim(),
      });

      if (response.status === 200) {
        alert(`เพิ่มหมวดหมู่สำเร็จ: ${response.data.name}`);
        setName("");
        setNameEn("");
        setNameCh("");
        onClose();
      } else {
        alert(`เกิดข้อผิดพลาด: ${response.statusText}`);
      }
    } catch (error) {
      alert(
        error.response?.data?.error || "ไม่สามารถเพิ่มหมวดหมู่ได้"
      );
      console.error("Error:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center">
      <div className="bg-white rounded-lg p-8 shadow-lg w-96">
        <h2 className="text-xl font-bold mb-4">เพิ่มหมวดหมู่</h2>
        <form onSubmit={handleAddCategory} className="space-y-4">
          <input
            type="text"
            placeholder="ชื่อหมวดหมู่ (ไทย)"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-4 py-2 border rounded-lg"
            required
          />
          <input
            type="text"
            placeholder="ชื่อหมวดหมู่ (อังกฤษ)"
            value={nameEn}
            onChange={(e) => setNameEn(e.target.value)}
            className="w-full px-4 py-2 border rounded-lg"
            required
          />
          <input
            type="text"
            placeholder="ชื่อหมวดหมู่ (จีน)"
            value={nameCh}
            onChange={(e) => setNameCh(e.target.value)}
            className="w-full px-4 py-2 border rounded-lg"
            required
          />
          <div className="flex justify-end space-x-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 bg-gray-300 rounded-lg"
            >
              ยกเลิก
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-green-500 text-white rounded-lg"
              disabled={loading}
            >
              {loading ? "กำลังเพิ่ม..." : "เพิ่มหมวดหมู่"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default MenuManagement;
