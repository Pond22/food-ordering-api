import React, { useState, useEffect, useRef } from "react";
import { Carousel, Card } from "flowbite-react";
import MenuItem from "./MenuItem";

export default function MenuList() {
  const [categories, setCategories] = useState([]);
  const [menus, setMenus] = useState([]);
  const [activeLink, setActiveLink] = useState(0);
  const scrollRef = useRef(null);

  const handleCategoryClick = (category) => {
    setActiveLink(category);
  };

  useEffect(() => {
    const fetchCategoryData = async () => {
      try {
        const response = await fetch("http://localhost:8080/api/categories");
        const data = await response.json();
        setCategories(data);
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };

    const fetchMenuData = async () => {
      try {
        const response = await fetch(
          "http://localhost:8080/api/menu/ActiveMenu"
        );
        const data = await response.json();
        setMenus(data);
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };

    fetchCategoryData();
    fetchMenuData();
  }, []);

  const handleDragScroll = () => {
    const container = scrollRef.current;
    let isDragging = false;
    let startX = 0;
    let scrollLeft = 0;

    const onMouseDown = (e) => {
      isDragging = true;
      startX = e.pageX - container.offsetLeft;
      scrollLeft = container.scrollLeft;
    };

    const onMouseMove = (e) => {
      if (!isDragging) return;
      e.preventDefault();
      const x = e.pageX - container.offsetLeft;
      const walk = (x - startX) * 1.5; // ความเร็วการเลื่อน
      container.scrollLeft = scrollLeft - walk;
    };

    const onMouseUp = () => {
      isDragging = false;
    };

    container.addEventListener("mousedown", onMouseDown);
    container.addEventListener("mousemove", onMouseMove);
    container.addEventListener("mouseup", onMouseUp);
    container.addEventListener("mouseleave", onMouseUp);

    return () => {
      container.removeEventListener("mousedown", onMouseDown);
      container.removeEventListener("mousemove", onMouseMove);
      container.removeEventListener("mouseup", onMouseUp);
      container.removeEventListener("mouseleave", onMouseUp);
    };
  };

  useEffect(() => {
    handleDragScroll();
  }, []);

  return (
    <div className="mt-[4.5rem] text-black px-4 py-2 max-w-screen-xl mx-auto">
      {/* CATEGORY */}
      <div
        className="mt-2 flex justify-start items-center gap-2 overflow-hidden cursor-grab"
        ref={scrollRef}
        style={{ whiteSpace: "nowrap" }}
      >
        {/* ปุ่ม "ทั้งหมด" */}
        <button
          className={
            activeLink === 0
              ? "bg-blue-500 hover:bg-blue-600 transition-colors px-4 py-1.5 rounded-md text-white"
              : "bg-gray-300 hover:bg-gray-400 transition-colors px-4 py-1.5 rounded-md text-black"
          }
          onClick={() => handleCategoryClick(0)}
        >
          ทั้งหมด
        </button>

        {/* สร้างปุ่มจาก categories */}
        {categories.map((item) => (
          <button
            key={item.ID}
            className={
              activeLink === item.ID
                ? "bg-blue-500 hover:bg-blue-600 transition-colors px-4 py-1.5 rounded-md text-white"
                : "bg-gray-300 hover:bg-gray-400 transition-colors px-4 py-1.5 rounded-md text-black"
            }
            onClick={() => handleCategoryClick(item.ID)}
          >
            {item.Name}
          </button>
        ))}
      </div>

      {/* PROMOTION */}
      <div className="my-4 h-48 sm:h-64 xl:h-80 2xl:h-96">
        <Carousel leftControl=" " rightControl=" ">
          <img
            src="https://flowbite.com/docs/images/carousel/carousel-1.svg"
            alt="..."
          />
          <img
            src="https://flowbite.com/docs/images/carousel/carousel-2.svg"
            alt="..."
          />
        </Carousel>
      </div>

      {/* MENU */}
      <div className="my-4">
        {categories.map((category) => {
          // กรองเมนูตามหมวดหมู่ที่เลือก
          const filteredMenus =
            activeLink === 0 // เลือก "ทั้งหมด"
              ? menus // แสดงทุกเมนู
              : menus.filter((menu) => menu.Category.ID === activeLink); // กรองตามหมวดหมู่

          // กรองเมนูในหมวดหมู่นี้
          const categoryMenus = filteredMenus.filter(
            (menu) => menu.Category.ID === category.ID
          );

          // หากไม่มีเมนูในหมวดหมู่นี้ ไม่ต้องแสดง
          if (categoryMenus.length === 0) return null;

          return (
            <div
              key={category.ID}
              id={`category-${category.ID}`}
              className="mb-6"
            >
              {/* ชื่อหมวดหมู่ */}
              <h2 className="text-xl font-semibold mb-2">{category.Name}</h2>

              {/* แสดงเมนูในหมวดหมู่นี้ */}
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {categoryMenus.map((menu) => (
                  <MenuItem
                    key={menu.ID}
                    item={menu}
                    // addToCart={() => handleAddToCart(menu)}
                  />
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
