import React, { useState } from "react";
import axios from "axios";

const AddCategory = () => {
  const [categoryName, setCategoryName] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  const categoryAPI = "http://127.0.0.1:8080/add_category";

  const handleCategorySubmit = async (e) => {
    e.preventDefault();

    // Reset messages
    setSuccessMessage("");
    setErrorMessage("");

    try {
      const response = await axios.post(
        categoryAPI,
        { name: categoryName },
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );

      if (response.status === 200) {
        setSuccessMessage(`หมวดหมู่ "${response.data.name}" ถูกเพิ่มเรียบร้อยแล้ว!`);
        setCategoryName(""); // รีเซ็ตฟิลด์
      }
    } catch (error) {
      if (error.response) {
        // ข้อผิดพลาดจากเซิร์ฟเวอร์
        if (error.response.status === 400) {
          setErrorMessage("ข้อมูลหมวดหมู่ไม่ถูกต้อง กรุณาลองใหม่อีกครั้ง");
        } else if (error.response.status === 500) {
          setErrorMessage("เกิดข้อผิดพลาดจากเซิร์ฟเวอร์ กรุณาลองใหม่ภายหลัง");
        } else {
          setErrorMessage("เกิดข้อผิดพลาดที่ไม่คาดคิด");
        }
      } else {
        setErrorMessage("ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้");
      }
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100 px-4">
      <div className="w-full max-w-lg bg-white rounded-lg shadow-lg p-8">
        <h1 className="text-2xl font-bold text-gray-800 mb-6 text-center">
          เพิ่มหมวดหมู่ใหม่
        </h1>
        <form onSubmit={handleCategorySubmit} className="space-y-4">
          {/* Input for Category Name */}
          <div>
            <label htmlFor="categoryName" className="block text-gray-700 font-medium mb-2">
              ชื่อหมวดหมู่
            </label>
            <input
              id="categoryName"
              type="text"
              value={categoryName}
              onChange={(e) => setCategoryName(e.target.value)}
              className="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="ใส่ชื่อหมวดหมู่"
              required
            />
          </div>

          {/* Submit Button */}
          <div>
            <button
              type="submit"
              className="w-full bg-blue-500 text-white font-medium py-2 px-4 rounded-lg hover:bg-blue-600 transition duration-200"
            >
              เพิ่มหมวดหมู่
            </button>
          </div>
        </form>

        {/* Success Message */}
        {successMessage && (
          <div className="mt-4 text-green-600 font-medium text-center">{successMessage}</div>
        )}

        {/* Error Message */}
        {errorMessage && (
          <div className="mt-4 text-red-600 font-medium text-center">{errorMessage}</div>
        )}
      </div>
    </div>
  );
};

export default AddCategory;
