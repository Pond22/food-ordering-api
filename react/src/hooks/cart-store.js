import { create } from 'zustand'
import { persist } from 'zustand/middleware'

const useCartStore = create(
  persist(
    (set) => ({
      cart: [], // ตะกร้าสินค้า

      // คำนวณราคาสินค้ารวมตัวเลือก
      calculateItemPrice: (item, selectedOptions) => {
        let totalPrice = item.Price || 0

        // คำนวณราคาของตัวเลือกที่เลือก
        if (selectedOptions) {
          Object.keys(selectedOptions).forEach((groupName) => {
            const option = selectedOptions[groupName]
            totalPrice += option.price || 0 // เพิ่มราคาของตัวเลือก
          })
        }

        return totalPrice
      },

      // เพิ่มสินค้าในตะกร้า
      addToCart: (item, quantity, note, selectedOptions) =>
        set((state) => {
          const itemPrice = state.calculateItemPrice(item, selectedOptions) // คำนวณราคาใหม่ของสินค้า

          // ตรวจสอบว่ามีสินค้าในตะกร้าแล้วหรือไม่
          const existingItem = state.cart.find((i) => i.ID === item.ID)

          if (existingItem) {
            // ถ้ามีสินค้ารายการเดิมอยู่แล้ว
            return {
              cart: state.cart.map((i) =>
                i.ID === item.ID
                  ? {
                      ...i,
                      quantity: i.quantity + quantity,
                      note,
                      selectedOptions,
                      Price: itemPrice, // อัปเดตราคาสินค้า
                    }
                  : i
              ),
            }
          } else {
            // ถ้าไม่มีสินค้าในตะกร้า
            return {
              cart: [
                ...state.cart,
                {
                  ...item,
                  quantity,
                  note,
                  selectedOptions,
                  Price: itemPrice, // อัปเดตราคาสินค้า
                },
              ],
            }
          }
        }),

      // เพิ่มจำนวนสินค้าในตะกร้า
      increaseQuantity: (itemId) =>
        set((state) => ({
          cart: state.cart.map((i) =>
            i.ID === itemId
              ? {
                  ...i,
                  quantity: i.quantity + 1,
                  Price: state.calculateItemPrice(i),
                }
              : i
          ),
        })),

      // ลดจำนวนสินค้าในตะกร้า
      decreaseQuantity: (itemId) =>
        set((state) => ({
          cart: state.cart
            .map((i) =>
              i.ID === itemId
                ? {
                    ...i,
                    quantity: i.quantity - 1,
                    Price: state.calculateItemPrice(i),
                  }
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
      name: 'cart-storage', // ชื่อ key ที่เก็บใน localStorage
      getStorage: () => localStorage, // ระบุว่าใช้ localStorage
    }
  )
)

export default useCartStore
