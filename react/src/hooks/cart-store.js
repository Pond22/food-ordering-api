import { create } from 'zustand'
import { persist } from 'zustand/middleware'

const CACHE_TIME = 2 * 60 * 60 * 1000

const useCartStore = create(
  persist(
    (set, get) => ({
      cart: [],
      promotions: [],
      tableId: null,
      uuid: null,

      setTableData: (newTableId, newUuid) => {
        const { tableId, uuid, clearCart } = get()

        if (tableId !== newTableId || uuid !== newUuid) {
          clearCart()
        }

        set({ tableId: newTableId, uuid: newUuid })
      },

      // ฟังก์ชันเพิ่มสินค้าลงในตะกร้า
      addToCart: (item, quantity, note, selectedOptions) =>
        set((state) => {
          const timestamp = new Date().getTime()

          // ตรวจสอบรายการในตะกร้าว่ามีสินค้านี้หรือไม่
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

      addPromotion: (promotion, quantity) => {
        set((state) => {
          const timestamp = new Date().getTime()

          // ตรวจสอบว่ามีโปรโมชั่นนี้ในตะกร้าหรือไม่
          const existingPromotion = state.promotions.find(
            (p) => p.promotion_id === promotion.ID
          )

          if (existingPromotion) {
            // ตรวจสอบว่า selectedOptions มีการเปลี่ยนแปลงหรือไม่
            const isSelectedOptionsSame =
              JSON.stringify(existingPromotion.selectedOptions) ===
              JSON.stringify(promotion.selectedOptions || [])

            if (isSelectedOptionsSame) {
              // ถ้า selectedOptions ไม่มีการเปลี่ยนแปลง, รวมจำนวนโปรโมชั่นเข้าด้วยกัน
              return {
                promotions: state.promotions.map((p) =>
                  p.promotion_id === promotion.ID &&
                  JSON.stringify(p.selectedOptions) ===
                    JSON.stringify(promotion.selectedOptions || [])
                    ? { ...p, quantity: p.quantity + quantity, timestamp }
                    : p
                ),
              }
            } else {
              // ถ้ามีการเปลี่ยนแปลงใน selectedOptions, เพิ่มเป็นโปรโมชันใหม่
              return {
                promotions: [
                  ...state.promotions,
                  {
                    promotion_id: promotion.ID,
                    promotion: promotion,
                    quantity: quantity,
                    timestamp,
                    selectedOptions: promotion.selectedOptions || [], // เพิ่ม selectedOptions ใหม่
                  },
                ],
              }
            }
          } else {
            // ถ้าไม่มีโปรโมชั่นนี้ในตะกร้า, เพิ่มโปรโมชั่นใหม่
            return {
              promotions: [
                ...state.promotions,
                {
                  promotion_id: promotion.ID,
                  promotion: promotion,
                  quantity: quantity,
                  timestamp,
                  selectedOptions: promotion.selectedOptions || [], // เพิ่ม selectedOptions ใหม่
                },
              ],
            }
          }
        })
      },

      // ฟังก์ชันเพิ่มจำนวนสินค้าในตะกร้า
      increaseQuantity: (menuItemId, note, selectedOptions) =>
        set((state) => ({
          cart: state.cart.map((i) =>
            i.menu_item_id === menuItemId &&
            i.notes === note &&
            JSON.stringify(i.options) ===
              JSON.stringify(
                selectedOptions.map((opt) => ({
                  menu_option_id: opt.menu_option_id,
                  name: opt.name,
                  price: opt.price,
                }))
              )
              ? { ...i, quantity: i.quantity + 1 }
              : i
          ),
        })),

      // ฟังก์ชันลดจำนวนสินค้าในตะกร้า
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
                    name: opt.name,
                    price: opt.price,
                  }))
                )
                ? { ...i, quantity: i.quantity - 1 }
                : i
            )
            .filter((i) => i.quantity > 0),
        })),

      // ฟังก์ชันลบสินค้าออกจากตะกร้า
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
                    name: opt.name,
                    price: opt.price,
                  }))
                )
          ),
        })),

      // ฟังก์ชันเพิ่มจำนวนโปรโมชันในตะกร้า
      increaseQuantityPromo: (promotion_id, quantity, selectedOptions) =>
        set((state) => ({
          promotions: state.promotions.map((p) =>
            p.promotion_id === promotion_id &&
            JSON.stringify(p.selectedOptions) ===
              JSON.stringify(selectedOptions) // ตรวจสอบ selectedOptions
              ? { ...p, quantity: p.quantity + quantity }
              : p
          ),
        })),

      // ฟังก์ชันลดจำนวนโปรโมชันในตะกร้า
      decreaseQuantityPromo: (promotion_id, quantity, selectedOptions) =>
        set((state) => ({
          promotions: state.promotions
            .map((p) =>
              p.promotion_id === promotion_id &&
              JSON.stringify(p.selectedOptions) ===
                JSON.stringify(selectedOptions) // ตรวจสอบ selectedOptions
                ? { ...p, quantity: Math.max(0, p.quantity - quantity) } // ควบคุมไม่ให้ quantity ต่ำกว่า 0
                : p
            )
            .filter((p) => p.quantity > 0), // ลบโปรโมชันที่มีจำนวนเป็น 0 ออก
        })),

      // ฟังก์ชันลบสินค้าออกจากตะกร้า
      removeFromCartPromo: (promotion) =>
        set((state) => ({
          promotions: state.promotions.filter(
            (p) => p.promotion_id !== promotion.ID
          ),
        })),

      // ฟังก์ชันตรวจสอบการหมดอายุของข้อมูลในตะกร้า
      checkCartExpiry: () => {
        const now = new Date().getTime()
        const updatedCart = get().cart.filter(
          (item) => now - item.timestamp < CACHE_TIME
        )

        if (updatedCart.length !== get().cart.length) {
          set({ cart: updatedCart })
        }
      },

      // ฟังก์ชันล้างข้อมูลในตะกร้า
      clearCart: () => set({ cart: [] }),
    }),
    {
      name: 'cart-storage',
      getStorage: () => localStorage,
    }
  )
)

export default useCartStore
