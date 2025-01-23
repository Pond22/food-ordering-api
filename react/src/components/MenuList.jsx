import React, { useState, useEffect, useRef } from "react";
import MenuItem from "./MenuItem";
import styles from "../styles/MenuList.module.css";
import { ShoppingCart } from "lucide-react";

const MenuList = () => {
  const [menuItems, setMenuItems] = useState([]); // State สำหรับเก็บข้อมูลเมนู
  const [categories, setCategories] = useState([]); // State สำหรับเก็บข้อมูลหมวดหมู่
  const [selectedCategory, setSelectedCategory] = useState(""); // State สำหรับเก็บหมวดหมู่ที่เลือก
  const [activeLink, setActiveLink] = useState(""); // State สำหรับเก็บลิงก์ที่ถูกคลิก
  const navRef = useRef(); // ใช้ ref เพื่อใช้งานกับ nav (ถ้าจำเป็น)
  const categoryRefs = useRef({}); // ใช้เก็บ refs สำหรับหมวดหมู่แต่ละอัน

  const [cart, setCart] = useState([]);
  const [isCartVisible, setIsCartVisible] = useState(false);

  // ฟังก์ชันดึงข้อมูลเมนูจาก API
  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(
          "http://localhost:8080/api/menu/ActiveMenu"
        );
        const data = await response.json();
        setMenuItems(data);

        // ดึงหมวดหมู่ที่ไม่ซ้ำกัน
        const uniqueCategories = [
          ...new Set(data.map((item) => item.Category.Name)),
        ];
        setCategories(uniqueCategories);
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };

    fetchData();
  }, []);

  // ฟังก์ชันสำหรับการคลิกหมวดหมู่จากเมนู
  const handleCategoryClick = (category) => {
    setSelectedCategory(category);
    setActiveLink(category); // ตั้งค่าหมวดหมู่ที่เลือก และเปลี่ยน active link

    // เลื่อนหน้าไปยังหมวดหมู่ที่เลือก
    categoryRefs.current[category]?.scrollIntoView({ behavior: "smooth" });
  };

  // ฟังก์ชันกรองเมนูตามหมวดหมู่ที่เลือก
  const filteredMenuItems = menuItems.filter(
    (item) => item.Category.Name === selectedCategory
  );

  const getCartItemCount = () => {
    return cart.reduce((total, item) => total + item.quantity, 0);
  };

  const handleCartToggle = () => {
    setIsCartVisible(!isCartVisible);
  };

  const handleAddToCart = (item) => {
    setCart((prevCart) => {
      const existingItem = prevCart.find((cartItem) => cartItem.id === item.id);
      if (existingItem) {
        return prevCart.map((cartItem) =>
          cartItem.id === item.id
            ? { ...cartItem, quantity: cartItem.quantity + 1 }
            : cartItem
        );
      }
      return [...prevCart, { ...item, quantity: 1 }];
    });
  };

  // เริ่มแสดงหน้าจอ
  return (
    <div>
      <header className={styles.header}>
        <div className="flex justify-between ">
          <div className="flex justify-center items-center ">
            <h1>โต๊ะที่{}</h1>
          </div>
          <div className={styles.cartIcon}>
            <button className={styles.cartButton} onClick={handleCartToggle}>
              {/* 🛒  */}
              <ShoppingCart />
              {getCartItemCount() > 0 && (
                <span className={styles.cartBadge}>{getCartItemCount()}</span>
              )}
            </button>
            {isCartVisible && (
              <div className={styles.cartDetails}>
                <ul>
                  {cart.map((item) => (
                    <li key={item.id}>
                      {item.name} x {item.quantity} -{" "}
                      {item.price * item.quantity} ฿
                    </li>
                  ))}
                </ul>
                <button onClick={() => alert("Checkout")}>Checkout</button>
              </div>
            )}
          </div>
        </div>

        {/* แสดงหมวดหมู่ในรูปแบบของลิงก์ใน <nav> */}
        <nav className={styles.nav}>
          {categories.map((category, index) => (
            <a
              key={index}
              href={`#${category}`} // ใช้ชื่อหมวดหมู่เป็น href
              className={activeLink === category ? styles.active : ""}
              onClick={() => handleCategoryClick(category)} // เรียกใช้ฟังก์ชันเมื่อคลิก
            >
              {category} {/* แสดงชื่อหมวดหมู่ */}
            </a>
          ))}
        </nav>
      </header>

      {/* แสดงเมนูโปรโมชัน */}
      {/* <div id="" className={styles.promotionSection}>
        <h2>Promotion</h2>
        
        <div className={styles.promotionBanner}>
          <button className={styles.arrow} onClick={() => handleArrowClick('left')}>&lt;</button>
          {promotionItems.map((item, index) => (
            <div
              key={item.id}
              className={
                index === currentIndex
                  ? `${styles.promotionItem} ${styles.active}` 
                  : styles.promotionItem
              }
            >
              <img
                src={item.image}
                alt={item.name}
                className={styles.promotionImage}
              />
              <h4>{item.name}</h4>
              <p>{item.description}</p>
              <span>{item.price} ฿</span>
            </div>
          ))}
          <button className={styles.arrow} onClick={() => handleArrowClick('right')}>&gt;</button>
        </div>
        <div className={styles.navigator}>
          {promotionItems.map((_, index) => (
            <span
              key={index}
              className={ 
                index === currentIndex 
                  ? `${styles.navigatorDot} ${styles.active}` 
                  : styles.navigatorDot
              }
              onClick={() => handleNavigatorClick(index)}
            ></span>
          ))}
        </div>
      </div> */}

      {/* แสดงเมนูในรูปแบบ flatlist โดยแยกตามหมวดหมู่ */}
      <div className={styles.menuList}>
        {categories.map((category, index) => {
          // กรองเมนูตามหมวดหมู่
          const filteredMenuItems = menuItems.filter(
            (item) => item.Category.Name === category
          );

          return (
            <div
              key={index}
              ref={(el) => (categoryRefs.current[category] = el)} // ตั้ง ref ให้กับแต่ละหมวดหมู่
              id={category}
              className={styles.categorySection}
            >
              <h2>{category}</h2> {/* แสดงชื่อหมวดหมู่ */}
              <div className={styles.menuGrid}>
                {filteredMenuItems.map((item) => (
                  <MenuItem
                    key={item.ID}
                    item={item}
                    addToCart={() => handleAddToCart(item)}
                  />
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default MenuList;
