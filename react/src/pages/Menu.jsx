import React from "react";
import MenuBar from "../components/MenuBar";
import MenuList from "../components/MenuList2";
import { useSearchParams } from 'react-router-dom';

export default function Menu() {
  const [searchParams] = useSearchParams();
  
  // ดึงค่า table และ uuid จาก URL
  const table = searchParams.get('table');
  const uuid = searchParams.get('uuid');
  console.log(table," + ", uuid);

  return (
    <div className="bg-gray-100 min-h-screen">
      <MenuBar />
      <MenuList />
    </div>
  );
}
