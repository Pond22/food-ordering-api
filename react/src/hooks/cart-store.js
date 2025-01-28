import { create } from "zustand";
import { persist } from "zustand/middleware";

const useCartStore = create(
  persist(
    (set) => ({
      cart: [], // ตะกร้าสินค้า

      // เพิ่มสินค้าในตะกร้า
      addToCart: (item, quantity, note, selectedOptions) =>
        set((state) => {
          const existingItem = state.cart.find((i) => i.ID === item.ID);
          if (existingItem) {
            return {
              cart: state.cart.map((i) =>
                i.ID === item.ID
                  ? {
                      ...i,
                      quantity: i.quantity + quantity,
                      note,
                      selectedOptions,
                    }
                  : i
              ),
            };
          } else {
            return {
              cart: [
                ...state.cart,
                {
                  ...item,
                  quantity,
                  note,
                  selectedOptions,
                },
              ],
            };
          }
        }),

      // เพิ่มจำนวนสินค้าในตะกร้า
      increaseQuantity: (itemId) =>
        set((state) => ({
          cart: state.cart.map((i) =>
            i.ID === itemId ? { ...i, quantity: i.quantity + 1 } : i
          ),
        })),

      // ลดจำนวนสินค้าในตะกร้า
      decreaseQuantity: (itemId) =>
        set((state) => ({
          cart: state.cart
            .map((i) =>
              i.ID === itemId
                ? { ...i, quantity: i.quantity - 1 }
                : i
            )
            .filter((i) => i.quantity > 0),
        })),

      // ลบสินค้าออกจากตะกร้า
      removeFromCart: (itemId) =>
        set((state) => ({
          cart: state.cart.filter((i) => i.ID !== itemId),
        })),

      // ล้างตะกร้าทั้งหมด
      clearCart: () => set({ cart: [] }),
    }),
    {
      name: "cart-storage", // ชื่อ key ที่เก็บใน localStorage
      getStorage: () => localStorage, // ระบุว่าใช้ localStorage
    }
  )
);

export default useCartStore;
