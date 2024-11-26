import React from "react";
import styles from "../styles/MenuItem.module.css";

const MenuItem = ({ item, onUpdateQuantity }) => {
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
    <div className={styles.menuItem}>
      <img src={item.image} alt={item.name} className={styles.image} />
      <div className={styles.details}>
        <h3>{item.name}</h3>
        <p>ราคา: ฿{item.price}</p>
        <div className={styles.quantityControl}>
          <button onClick={handleDecrease} disabled={item.quantity === 0}>
            -
          </button>
          <span>{item.quantity}</span>
          <button onClick={handleIncrease} disabled={item.quantity === 50}>
            +
          </button>
        </div>
      </div>
    </div>
  );
};

export default MenuItem;
