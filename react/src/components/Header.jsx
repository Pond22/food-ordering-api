import React from 'react';
import styles from '../styles/Header.module.css';

const Header = () => {

  
  return (
    
    <header className={styles.header}>
      {/* <img src="/image/images (1).jpg" alt="Logo" className='w-36 h-36'/> */}
      
      {/* <div className="flex items-center my-5 text-2xl font-bold"><img src="https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQ0rYWmcUDrqEqeLt0_qIwNjrBEYsLP8Fwunw&s" 
      alt="External Image" className='w-12 h-12 mr-2 rounded-full'/> Grand Kaze Yakiniku Changemai</div> */}
      <div className="flex items-center my-5 text-2xl font-bold"> โต๊ะที่{}</div>
      <nav className={styles.nav}>
      <a href="#menu">เมนูแนะนำ</a>
        <a href="#contact">เมนูใหม่</a>
        <a href="#about">ของหวาน</a>
        <a href="#about">ของหวาน</a>
        <a href="#about">ของหวาน</a>
        <a href="#about">ของหวาน</a>
      </nav>
      
    </header>
  );
};

export default Header;
