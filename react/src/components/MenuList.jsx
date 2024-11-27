import React, { useState } from 'react';
import MenuItem from './MenuItem';
import styles from '../styles/MenuList.module.css';
import axios from 'axios';
const baseURL = "http://127.0.0.1:8080/getCategory"
const MenuList = () => {
  const [menuItems, setMenuItems] = useState([
    { id: 1, name: 'เนื้อหมู', price: 150, image: 'https://www.mcfs.co.th/wp-content/uploads/2022/08/XL_01094.jpg', quantity: 0 },
    { id: 2, name: 'เนื้อวัว', price: 200, image: 'https://thagoonmanee.com/wp-content/uploads/2019/04/%E0%B8%8A%E0%B8%B2%E0%B8%9A%E0%B8%B9.png', quantity: 0 },
    { id: 3, name: 'ไก่', price: 120, image: 'https://s359.kapook.com/pagebuilder/a265eefe-c8d8-4559-9bdd-325004243176.jpg', quantity: 0 },
    { id: 4, name: 'เนื้อii', price: 150, image: 'https://www.mcfs.co.th/wp-content/uploads/2022/08/XL_01094.jpg', quantity: 0 },
    { id: 5, name: 'เนื้kk', price: 150, image: 'https://www.mcfs.co.th/wp-content/uploads/2022/08/XL_01094.jpg', quantity: 0 },
  ]);

  const handleUpdateQuantity = (id, quantity) => {
    setMenuItems(menuItems.map(item =>
      item.id === id ? { ...item, quantity } : item
    ));
  };

  return (
    <div id="menu" className={styles.menuList}>
      <h2>เมนูอาหาร</h2>
      <div className={styles.menuGrid}>
        {menuItems.map(item => (
          <MenuItem
            key={item.id}
            item={item}
            onUpdateQuantity={handleUpdateQuantity}
          />
        ))}
      </div>
    </div>
  );
};

export default MenuList;
