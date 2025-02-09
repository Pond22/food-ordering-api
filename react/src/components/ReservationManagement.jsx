import React, { useState, useEffect } from 'react';
import axios from 'axios';

const ReservationManagement = ({ isOpen, onClose }) => {
  const [reservations, setReservations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('upcoming'); // upcoming, past, cancelled

  useEffect(() => {
    if (isOpen) {
      fetchReservations();
    }
  }, [isOpen]);

  const fetchReservations = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      const response = await fetch('http://localhost:8080/api/table/reservations', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      
      const data = await response.json();
      setReservations(data);
    } catch (error) {
      console.error('Error fetching reservations:', error);
      alert('ไม่สามารถดึงข้อมูลการจองได้');
    } finally {
      setLoading(false);
    }
  };
  
  const handleCancelReservation = async (tableId, reservationId) => {
    if (!window.confirm('ยืนยันการยกเลิกการจอง?')) return;
  
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/api/v2/reservation/cancel/${reservationId}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'ไม่สามารถยกเลิกการจองได้');
      }
      
      alert('ยกเลิกการจองสำเร็จ');
      fetchReservations();
    } catch (error) {
      console.error('Error cancelling reservation:', error);
      alert(error.message || 'ไม่สามารถยกเลิกการจองได้');
    }
  };

  const filterReservations = () => {
    const now = new Date();
    return reservations.filter(reservation => {
      const reservationTime = new Date(reservation.ReservedFor);
      switch (filter) {
        case 'upcoming':
          return reservationTime > now && reservation.Status === 'active';
        case 'past':
          return reservationTime <= now || reservation.Status === 'expired';
        case 'cancelled':
          return reservation.Status === 'cancelled';
        default:
          return true;
      }
    });
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white p-6 rounded-lg shadow-xl w-[1000px] max-h-[80vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-2xl font-bold">จัดการการจองโต๊ะ</h2>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700"
          >
            ✕
          </button>
        </div>

        {/* Filter Buttons */}
        <div className="flex space-x-4 mb-6">
          <button
            className={`px-4 py-2 rounded-lg ${
              filter === 'upcoming' ? 'bg-blue-500 text-white' : 'bg-gray-200'
            }`}
            onClick={() => setFilter('upcoming')}
          >
            การจองที่กำลังจะมาถึง
          </button>
          <button
            className={`px-4 py-2 rounded-lg ${
              filter === 'past' ? 'bg-blue-500 text-white' : 'bg-gray-200'
            }`}
            onClick={() => setFilter('past')}
          >
            การจองที่ผ่านมา
          </button>
          <button
            className={`px-4 py-2 rounded-lg ${
              filter === 'cancelled' ? 'bg-blue-500 text-white' : 'bg-gray-200'
            }`}
            onClick={() => setFilter('cancelled')}
          >
            การจองที่ถูกยกเลิก
          </button>
        </div>

        {loading ? (
          <div className="text-center py-8">กำลังโหลด...</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    โต๊ะ
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    ชื่อผู้จอง
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    เบอร์โทร
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    จำนวนลูกค้า
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    วันเวลาที่จอง
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    สถานะ
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    การจัดการ
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {filterReservations().map((reservation) => {
                  const reservationTime = new Date(reservation.ReservedFor);
                  const isUpcoming = reservationTime > new Date() && reservation.Status === 'active';
                  
                  return (
                    <tr key={reservation.ID}>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {reservation.table_name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {reservation.CustomerName}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {reservation.PhoneNumber}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {reservation.GuestCount} ท่าน
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {reservationTime.toLocaleString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                          ${isUpcoming ? 'bg-green-100 text-green-800' : 
                            reservation.Status === 'cancelled' ? 'bg-red-100 text-red-800' : 
                            'bg-gray-100 text-gray-800'}`}
                        >
                          {isUpcoming ? 'กำลังจะมาถึง' : 
                           reservation.Status === 'cancelled' ? 'ยกเลิกแล้ว' : 
                           'ผ่านไปแล้ว'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        {isUpcoming && (
                          <button
                            onClick={() => handleCancelReservation(reservation.TableID, reservation.ID)}
                            className="text-red-600 hover:text-red-900"
                          >
                            ยกเลิกการจอง
                          </button>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default ReservationManagement;