import React, { useEffect } from "react";
import MenuBar from "../components/MenuBar";
import MenuList from "../components/MenuList2";
import useCartStore from "../hooks/cart-store";

export default function Menu() {
  const { checkCartExpiry, setTableData } = useCartStore();

  const queryParams = new URLSearchParams(window.location.search);
  const tableId = queryParams.get("table");
  const uuid = queryParams.get("uuid");

  useEffect(() => {
    checkCartExpiry();
  }, []);

  useEffect(() => {
    if (tableId && uuid) {
      setTableData(tableId, uuid);
    }
  }, [tableId, uuid, setTableData]);

  return (
    <div className="bg-gray-100 min-h-screen">
      <MenuBar />
      <MenuList />
    </div>
  );
}
