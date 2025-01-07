import React, { useState, useEffect, useMemo } from 'react';
import axios from 'axios';
import { ShieldCheck, Plus, Search, Filter, X, User, Edit, Trash2 } from 'lucide-react';

const API_BASE_URL = 'http://127.0.0.1:8080/api/member'; // URL ของ API

const userRoles = ['manager', 'business owner', 'staff'];

const UserManagement = () => {
  const [users, setUsers] = useState([]);
  const [currentUser, setCurrentUser] = useState({ role: 'manager' }); // กำหนด role ของผู้ใช้งานในระบบ
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);
  const [newUser, setNewUser] = useState({
    name: '',
    username: '',
    password: '',
    role: 'staff',
  });
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  // ดึงข้อมูลผู้ใช้จาก API
  useEffect(() => {
    const fetchUsers = async () => {
      setIsLoading(true);
      try {
        const response = await axios.get(`${API_BASE_URL}`);
        setUsers(response.data);
      } catch (error) {
        console.error('Error fetching users:', error);
        alert('ไม่สามารถโหลดข้อมูลผู้ใช้ได้');
      } finally {
        setIsLoading(false);
      }
    };

    fetchUsers();
  }, []);

  const updateUsers = (newUserList) => {
    setUsers(newUserList);
  };

  const handleAddUser = async () => {
    if (!newUser.name || !newUser.username || !newUser.password) {
      alert('กรุณากรอกข้อมูลให้ครบ');
      return;
    }

    try {
      const response = await axios.post(`${API_BASE_URL}`, newUser, {
        headers: {
          'Content-Type': 'application/json',
        },
      });
      const addedUser = response.data;
      setUsers([...users, addedUser]);
      setIsAddModalOpen(false);
      setNewUser({ name: '', username: '', password: '', role: 'staff' });
      alert('เพิ่มผู้ใช้สำเร็จ');
    } catch (error) {
      console.error('Error adding user:', error);
      if (error.response?.status === 409) {
        alert('ชื่อผู้ใช้นี้มีอยู่ในระบบแล้ว');
      } else {
        alert('เกิดข้อผิดพลาดในการเพิ่มผู้ใช้');
      }
    }
  };

  // ฟังก์ชันสำหรับลบผู้ใช้
  const handleDeleteUser = async (userId) => {
    // ตรวจสอบบทบาทก่อนดำเนินการลบ
    if (currentUser.role !== 'manager') {
      alert('คุณต้องเป็นผู้จัดการเท่านั้นที่สามารถลบผู้ใช้ได้');
      return;
    }
  
    const confirmDelete = window.confirm('คุณแน่ใจหรือไม่ว่าต้องการลบผู้ใช้นี้?');
    if (!confirmDelete) return;
  
    try {
      // ส่งคำขอลบผู้ใช้ไปยัง API โดยใช้ URL ที่มีรูปแบบใหม่
      const response = await axios.delete(`${API_BASE_URL}/${userId}/delete-member`);
      console.log('User deleted:', response.data); // ตรวจสอบข้อมูลที่ตอบกลับจาก server
  
      // อัพเดทสถานะผู้ใช้หลังจากลบสำเร็จ
      const updatedUsers = users.filter((user) => user.id !== userId);
      setUsers(updatedUsers);
      alert('ลบผู้ใช้สำเร็จ');
    } catch (error) {
      console.error('Error deleting user:', error);
      if (error.response) {
        // หากมีการตอบกลับจาก server
        console.error('Response error data:', error.response.data);
        console.error('Response status:', error.response.status);
        alert('ไม่สามารถลบผู้ใช้ได้: ' + error.response.data.message || 'เกิดข้อผิดพลาด');
      } else {
        // ถ้าไม่มีการตอบกลับจาก server
        alert('ไม่สามารถลบผู้ใช้ได้: ' + error.message);
      }
    }
  };

  const filteredUsers = useMemo(() => {
    if (!users) return []; // หาก users เป็น null หรือ undefined ให้คืนค่าเป็น array ว่าง
    return users.filter((user) => {
      const matchesSearch =
        user.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        user.username.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesRole = !roleFilter || user.role === roleFilter;
  
      return matchesSearch && matchesRole;
    });
  }, [users, searchTerm, roleFilter]);

  const resetFilters = () => {
    setSearchTerm('');
    setRoleFilter('');
  };

  return (
    <div className="w-min-screen p-5  bg-gray-50  lg:ml-60">
      <div className="bg-white shadow-md rounded-lg overflow-hidden">
        <div className="px-6 py-4 bg-blue-600 text-white flex items-center justify-between">
          <div className="flex items-center">
            <ShieldCheck className="mr-3" />
            <h1 className="text-2xl font-bold">การจัดการผู้ใช้</h1>
          </div>
          <button
            onClick={() => setIsAddModalOpen(true)}
            className="bg-green-500 text-white px-4 py-2 rounded-md hover:bg-green-600 flex items-center"
          >
            <Plus className="mr-2" size={20} /> เพิ่มผู้ใช้
          </button>
        </div>

        <div className="p-4 bg-gray-100 flex space-x-4">
          <div className="relative flex-grow">
            <input
              type="text"
              placeholder="ค้นหาชื่อหรือชื่อผู้ใช้"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full p-2 pl-10 border rounded-md"
            />
            <Search className="absolute left-3 top-3 text-gray-500" size={20} />
          </div>
          <select
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            className="p-2 border rounded-md"
          >
            <option value="">ทุกบทบาท</option>
            {userRoles.map((role) => (
              <option key={role} value={role}>{role}</option>
            ))}
          </select>
          {(searchTerm || roleFilter) && (
            <button
              onClick={resetFilters}
              className="bg-red-500 text-white p-2 rounded-md hover:bg-red-600"
            >
              <X size={20} />
            </button>
          )}
        </div>

        {isLoading ? (
          <div className="text-center p-4 text-gray-500">กำลังโหลดข้อมูล...</div>
        ) : (
          <table className="w-full">
            <thead className="bg-gray-100 border-b">
              <tr>
                <th className="p-4 text-left">ชื่อ</th>
                <th className="p-4 text-left">ชื่อผู้ใช้</th>
                <th className="p-4 text-left">บทบาท</th>
                <th className="p-4 text-right">Action</th>
              </tr>
            </thead>
            <tbody>
              {filteredUsers.map((user) => (
                <tr key={user.id} className="border-b hover:bg-gray-50">
                  <td className="p-4 flex items-center">
                    <User className="mr-3 text-gray-500" />
                    {user.name}
                  </td>
                  <td className="p-4">{user.username}</td>
                  <td className="p-4">
                    <span
                      className={`px-3 py-1 rounded-full text-sm font-semibold ${
                        user.role === 'business owner'
                          ? 'bg-red-100 text-red-800'
                          : user.role === 'manager'
                          ? 'bg-yellow-100 text-yellow-800'
                          : 'bg-green-100 text-green-800'
                      }`}
                    >
                      {user.role}
                    </span>
                  </td>
                  <td className="p-4 text-right">
                    <button
                      onClick={() => handleDeleteUser(user.id)}
                      className="text-red-600 hover:bg-red-100 p-2 rounded"
                    >
                      <Trash2 size={20} />
                    </button>
                  </td>
                </tr>
              ))}
              {filteredUsers.length === 0 && (
                <tr>
                  <td colSpan="4" className="text-center p-4 text-gray-500">
                    ไม่พบผู้ใช้
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        )}
      </div>

      {/* Modal สำหรับเพิ่มผู้ใช้ */}
      {isAddModalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center">
          <div className="bg-white p-6 rounded-lg w-96">
            <h2 className="text-xl font-bold mb-4">เพิ่มผู้ใช้</h2>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700">ชื่อ</label>
              <input
                type="text"
                value={newUser.name}
                onChange={(e) => setNewUser({ ...newUser, name: e.target.value })}
                className="w-full p-2 border rounded-md"
              />
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700">ชื่อผู้ใช้</label>
              <input
                type="text"
                value={newUser.username}
                onChange={(e) => setNewUser({ ...newUser, username: e.target.value })}
                className="w-full p-2 border rounded-md"
              />
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700">รหัสผ่าน</label>
              <input
                type="password"
                value={newUser.password}
                onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
                className="w-full p-2 border rounded-md"
              />
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700">บทบาท</label>
              <select
                value={newUser.role}
                onChange={(e) => setNewUser({ ...newUser, role: e.target.value })}
                className="w-full p-2 border rounded-md"
              >
                {userRoles.map((role) => (
                  <option key={role} value={role}>{role}</option>
                ))}
              </select>
            </div>
            <div className="flex justify-end space-x-2">
              <button
                onClick={() => setIsAddModalOpen(false)}
                className="bg-gray-500 text-white px-4 py-2 rounded-md hover:bg-gray-600"
              >
                ยกเลิก
              </button>
              <button
                onClick={handleAddUser}
                className="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600"
              >
                เพิ่ม
              </button>
            </div>
          </div>
        </div>
      )}

      
    </div>
  );
};

export default UserManagement;
