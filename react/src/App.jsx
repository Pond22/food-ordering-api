import React from "react";
import { BrowserRouter as Router, Routes, Route, Link } from "react-router-dom";
import Header from "./components/Header";
import AddCategory from "./components/AddCategory";
import MenuList from "./components/MenuList";
import TableManager from "./components/TableManager";
import styles from "./styles/App.module.css";
import AddMenu from "./components/AddMenu";

// import styles from "./index.css";

const App = () => {
  return (
    <Router>
      <div className={styles.app}>
        {/* <Header /> */}
        <nav className="pl-2 ">
          <Link to="/menu" className="px-2">เมนูอาหาร</Link>
          <Link to="/tables" className="px-2">จัดการโต๊ะ</Link>
          <Link to="/addMenu" className="px-2">เพิ่มเมนู</Link>
          <Link to="/addCategory" className="px-2">เพิ่มหมวดหมู่</Link>
          
        </nav>
        <Routes>
          <Route path="/menu" element={<MenuList />} />
          <Route path="/tables" element={<TableManager />} />
          <Route path="/addMenu"element={<AddMenu/>}/>
          <Route path="/addCategory"element={<AddCategory/>}/>
          
        </Routes>
      </div>
    </Router>
  );
};

export default App;
