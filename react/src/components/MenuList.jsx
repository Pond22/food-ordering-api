import React, { useState, useEffect, useRef } from 'react';
import MenuItem from './MenuItem';
import styles from '../styles/MenuList.module.css';

const MenuList = () => {
  
  const [menuItems, setMenuItems] = useState([
    { id: 1, type: 'Promotion', name: '特選リブロース | Premium Ribeye', price: 150, image: 'https://www.mcfs.co.th/wp-content/uploads/2022/08/XL_01094.jpg', description: '特選リブロース | Premium Ribeye', quantity: 0 },
    { id: 2, type: 'NewMenu', name: 'เนื้อวัว', price: 200, image: 'https://thagoonmanee.com/wp-content/uploads/2019/04/%E0%B8%8A%E0%B8%B2%E0%B8%9A%E0%B8%B9.png', description: '特選リブロース | Premium Ribeye', quantity: 0 },
    { id: 3, type: 'Aracart', name: 'ไก่', price: 120, image: 'https://s359.kapook.com/pagebuilder/a265eefe-c8d8-4559-9bdd-325004243176.jpg', description: '特選リブロース | Premium Ribeye', quantity: 0 },
    { id: 4, type: 'Desert', name: 'เนื้อii', price: 150, image: 'https://www.mcfs.co.th/wp-content/uploads/2022/08/XL_01094.jpg', description: '特選リブロース | Premium Ribeye', quantity: 0 },
    { id: 5, type: 'NewMenu', name: 'เนื้kk', price: 150, image: 'https://www.mcfs.co.th/wp-content/uploads/2022/08/XL_01094.jpg', description: '特選リブロース | Premium Ribeye', quantity: 0 },
    { id: 6, type: 'Promotion', name: 'เนื้อวัว', price: 200, image: 'https://thagoonmanee.com/wp-content/uploads/2019/04/%E0%B8%8A%E0%B8%B2%E0%B8%9A%E0%B8%B9.png', description: '特選リブロース | Premium Ribeye', quantity: 0 },
    { id: 7, type: 'NewMenu', name: 'เนื้อวัว', price: 200, image: 'https://thagoonmanee.com/wp-content/uploads/2019/04/%E0%B8%8A%E0%B8%B2%E0%B8%9A%E0%B8%B9.png', description: '特選リブロース | Premium Ribeye', quantity: 0 },
    { id: 8, type: 'Aracart', name: 'เนื้อวัว', price: 200, image: 'https://thagoonmanee.com/wp-content/uploads/2019/04/%E0%B8%8A%E0%B8%B2%E0%B8%9A%E0%B8%B9.png', description: '特選リブロース | Premium Ribeye', quantity: 0 },

    { id: 9, type: 'Aracart', name: 'เนื้อวัว', price: 200, image: 'https://thagoonmanee.com/wp-content/uploads/2019/04/%E0%B8%8A%E0%B8%B2%E0%B8%9A%E0%B8%B9.png', description: '特選リブロース | Premium Ribeye', quantity: 0 },
  ]);

  
  const [activeLink, setActiveLink] = useState('Promotion'); // เริ่มต้นที่ "Promotion"
  const navRef = useRef(null);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isManual, setIsManual] = useState(false);
  const intervalRef = useRef(null);
  const promotionItems = menuItems.filter(item => item.type === 'Promotion');

  const filteredItems = menuItems.filter(item => item.type === activeLink); // กรองเมนูตามประเภทที่เลือก
  // ตระกร้า
  const [cart, setCart] = useState([]);
  const [isCartVisible, setIsCartVisible] = useState(false);

  const addToCart = (item) => {
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

  const getCartItemCount = () => {
    return cart.reduce((total, item) => total + item.quantity, 0);
  };

  const handleCartToggle = () => {
    setIsCartVisible(!isCartVisible);
  };



  const startAutoScroll = () => {
    intervalRef.current = setInterval(() => {
      setCurrentIndex((prevIndex) => (prevIndex + 1) % promotionItems.length);
    }, 3000);
  };

  const stopAutoScroll = () => {
    clearInterval(intervalRef.current);
  };

  useEffect(() => {
    if (!isManual) {
      startAutoScroll();
    }
    return () => stopAutoScroll();
  }, [isManual, promotionItems.length]);

  const handleNavigatorClick = (index) => {
    stopAutoScroll();
    setCurrentIndex(index);
    setIsManual(true);
    setTimeout(() => setIsManual(false), 3000); // กลับไปเลื่อนอัตโนมัติหลัง 10 วินาที
  };

  // ลูกศรใน banner
  const handleArrowClick = (direction) => {
    stopAutoScroll();
    setCurrentIndex((prevIndex) => {
      if (direction === 'left') {
        return (prevIndex - 1 + promotionItems.length) % promotionItems.length;
      } else {
        return (prevIndex + 1) % promotionItems.length;
      }
    });
    setIsManual(true);
    setTimeout(() => setIsManual(false), 10000);
  };

  useEffect(() => {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const id = entry.target.id;
          setActiveLink(id);  // เมื่อ div เข้าในมุมมอง จะทำให้ลิงก์นั้นเปลี่ยนสี
        }
      });
    }, { threshold: 0.5 });

    document.querySelectorAll('[id]').forEach((el) => observer.observe(el));

    return () => observer.disconnect();
  }, []);

  const handleAddToCart = (item) => {
    setCart((prevCart) => {
      const existingItem = prevCart.find(cartItem => cartItem.id === item.id);
      if (existingItem) {
        return prevCart.map(cartItem =>
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
        <div className='flex justify-between'>
          <div className='flex justify-center items-center'>
        <h1>โต๊ะที่{}</h1></div>
        <div className={styles.cartIcon}>
          <button className={styles.cartButton} onClick={handleCartToggle}>
          🛒 {getCartItemCount() > 0 && (
              <span className={styles.cartBadge}>{getCartItemCount()}</span>
            )}
          </button>
          {isCartVisible && (
            <div className={styles.cartDetails}>
              <ul>
                {cart.map((item) => (
                  <li key={item.id}>
                    {item.name} x {item.quantity} - {item.price * item.quantity} ฿
                  </li>
                ))}
              </ul>
              <button onClick={() => alert('Checkout')}>Checkout</button>
            </div>
          )}
        </div>
        </div>

        <nav className={styles.nav} ref={navRef}>
          {/* ลิงก์เมนูที่กรองจากประเภทของเมนู */}
          <a
            href="#Promotion"
            className={activeLink === 'Promotion' ? styles.active : ''}
            onClick={() => setActiveLink('Promotion')}
          >
            Promotion
          </a>
          <a
            href="#NewMenu"
            className={activeLink === 'NewMenu' ? styles.active : ''}
            onClick={() => setActiveLink('NewMenu')}
          >
            เมนูใหม่
          </a>
          <a
            href="#Aracart"
            className={activeLink === 'Aracart' ? styles.active : ''}
            onClick={() => setActiveLink('Aracart')}
          >
            อาหารจานเดียว
          </a>
          <a
            href="#Desert"
            className={activeLink === 'Desert' ? styles.active : ''}
            onClick={() => setActiveLink('Desert')}
          >
            ของหวาน
          </a>
        </nav>
      </header>

      {/* แสดงเมนูโปรโมชัน */}
      <div id="Promotion" className={styles.promotionSection}>
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
      </div>

{/* แสดงเมนูตามประเภทที่เลือก */}
      <div id="NewMenu" className={styles.menuList}>
        <h2>เมนูใหม่</h2>
        <div className={styles.menuGrid}>
          {menuItems.filter(item => item.type === 'NewMenu').map(item => (
            <MenuItem
              key={item.id}
              item={item}
              onUpdateQuantity={() => {}}
              // เพิ่มลงตระกร้า
              addToCart={addToCart} 
            />
          ))}
        </div>
      </div>

      <div id="Aracart" className={styles.menuList}>
        <h2>อาหารจานเดียว</h2>
        <div className={styles.menuGrid}>
          {menuItems.filter(item => item.type === 'Aracart').map(item => (
            <MenuItem
              key={item.id}
              item={item}
              onUpdateQuantity={() => {}}
              // เพิ่มลงตระกร้า
              addToCart={addToCart} 
            />
          ))}
        </div>
      </div>

      <div id="Desert" className={styles.menuList}>
        <h2>ของหวาน</h2>
        <div className={styles.menuGrid}>
          {menuItems.filter(item => item.type === 'Desert').map(item => (
            <MenuItem
              key={item.id}
              item={item}
              onUpdateQuantity={() => {}}
              // เพิ่มลงตระกร้า
              addToCart={addToCart} 
            />
          ))}
        </div>
      </div>
    </div>
  );
};

export default MenuList;
