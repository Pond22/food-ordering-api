// ReservationCheckin.jsx
import React, { useState } from 'react';
import axios from 'axios';
import { CalendarDays } from 'lucide-react';

const API_BASE_URL = 'http://127.0.0.1:8080/api/v2/reservation'

const ReservationCheckin = ({ table, onCheckinSuccess, onUnreserve }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [qrImage, setQrImage] = useState(null);
  const [showQRModal, setShowQRModal] = useState(false);

  const handleCheckin = async () => {
    try {
      setIsLoading(true);
      
      // เรียกใช้ API endpoint ใหม่
      const response = await axios.post(
        `${API_BASE_URL}/checkin/${table.ID}`,
        {},
        {
          responseType: 'blob' // รับ response เป็น blob เพราะเป็นรูปภาพ
        }
      );

      // สร้าง URL สำหรับแสดงรูปภาพ
      const imageUrl = URL.createObjectURL(response.data);
      setQrImage(imageUrl);
      setShowQRModal(true);

      // อัพเดทสถานะโต๊ะใน parent component
      onCheckinSuccess({
        table: { ...table, Status: 'occupied' }
      });

    } catch (error) {
      console.error('Error checking in:', error);
      alert(error.response?.data?.error || 'เกิดข้อผิดพลาดในการเช็คอิน');
    } finally {
      setIsLoading(false);
    }
  };

  // Modal แสดง QR Code
  const QRModal = () => (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center z-50">
      <div className="bg-white p-6 rounded-lg shadow-lg max-w-md w-full">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-xl font-bold">QR Code สำหรับโต๊ะ {table.Name}</h3>
          <button 
            onClick={() => setShowQRModal(false)}
            className="text-gray-500 hover:text-gray-700"
          >
            ✕
          </button>
        </div>
        {qrImage && (
          <div className="flex flex-col items-center">
            <img src={qrImage} alt="QR Code" className="mb-4" />
            <p className="text-sm text-gray-600 mb-4">
              QR Code นี้จะถูกพิมพ์อัตโนมัติ
            </p>
          </div>
        )}
        <div className="flex justify-end">
          <button
            onClick={() => setShowQRModal(false)}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            ปิด
          </button>
        </div>
      </div>
    </div>
  );

  return (
    <div className="flex gap-2">
      <button
        className={`flex items-center gap-2 px-4 py-2 rounded-md ${
          isLoading 
            ? 'bg-gray-400 cursor-not-allowed' 
            : 'bg-green-500 hover:bg-green-600'
        } text-white`}
        onClick={handleCheckin}
        disabled={isLoading}
      >
        <CalendarDays className="w-5 h-5" />
        {isLoading ? 'กำลังดำเนินการ...' : 'เช็คอิน'}
      </button>
      
      <button
        className="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded-md"
        onClick={() => onUnreserve(table)}
        disabled={isLoading}
      >
        ยกเลิกการจอง
      </button>

      {showQRModal && <QRModal />}
    </div>
  );
};

export default ReservationCheckin;