import React from "react";
import MenuBar from "../components/MenuBar";
import MenuList from "../components/MenuList2";

export default function Menu() {
  return (
    <div className="bg-gray-100 min-h-screen">
      <MenuBar />
      <MenuList />
    </div>
  );
}
