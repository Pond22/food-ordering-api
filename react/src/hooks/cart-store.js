import { create } from 'zustand'
import { persist } from 'zustand/middleware'

const CACHE_TIME = 2 * 60 * 60 * 1000

const useCartStore = create(
  persist(
    (set, get) => ({
      cart: [],
      tableId: null,
      uuid: null,

      setTableData: (newTableId, newUuid) => {
        const { tableId, uuid, clearCart } = get()

        if (tableId !== newTableId || uuid !== newUuid) {
          clearCart()
        }

        set({ tableId: newTableId, uuid: newUuid })
      },

      addToCart: (item, quantity, note, selectedOptions) =>
        set((state) => {
          const timestamp = new Date().getTime()

          // ตรวจสอบรายการในตะกร้า
          const existingItem = state.cart.find(
            (i) =>
              i.menu_item_id === item.ID &&
              i.notes === note &&
              i.options.length === selectedOptions.length &&
              i.options.every((opt, index) => {
                return (
                  opt.menu_option_id ===
                    selectedOptions[index].menu_option_id &&
                  opt.name === selectedOptions[index].name &&
                  opt.price === selectedOptions[index].price
                )
              })
          )

          // ถ้ามีสินค้านี้ในตะกร้าแล้ว ให้เพิ่มจำนวน
          if (existingItem) {
            return {
              cart: state.cart.map((i) =>
                i.menu_item_id === item.ID &&
                i.notes === note &&
                JSON.stringify(i.options) === JSON.stringify(selectedOptions)
                  ? { ...i, quantity: i.quantity + quantity, timestamp }
                  : i
              ),
            }
          } else {
            // ถ้าไม่มี ให้เพิ่มสินค้าใหม่ในตะกร้า
            return {
              cart: [
                ...state.cart,
                {
                  menuItem: item,
                  menu_item_id: item.ID,
                  notes: note || '',
                  options: selectedOptions.map((opt) => ({
                    menu_option_id: opt.menu_option_id,
                    name: opt.name,
                    price: opt.price,
                  })),
                  quantity: quantity,
                  timestamp,
                },
              ],
            }
          }
        }),

      increaseQuantity: (menuItemId, note, selectedOptions) =>
        set((state) => ({
          cart: state.cart.map((i) =>
            i.menu_item_id === menuItemId &&
            i.notes === note &&
            JSON.stringify(i.options) ===
              JSON.stringify(
                selectedOptions.map((opt) => ({
                  menu_option_id: opt.menu_option_id,
                  name: opt.name, // เพิ่มชื่อ option
                  price: opt.price, // เพิ่มราคาของ option
                }))
              )
              ? { ...i, quantity: i.quantity + 1 }
              : i
          ),
        })),

      decreaseQuantity: (menuItemId, note, selectedOptions) =>
        set((state) => ({
          cart: state.cart
            .map((i) =>
              i.menu_item_id === menuItemId &&
              i.notes === note &&
              JSON.stringify(i.options) ===
                JSON.stringify(
                  selectedOptions.map((opt) => ({
                    menu_option_id: opt.menu_option_id,
                    name: opt.name, // เพิ่มชื่อ option
                    price: opt.price, // เพิ่มราคาของ option
                  }))
                )
                ? { ...i, quantity: i.quantity - 1 }
                : i
            )
            .filter((i) => i.quantity > 0),
        })),

      removeFromCart: (menuItemId, note, selectedOptions) =>
        set((state) => ({
          cart: state.cart.filter(
            (i) =>
              i.menu_item_id !== menuItemId ||
              i.notes !== note ||
              JSON.stringify(i.options) !==
                JSON.stringify(
                  selectedOptions.map((opt) => ({
                    menu_option_id: opt.menu_option_id,
                    name: opt.name, // เพิ่มชื่อ option
                    price: opt.price, // เพิ่มราคาของ option
                  }))
                )
          ),
        })),

      checkCartExpiry: () => {
        const now = new Date().getTime()
        const updatedCart = get().cart.filter(
          (item) => now - item.timestamp < CACHE_TIME
        )

        if (updatedCart.length !== get().cart.length) {
          set({ cart: updatedCart })
        }
      },

      clearCart: () => set({ cart: [] }),
    }),
    {
      name: 'cart-storage',
      getStorage: () => localStorage,
    }
  )
)

export default useCartStore
