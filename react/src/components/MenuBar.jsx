import React, { useState, useEffect } from "react";
import { Modal } from "flowbite-react";
import {
  Banknote,
  Bell,
  ShoppingCart,
  SquareMenu,
  User,
  PlusIcon,
  Minus,
} from "lucide-react";
import useCartStore from "../hooks/cart-store";

export default function MenuBar() {
  const [openCallModal, setOpenCallModal] = useState(false);
  const [openCartModal, setOpenCartModal] = useState(false);
  const { cart, increaseQuantity, decreaseQuantity } = useCartStore();
  console.log(cart);
  const calculateTotal = () => {
    return cart.reduce((total, item) => total + item.Price * item.quantity, 0);
  };

  useEffect(() => {
    calculateTotal();
  }, [cart]);

  return (
    <nav className="bg-[#1C2B41] shadow-sm fixed top-0 w-full z-50">
      <div className="max-w-screen-xl flex flex-wrap items-center justify-between mx-auto px-4 py-2">
        <div className="flex flex-col gap-1 justify-center items-start text-white">
          <span className="text-xl font-semibold whitespace-nowrap">
            EasyOrder
          </span>
          <span className="text-base font-semibold whitespace-nowrap">
            โต๊ะที่...
          </span>
        </div>
        <div className="block w-auto">
          <ul className="font-medium flex gap-4">
            <li>
              <button
                type="button"
                onClick={() => setOpenCallModal(true)}
                className="block py-2 text-white"
              >
                <Bell />
              </button>
              <Modal
                show={openCallModal}
                size="md"
                onClose={() => setOpenCallModal(false)}
                popup
                className=""
              >
                <Modal.Header className="bg-gray-200" />
                <Modal.Body className="bg-gray-200 rounded-md">
                  <div className="flex justify-center items-center gap-4">
                    <button
                      type="button"
                      className="p-4 bg-white gap-1 hover:bg-blue-500 transition-colors shadow-sm rounded-md text-black hover:text-white flex flex-col items-center justify-center"
                    >
                      <User className="size-7" />
                      <span>Call Staff</span>
                    </button>
                    <button
                      type="button"
                      className="p-4 bg-white gap-1 hover:bg-blue-500 transition-colors shadow-sm rounded-md text-black hover:text-white flex flex-col items-center justify-center"
                    >
                      <Banknote className="size-7" />
                      <span>Payment</span>
                    </button>
                  </div>
                </Modal.Body>
              </Modal>
            </li>
            <li>
              <button type="button" className="block py-2 text-white">
                <SquareMenu />
              </button>
            </li>
            <li>
              <button
                type="button"
                onClick={() => setOpenCartModal((prev) => !prev)}
                className="relative my-2 inline-flex items-center text-sm font-medium text-center text-white"
              >
                <ShoppingCart />
                <span className="sr-only">Cart</span>
                <div className="absolute inline-flex items-center justify-center w-5 h-5 text-xs font-bold text-white bg-red-500 rounded-full -top-2 -end-2">
                  {cart.length}
                </div>
              </button>
              {openCartModal && (
                <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50 p-4">
                  <div className="bg-white px-2 pt-2 pb-4 border border-gray/50 rounded-xl w-full max-w-xl min-h-full max-h-screen flex flex-col">
                    {/* Header */}
                    <div className="flex top-0 justify-between items-center">
                      <h4 className="text-xl font-semibold text-black">
                        รายการอาหารที่สั่ง
                      </h4>
                      <button
                        onClick={() => setOpenCartModal((prev) => !prev)}
                        className="text-white/70 hover:text-white z-10 bg-black/30 rounded-full p-2"
                      >
                        <svg
                          className="w-5 h-5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth="2"
                            d="M6 18L18 6M6 6l12 12"
                          />
                        </svg>
                      </button>
                    </div>
                    <hr className="mt-4" />

                    {/* Scrollable Content */}
                    <div className="flex-1 overflow-y-auto py-2">
                      {cart.map((item, i) => (
                        <div
                          key={item.ID}
                          className="flex justify-between items-center gap-2 p-2 border-b last:border-none"
                        >
                          <div className="flex flex-col gap-1">
                            <h6 className="text-lg font-medium">
                              {item.Name}
                            </h6>
                            <span className="text-sm text-gray-600">
                              หมายเหตุ : {item.note || "-"}
                            </span>
                          </div>
                          <div className="flex flex-col items-center gap-1">
                            <span className="text-base font-medium">
                              {item.Price * item.quantity}&nbsp;บาท
                            </span>
                            <div className="flex justify-between items-center gap-3">
                              <button
                                className="bg-gray-200 hover:bg-gray-300 disabled:opacity-70 disabled:hover:bg-gray-200 p-2 rounded-md text-black"
                                onClick={() => decreaseQuantity(item.ID)}
                              >
                                <Minus className="size-2" />
                              </button>
                              <span className="text-base font-semibold">
                                {item.quantity}
                              </span>
                              <button
                                className="bg-gray-200 hover:bg-gray-300 p-2 rounded-md text-black"
                                onClick={() => increaseQuantity(item.ID)}
                              >
                                <PlusIcon className="size-2" />
                              </button>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>

                    {/* Footer */}
                    <div className="flex flex-col gap-4 px-2 pt-4 border-t border-gray-200">
                      <div className="flex justify-between items-center">
                        <span className="text-xl font-bold">รวมทั้งหมด</span>
                        <span className="text-xl font-bold">
                          {calculateTotal()}&nbsp;บาท
                        </span>
                      </div>
                      <button className="px-4 py-3 w-full rounded-md text-white font-semibold bg-[#4bcc37] hover:bg-[#4aac3b]">
                        ยืนยันการสั่งอาหาร
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </li>
          </ul>
        </div>
      </div>
    </nav>
  );
}
