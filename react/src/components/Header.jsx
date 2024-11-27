import React from 'react';
import styles from '../styles/Header.module.css';

const Header = () => {
  return (
    <header className={styles.header}>
      <div className={styles.logo}>🍲 ชาบูเชียงใหม่</div>
      <nav className={styles.nav}>
      <a href="#menu">เมนูแนะนำ</a>
        <a href="#contact">เมนูใหม่</a>
        <a href="#about">ของหวาน</a>
      </nav>
      
    </header>
  );
};

export default Header;
