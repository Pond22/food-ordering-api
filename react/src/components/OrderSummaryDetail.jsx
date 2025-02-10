import React, { useState } from 'react'
import axios from 'axios'

const OrderSummaryDetail = ({ 
  billableItems, 
  user,
  tableID,
  onCancelItem
}) => {
  // State for tracking items pending deletion
  const [pendingDelete, setPendingDelete] = useState({
    itemId: null,
    optionId: null,
    timestamp: null
  });

  // Add click outside listener
  React.useEffect(() => {
    const handleClickOutside = (event) => {
      if (pendingDelete.itemId && 
          !event.target.closest('button[data-delete-button="true"]')) {
        setPendingDelete({ itemId: null, optionId: null, timestamp: null });
      }
    };

    document.addEventListener('click', handleClickOutside);
    return () => {
      document.removeEventListener('click', handleClickOutside);
    };
  }, [pendingDelete.itemId]);

  const groupByPromotion = (items) => {
    const promotionGroups = {}
    const nonPromoItems = []

    items.forEach(item => {
      if (item.promotion) {
        if (!promotionGroups[item.promotion.id]) {
          promotionGroups[item.promotion.id] = {
            promotion: item.promotion,
            items: [],
            totalPrice: 0
          }
        }
        promotionGroups[item.promotion.id].items.push(item)
        promotionGroups[item.promotion.id].totalPrice += item.item_total
      } else {
        nonPromoItems.push(item)
      }
    })

    return {
      promotionGroups,
      nonPromoItems
    }
  }

  const handleDeleteClick = (itemId, optionId = null) => {
    // If there's already a pending delete and it's the same item/option
    if (pendingDelete.itemId === itemId && pendingDelete.optionId === optionId) {
      // Check if within 3 seconds
      if (Date.now() - pendingDelete.timestamp < 3000) {
        // Proceed with actual deletion
        if (optionId) {
          handleCancelOption(
            billableItems.find(item => item.id === itemId),
            { id: optionId }
          );
        } else {
          handleCancelItem(billableItems.find(item => item.id === itemId));
        }
      }
      // Reset pending state
      setPendingDelete({ itemId: null, optionId: null, timestamp: null });
    } else {
      // Set new pending delete
      setPendingDelete({
        itemId,
        optionId,
        timestamp: Date.now()
      });
      
      // Auto-reset after 3 seconds
      setTimeout(() => {
        setPendingDelete(current => {
          if (current.itemId === itemId && current.optionId === optionId) {
            return { itemId: null, optionId: null, timestamp: null };
          }
          return current;
        });
      }, 3000);
    }
  };

  const handleCancelOption = async (item, option) => {
    try {
      const token = localStorage.getItem('token')
      const response = await axios.post(
        'http://localhost:8080/api/orders/finalize', 
        {
          table_id: tableID,
          staff_id: user.id,
          reason: 'ยกเลิกตัวเลือกเสริม',
          options: [{
            order_item_id: item.id,
            option_id: option.id
          }]
        },
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      )
  
      if (response.data.message) {
        onCancelItem && onCancelItem(
          billableItems.map(currentItem => 
            currentItem.id === item.id 
              ? {...currentItem, options: currentItem.options.filter(opt => opt.id !== option.id)} 
              : currentItem
          )
        )
      }
    } catch (error) {
      console.error('Error cancelling option:', error)
      console.log('Error response:', error.response?.data)
      alert(error.response?.data?.error || 'ไม่สามารถยกเลิกตัวเลือกได้')
    }
  }

  const handleCancelItem = async (item) => {
    try {
      const token = localStorage.getItem('token')
      const response = await axios.post(
        'http://localhost:8080/api/orders/finalize', 
        {
          table_id: tableID,
          staff_id: user.id,
          reason: 'ยกเลิกรายการอาหาร',
          items: [{
            order_item_id: item.id,
            quantity: 1  // ลบทีละ 1 รายการเท่านั้น
          }]
        },
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      )

      if (response.data.message) {
        onCancelItem && onCancelItem(
          billableItems.map(currentItem => {
            if (currentItem.id === item.id) {
              // ถ้าเหลือ 1 ชิ้น ให้ลบรายการออก
              if (currentItem.quantity <= 1) {
                return null;
              }
              // ถ้ามากกว่า 1 ชิ้น ให้ลดจำนวนลง 1
              return {
                ...currentItem,
                quantity: currentItem.quantity - 1
              };
            }
            return currentItem;
          }).filter(Boolean) // Remove null items
        );
      }
    } catch (error) {
      console.error('Error cancelling item:', error)
      alert(error.response?.data?.error || 'ไม่สามารถยกเลิกรายการได้')
    }
  }

  const DeleteButton = ({ itemId, optionId = null }) => {
    const isPending = pendingDelete.itemId === itemId && pendingDelete.optionId === optionId;
    
    return (
      <button 
        data-delete-button="true"
        onClick={(e) => {
          e.stopPropagation();
          handleDeleteClick(itemId, optionId);
        }}
        className={`
          transition-all duration-200 px-3 py-1 rounded-full text-sm font-medium
          ${isPending 
            ? 'bg-red-500 text-white hover:bg-red-600' 
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          }
          flex items-center space-x-1
        `}
      >
        <span>{isPending ? 'ยืนยันการลบ' : 'ลบ'}</span>
      </button>
    );
  };

  const renderItem = (item, isPromoItem = false) => (
    <div 
      key={item.id} 
      className="bg-white rounded-lg p-4 border border-gray-200 transition-all hover:shadow-md"
    >
      <div className="flex justify-between items-center mb-2">
        <div className="flex items-center space-x-3">
          <div className="bg-blue-500 text-white rounded-full w-8 h-8 flex items-center justify-center font-semibold shadow-sm">
            {item.quantity}
          </div>
          <div>
            <h3 className="text-base font-medium text-gray-800 flex items-center">
              {item.name}
              {item.promotion && (
                <span className="ml-2 text-xs text-green-600 bg-green-50 px-2 py-1 rounded-full">
                  {item.promotion.name}
                </span>
              )}
            </h3>
            {item.notes && (
              <p className="text-xs text-gray-500 italic">
                {item.notes}
              </p>
            )}
          </div>
        </div>
        <div className="flex items-center space-x-3">
          {!item.promotion && (
            <span className="text-base font-semibold text-gray-900">
              {item.price.toLocaleString()} ฿
            </span>
          )}
          <DeleteButton itemId={item.id} />
        </div>
      </div>
      
      {item.options && item.options.length > 0 && (
        <div className="pl-11 mt-2 space-y-2">
          {item.options.map((option, index) => (
            <div 
              key={index} 
              className="flex justify-between items-center text-sm text-gray-600 border-l-2 border-gray-200 pl-3 py-1"
            >
              <span>{option.name}</span>
              {option.price > 0 && (
                <div className="flex items-center space-x-3">
                  <span className="font-medium text-green-600">
                    +{(option.price * option.quantity).toLocaleString()} ฿
                  </span>
                  <DeleteButton itemId={item.id} optionId={option.id} />
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )

  const { promotionGroups, nonPromoItems } = groupByPromotion(billableItems)

  return (
    <div className="space-y-4">
      {Object.values(promotionGroups).map((promoGroup) => (
        <div 
          key={promoGroup.promotion.id} 
          className="bg-green-50 rounded-lg p-4 border border-green-100 shadow-sm"
        >
          <div className="flex justify-between items-center mb-3">
            <div className="flex items-center">
              <span className="text-base font-semibold text-green-700">
                {promoGroup.promotion.name}
              </span>
            </div>
            <span className="text-base font-semibold text-green-700">
              {promoGroup.totalPrice.toLocaleString()} ฿
            </span>
          </div>
          <div className="space-y-2">
            {promoGroup.items.map(item => renderItem(item, true))}
          </div>
        </div>
      ))}

      <div className="space-y-2">
        {nonPromoItems.map(renderItem)}
      </div>
    </div>
  )
}

export default OrderSummaryDetail