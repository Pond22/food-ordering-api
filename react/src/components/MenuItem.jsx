import React, { useState } from "react";
import styles from "../styles/MenuItem.module.css";



const MenuItem = ({ item, addToCart }) => {
  const [isPopupOpen, setIsPopupOpen] = useState(false);
  const [note, setNote] = useState(""); // สำหรับเก็บข้อความถึงครัว

  const togglePopup = () => {
    setIsPopupOpen(!isPopupOpen);
  };

  const handleAddToCart = () => {
    addToCart(item, note); // ส่ง item และ note ไปที่ addToCart
    setNote(""); // เคลียร์ข้อความหลังจากเพิ่มสินค้า
    togglePopup(); // ปิด popup หลังจากเพิ่มสินค้า
  };



  const handleIncrease = () => {
    if (item.quantity < 50) {
      onUpdateQuantity(item.id, item.quantity + 1);
    }
  };

  const handleDecrease = () => {
    if (item.quantity > 0) {
      onUpdateQuantity(item.id, item.quantity - 1);
    }
  };


  return (
    <div>
      
    <div className={styles.menuItem} >
      <img
        src={item.image}
        alt={item.name}
        className={styles.image}
        onClick={togglePopup} // เมื่อคลิกที่รูปหรือชื่อจะเปิด/ปิด popup
      />
      <div className={styles.details}>
        <h3 onClick={togglePopup}>{item.name}</h3> {/* เมื่อคลิกที่ชื่อเมนูจะเปิด/ปิด popup */}
        <h4>{item.description}</h4>
        <div className={styles.quantityControl}>
          <p>{item.price}฿</p>
          {/* // เมื่อคลิกที่รูปจะเปิด/ปิด popup */}
          <button onClick={togglePopup} >
            +
          </button>
        </div>
      </div>

      {/* Popup: แสดงเมื่อ isPopupOpen เป็น true */}
      {isPopupOpen && (
        <div className={styles.popup}>
          <div className={styles.popupContent}>
            <button onClick={togglePopup} className={styles.closeButton}>X</button>
            <img src={item.image} alt={item.name} className={styles.popupImage} />


            <div className="h-3/6 mx-1 border border-slate-200 rounded-b-xl bg-white text-black">
              <h3>{item.name}</h3>

              <p>{item.description}</p>
              <p>{item.price}฿</p>
            </div>

            {/* ช่องกรอกข้อความถึงครัว */}
            <div className="mt-1 mx-1 w-auto pb-2 border border-slate-200 bg-white justify-center items-center text-white rounded-xl">
              <label for="message" class="block my-3 ml-2 text-sm text-left font-medium  ">ข้อความถึงครัว</label>
              <div className="mx-2  flex justify-center">
                <textarea id="message" rows="4" class="block px-2 w-11/12 text-sm text-gray-900 bg-gray-50 rounded-lg border
               border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600
                dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"placeholder="หมายเหตุ..."
                value={note}
                  onChange={(e) => setNote(e.target.value)} // เก็บข้อความที่พิมพ์
                  >

                </textarea>
              </div>
              
              {/* ปุ่ม Add to Cart */}
              <button className={styles.btnaddCart}  onClick={() => addToCart(item)}>
                <p className="text-white ">Add To Cart</p>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
    </div>
  );
};

export default MenuItem;
