import React, { useState } from 'react';
import axios from 'axios';
import { Printer } from 'lucide-react';

const API_BASE_URL = 'http://127.0.0.1:8080/api/qr'


const QRCodeReprint = ({ tableId, uuid }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleReprint = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await axios.post(
        `${API_BASE_URL}/reprint/${tableId}`,
        { uuid }
      );

      if (response.status === 200) {
        alert('สั่งพิมพ์ QR Code สำเร็จ');
      }
    } catch (err) {
      setError(err.response?.data?.error || 'เกิดข้อผิดพลาดในการพิมพ์ QR Code');
      alert(err.response?.data?.error || 'เกิดข้อผิดพลาดในการพิมพ์ QR Code');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative group">
      <button
        onClick={handleReprint}
        disabled={loading}
        className={`bg-blue-500 text-white p-2 rounded-full hover:bg-blue-600 ${
          loading ? 'opacity-50 cursor-not-allowed' : ''
        }`}
      >
        <Printer className="w-5 h-5" />
      </button>
      <div className="absolute left-1/2 -translate-x-1/2 w-32 bottom-full mb-2 hidden group-hover:block bg-gray-700 text-white text-sm py-1 px-2 rounded-md shadow-lg text-center">
        พิมพ์ QR Code ใหม่
      </div>
    </div>
  );
};

export default QRCodeReprint;