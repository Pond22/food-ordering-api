import React, { useState, useEffect } from 'react';
import axios from 'axios';

const ReservationConfig = () => {
  const [rules, setRules] = useState({
    gracePeriodMinutes: 0,
    preReservationMinutes: 0,
  });
  const [isEditing, setIsEditing] = useState(false);
  const [loading, setLoading] = useState(true);

  // โหลด rules ที่ใช้งานอยู่
  useEffect(() => {
    fetchActiveRule();
  }, []);

  const fetchActiveRule = async () => {
    try {
      const response = await axios.get('http://localhost:8080/api/reservation/rules/active');
      if (response.data) {
        setRules({
          gracePeriodMinutes: response.data.grace_period_minutes,
          preReservationMinutes: response.data.pre_reservation_minutes,
        });
      }
      setLoading(false);
    } catch (error) {
      console.error('Error fetching rules:', error);
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const response = await axios.post('http://localhost:8080/api/reservation/rules', {
        grace_period_minutes: parseInt(rules.gracePeriodMinutes),
        pre_reservation_minutes: parseInt(rules.preReservationMinutes),
      });

      if (response.status === 200) {
        alert('บันทึกการตั้งค่าสำเร็จ');
        setIsEditing(false);
        fetchActiveRule();
      }
    } catch (error) {
      console.error('Error saving rules:', error);
      alert('เกิดข้อผิดพลาดในการบันทึกการตั้งค่า');
    }
  };

  if (loading) {
    return <div>กำลังโหลด...</div>;
  }

  return (
    <div className="p-6 max-w-2xl mx-auto bg-white rounded-lg shadow-lg">
      <h2 className="text-2xl font-bold mb-6">ตั้งค่าการจองโต๊ะ</h2>
      
      {!isEditing ? (
        <div className="space-y-4">
          <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
            <div>
              <h3 className="font-semibold">เวลาที่ยอมให้มาสาย</h3>
              <p className="text-gray-600">{rules.gracePeriodMinutes} นาที</p>
            </div>
          </div>

          <div className="flex justify-between items-center p-4 bg-gray-50 rounded">
            <div>
              <h3 className="font-semibold">เวลากั้นโต๊ะก่อนลูกค้ามาถึง</h3>
              <p className="text-gray-600">{rules.preReservationMinutes} นาที</p>
            </div>
          </div>

          <button
            onClick={() => setIsEditing(true)}
            className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            แก้ไขการตั้งค่า
          </button>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700">
              เวลาที่ยอมให้มาสาย (นาที)
            </label>
            <input
              type="number"
              min="0"
              value={rules.gracePeriodMinutes}
              onChange={(e) => setRules({...rules, gracePeriodMinutes: e.target.value})}
              className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">
              เวลากั้นโต๊ะก่อนลูกค้ามาถึง (นาที)
            </label>
            <input
              type="number"
              min="0"
              value={rules.preReservationMinutes}
              onChange={(e) => setRules({...rules, preReservationMinutes: e.target.value})}
              className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
            />
          </div>

          <div className="flex space-x-4">
            <button
              type="submit"
              className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600"
            >
              บันทึก
            </button>
            <button
              type="button"
              onClick={() => setIsEditing(false)}
              className="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600"
            >
              ยกเลิก
            </button>
          </div>
        </form>
      )}
    </div>
  );
};

export default ReservationConfig;