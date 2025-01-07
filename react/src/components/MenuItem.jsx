import React, { useState, useEffect } from "react";
import styles from "../styles/MenuItem.module.css";
import { Image, Plus, Search, Filter, X, Ellipsis, Edit, Trash2 } from 'lucide-react';

const MenuItem = ({ item, addToCart, groupOptions }) => {
  const [isPopupOpen, setIsPopupOpen] = useState(false); // Popup state
  const [note, setNote] = useState(""); // Note for the kitchen
  const [selectedOptions, setSelectedOptions] = useState({}); // Store selected options for each group

  // Toggle popup visibility
  const togglePopup = () => {
    setIsPopupOpen((prev) => !prev);
  };

  // Handle adding item to cart
  const handleAddToCart = () => {
    addToCart(item, note, selectedOptions); // Pass item, note, and selected options to parent
    setNote(""); // Clear the note
    setSelectedOptions({}); // Clear selected options
    togglePopup(); // Close popup
  };

  // Handle option selection change
  const handleOptionChange = (groupName, optionName, price) => {
    setSelectedOptions((prev) => ({
      ...prev,
      [groupName]: { optionName, price },
    }));
  };

  return (
    <div className={styles.menuItem}>
      {/* Menu Item Display */}
      {item.Image ? (
        <img
          src={`data:image/png;base64,${item.Image}`}
          alt={item.Name}
          className={styles.image}
          onClick={togglePopup} // When clicked, toggle the popup
        />
      ) : (
        <div className={styles.placeholderImage}>No Image Available</div>
      )}

      <div className={styles.details}>
        <h3 onClick={togglePopup}>{item.Name}</h3>
        {/* <h4>{item.Description}</h4> */}
        
          
          <div className={styles.quantityControl}>
          <p>{item.Price}฿</p>
          <button onClick={togglePopup} className="px-2 py-1 flex ">
            <Plus className="w-8 h-5"/>
          </button></div>
          </div>
      

      {isPopupOpen && (
        <div className="fixed inset-0 bg-black/90 flex items-center justify-center z-50 p-4">
          <div className="bg-white border border-gold/50 rounded-xl w-full max-w-xl overflow-y-auto max-h-screen">
            <div className="relative">
              <button
                onClick={togglePopup}
                className="absolute right-4 top-4 text-white/70 hover:text-white z-10 
                bg-black/50 rounded-full p-2"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
              
              {Image ? (
                <img
                  src={`data:image/png;base64,${item.Image}`}
                  alt={item.Name|| 'Menu item'}
                  className="w-full aspect-video object-cover"
                />
              ) : (
                <div className="w-full aspect-video bg-zinc-900 flex items-center justify-center">
                  <span className="text-gold/50 text-lg">No Image Available</span>
                </div>
              )}
            </div>

            <div className="p-6">
              <div className="flex justify-between">
              <h3 className="text-black text-2xl font-medium mb-2">{item.Name|| 'Untitled Item'}</h3>
              <p className="text-red-500 text-xl font-semibold mb-6">{item.Price || 0}฿</p></div>
              <p className="text-gray-400 mb-4">{item.Description || 'No description available'}</p>

              {/* Displaying Option Groups if they exist */}
        {item.OptionGroups && item.OptionGroups.length > 0 && (
          <div className="mb-6">
            {item.OptionGroups.map((group) => (
              <div key={group.ID} className="mb-4">
                <h4 className="text-black font-medium">{group.Name || 'Option Group'}</h4>
                {group.Options.map((option) => (
                  <label key={option.ID} className="block text-black/70 text-sm">
                    <input
                      type="radio"
                      name={group.Name}
                      value={option.Name}
                      className="mr-2"
                    />
                    {option.Name} (+{option.Price}฿)
                  </label>
                ))}
              </div>
            ))}
          </div>
        )}
          
              <div className="mb-6">
                <label className="block text-black/50 text-sm font-medium mb-2">
                  Special Instructions
                </label>
                <textarea
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  placeholder="Add your special request here..."
                  className="w-full px-4 py-3 bg-gray-100 border border-gold/30 rounded-lg 
                   placeholder-gray-500 focus:border-gold focus:ring-1 focus:ring-gold"
                  rows="3"
                />
              </div>

              <button
                onClick={handleAddToCart}
                className="w-full bg-red-600 hover:bg-red-700 text-white py-3 rounded-lg 
                font-medium transition-colors"
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
