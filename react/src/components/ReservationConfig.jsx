import React, { useState, useEffect } from 'react';
import { Clock, AlertCircle } from 'lucide-react';

const API_BASE_URL = 'http://127.0.0.1:8080/api/reservation'

const ReservationConfig = () => {
  const [rules, setRules] = useState({
    gracePeriodMinutes: 0,
    preReservationMinutes: 0,
  });
  const [isEditing, setIsEditing] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchActiveRule();
  }, []);

  const fetchActiveRule = async () => {
    try {
      const token = localStorage.getItem('token'); // หรือดึง token จากที่เก็บอื่นๆ
      const response = await fetch(`${API_BASE_URL}/rules/active`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        const data = await response.json();
        setRules({
          gracePeriodMinutes: data.grace_period_minutes,
          preReservationMinutes: data.pre_reservation_minutes,
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
      const response = await fetch(`${API_BASE_URL}/rules`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          grace_period_minutes: parseInt(rules.gracePeriodMinutes),
          pre_reservation_minutes: parseInt(rules.preReservationMinutes),
        }),
      });

      if (response.ok) {
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
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto bg-white rounded-xl shadow-lg overflow-hidden">
      <div className="px-8 py-6 border-b border-gray-100">
        <h2 className="text-2xl font-semibold text-gray-800">
          ตั้งค่าการจองโต๊ะ
        </h2>
      </div>

      <div className="p-8">
        {!isEditing ? (
          <div className="space-y-6">
            <div className="rounded-lg bg-white p-6 shadow-sm border border-gray-100 hover:border-gray-200 transition-colors">
              <div className="flex items-start space-x-4">
                <div className="p-2 bg-blue-50 rounded-lg">
                  <Clock className="h-6 w-6 text-blue-500" />
                </div>
                <div className="flex-1">
                  <h3 className="font-medium text-gray-900">เวลาที่ยอมให้มาสาย</h3>
                  <p className="mt-1 text-gray-500">ระยะเวลาผ่อนผันสำหรับการมาสาย</p>
                  <p className="mt-2 text-2xl font-semibold text-blue-600">
                    {rules.gracePeriodMinutes} นาที
                  </p>
                </div>
              </div>
            </div>

            <div className="rounded-lg bg-white p-6 shadow-sm border border-gray-100 hover:border-gray-200 transition-colors">
              <div className="flex items-start space-x-4">
                <div className="p-2 bg-purple-50 rounded-lg">
                  <AlertCircle className="h-6 w-6 text-purple-500" />
                </div>
                <div className="flex-1">
                  <h3 className="font-medium text-gray-900">เวลากั้นโต๊ะก่อนลูกค้ามาถึง</h3>
                  <p className="mt-1 text-gray-500">ระยะเวลาเตรียมโต๊ะก่อนลูกค้ามาถึง</p>
                  <p className="mt-2 text-2xl font-semibold text-purple-600">
                    {rules.preReservationMinutes} นาที
                  </p>
                </div>
              </div>
            </div>

            <button
              onClick={() => setIsEditing(true)}
              className="w-full mt-6 px-4 py-3 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-lg font-medium hover:from-blue-600 hover:to-blue-700 transition-all duration-200 shadow-sm hover:shadow focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            >
              แก้ไขการตั้งค่า
            </button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-4">
              <label className="block">
                <span className="text-gray-700 font-medium">เวลาที่ยอมให้มาสาย (นาที)</span>
                <input
                  type="number"
                  min="0"
                  value={rules.gracePeriodMinutes}
                  onChange={(e) => setRules({...rules, gracePeriodMinutes: e.target.value})}
                  className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
                />
              </label>

              <label className="block">
                <span className="text-gray-700 font-medium">เวลากั้นโต๊ะก่อนลูกค้ามาถึง (นาที)</span>
                <input
                  type="number"
                  min="0"
                  value={rules.preReservationMinutes}
                  onChange={(e) => setRules({...rules, preReservationMinutes: e.target.value})}
                  className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
                />
              </label>
            </div>

            <div className="flex space-x-4 pt-4">
              <button
                type="submit"
                className="flex-1 px-4 py-3 bg-gradient-to-r from-green-500 to-green-600 text-white rounded-lg font-medium hover:from-green-600 hover:to-green-700 transition-all duration-200 shadow-sm hover:shadow focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
              >
                บันทึกการเปลี่ยนแปลง
              </button>
              <button
                type="button"
                onClick={() => setIsEditing(false)}
                className="flex-1 px-4 py-3 bg-gradient-to-r from-gray-500 to-gray-600 text-white rounded-lg font-medium hover:from-gray-600 hover:to-gray-700 transition-all duration-200 shadow-sm hover:shadow focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
              >
                ยกเลิก
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
};

export default ReservationConfig;