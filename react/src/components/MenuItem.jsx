import React, { useState, useEffect } from "react";
import styles from "../styles/MenuItem.module.css";
import { Image, Minus, Plus, PlusIcon } from "lucide-react";
import useCartStore from "../hooks/cart-store";

const MenuItem = ({ item }) => {
  const [isPopupOpen, setIsPopupOpen] = useState(false); // Popup state
  const [note, setNote] = useState(""); // Note for the kitchen
  const [selectedOptions, setSelectedOptions] = useState({}); // Store selected options for each group
  const [quantity, setQuantity] = useState(1);

  const { addToCart } = useCartStore();

  const handleIncrease = () => setQuantity((prev) => prev + 1); // เพิ่มจำนวน
  const handleDecrease = () => {
    setQuantity((prev) => (prev > 1 ? prev - 1 : 1));
  };

  // Toggle popup visibility
  const togglePopup = () => setIsPopupOpen((prev) => !prev);

  // Handle adding item to cart
  const handleAddToCart = () => {
    if (quantity > 0) {
      addToCart(item, quantity, note, selectedOptions);
      setNote(""); // Clear the note
      setSelectedOptions({}); // Clear selected options
      setQuantity(1);
      togglePopup();
    }
  };

  // Handle option selection change
  const handleOptionChange = (groupName, optionName, price) => {
    setSelectedOptions((prev) => ({
      ...prev,
      [groupName]: { optionName, price },
    }));
  };

  return (
    <div className={"bg-white border rounded-md shadow-sm"}>
      {/* Menu Item Display */}
      {item.Image ? (
        <img
          src={`data:image/png;base64,${item.Image}`}
          alt={item.Name}
          className={
            "h-28 sm:h-36 w-full rounded-tl-md rounded-tr-md object-cover"
          }
          onClick={togglePopup} // When clicked, toggle the popup
        />
      ) : (
        <div className={styles.placeholderImage}>No Image Available</div>
      )}

      <div className={"p-2 flex flex-col gap-4"}>
        <h3 className="text-lg text-black font-semibold" onClick={togglePopup}>
          {item.Name}
        </h3>
        {/* <h4>{item.Description}</h4> */}

        <div className={"flex justify-between items-center"}>
          <p className="text-base text-black font-semibold">{item.Price} บาท</p>
          <button
            onClick={togglePopup}
            className=" bg-green-400 hover:bg-green-500 rounded-full p-1 text-white"
          >
            <Plus className="size-5" />
          </button>
        </div>
      </div>

      {isPopupOpen && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50 p-4">
          <div className="bg-white border border-gray/50 rounded-xl w-full max-w-xl overflow-y-auto max-h-screen">
            <div className="relative">
              <button
                onClick={togglePopup}
                className="absolute right-4 top-4 text-white/70 hover:text-white z-10 
                bg-black/50 rounded-full p-2"
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

              {Image ? (
                <img
                  src={`data:image/png;base64,${item.Image}`}
                  alt={item.Name || "Menu item"}
                  className="w-full aspect-video object-cover"
                />
              ) : (
                <div className="w-full aspect-video bg-zinc-900 flex items-center justify-center">
                  <span className="text-gold/50 text-lg">
                    No Image Available
                  </span>
                </div>
              )}
            </div>

            <div className="p-6">
              <div className="flex justify-between">
                <h3 className="text-black text-2xl font-medium mb-2">
                  {item.Name || "Untitled Item"}
                </h3>
                <p className="text-black text-xl font-semibold mb-6">
                  {item.Price || 0}&nbsp;บาท
                </p>
              </div>
              <p className="text-gray-400 mb-4">
                {item.Description || "No description available"}
              </p>

              {/* Displaying Option Groups if they exist */}
              {item.OptionGroups && item.OptionGroups.length > 0 && (
                <div className="mb-6">
                  {item.OptionGroups.map((group) => (
                    <div key={group.ID}>
                      <h4 className="text-black text-lg font-medium">
                        {group.Name || "Option Group"}
                      </h4>
                      <div className="flex flex-col gap-2 mt-2">
                        {group.Options.map((option) => (
                          <div
                            key={option.ID}
                            className="flex justify-start items-center gap-3 text-black/70 text-base"
                          >
                            <input
                              type="radio"
                              name={group.Name}
                              value={option.Name}
                              className=""
                            />
                            {option.Name} (+{option.Price}฿)
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}

              <div className="mb-4">
                <label className="block text-black/50 text-sm font-medium mb-2">
                  Special Instructions
                </label>
                <textarea
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  placeholder="Add your special request here..."
                  className="w-full px-4 py-3 bg-gray-100 border border-gray/30 rounded-lg 
                   placeholder-gray-500"
                  rows="3"
                />
              </div>

              {/* ปุ่มเพิ่ม-ลด */}
              <div className="my-4 flex justify-center items-center gap-6">
                <button
                  className="bg-gray-200 hover:bg-gray-300 disabled:opacity-70 disabled:hover:bg-gray-200 p-2 rounded-md text-black"
                  onClick={handleDecrease}
                  disabled={quantity == 1}
                >
                  <Minus className="size-4" />
                </button>
                <span className="text-lg font-semibold">{quantity}</span>
                <button
                  className="bg-gray-200 hover:bg-gray-300 p-2 rounded-md text-black"
                  onClick={handleIncrease}
                >
                  <PlusIcon className="size-4" />
                </button>
              </div>

              <button
                onClick={handleAddToCart}
                className="bg-blue-500 hover:bg-blue-600 text-white py-3 w-full rounded-lg px-4 font-medium transition-colors"
              >
                Add to Cart
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default MenuItem;
