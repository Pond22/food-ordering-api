import { create } from "zustand";
import { persist } from "zustand/middleware";

const CACHE_TIME = 2 * 60 * 60 * 1000;

const useCartStore = create(
  persist(
    (set, get) => ({
      cart: [],
      tableId: null,
      uuid: null,

      setTableData: (newTableId, newUuid) => {
        const { tableId, uuid, clearCart } = get();

        if (tableId !== newTableId || uuid !== newUuid) {
          clearCart();
        }

        set({ tableId: newTableId, uuid: newUuid });
      },

      addToCart: (item, quantity, note, selectedOptions) =>
        set((state) => {
          const timestamp = new Date().getTime();

          const existingItem = state.cart.find(
            (i) =>
              i.menu_item_id === item.ID &&
              i.notes === note &&
              JSON.stringify(i.options) ===
                JSON.stringify(
                  selectedOptions.map((opt) => ({
                    menu_option_id: opt.menu_option_id,
                  }))
                )
          );

          if (existingItem) {
            // อัปเดตจำนวนสินค้า หากมีอยู่แล้ว
            return {
              cart: state.cart.map((i) =>
                i.menuItem === item &&
                i.menu_item_id === item.ID &&
                i.notes === note &&
                JSON.stringify(i.options) ===
                  JSON.stringify(
                    selectedOptions.map((opt) => ({
                      menu_option_id: opt.menu_option_id,
                    }))
                  )
                  ? { ...i, quantity: i.quantity + quantity, timestamp }
                  : i
              ),
            };
          } else {
            // เพิ่มเป็นรายการใหม่ ถ้าเป็นสินค้าคนละตัว (note หรือ options ต่างกัน)
            return {
              cart: [
                ...state.cart,
                {
                  menuItem: item,
                  menu_item_id: item.ID,
                  notes: note || "",
                  options: selectedOptions.map((opt) => ({
                    menu_option_id: opt.menu_option_id, 
                  })),
                  quantity: quantity,
                  timestamp,
                },
              ],
            };
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
                  }))
                )
          ),
        })),

      checkCartExpiry: () => {
        const now = new Date().getTime();
        const updatedCart = get().cart.filter(
          (item) => now - item.timestamp < CACHE_TIME
        );

        if (updatedCart.length !== get().cart.length) {
          set({ cart: updatedCart });
        }
      },

      clearCart: () => set({ cart: [] }),
    }),
    {
      name: "cart-storage", 
      getStorage: () => localStorage,
    }
  )
);

export default useCartStore;
