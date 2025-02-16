import React, { useState, useEffect, useMemo } from 'react';
import axios from 'axios';
import { ShieldCheck, Plus, Search, Filter, X, User, Edit, Trash2 } from 'lucide-react';

const API_BASE_URL = `${import.meta.env.VITE_APP_API_URL}/api/member`; // URL ของ API

// กำหนดค่า role ที่อนุญาตให้ใช้งาน
const userRoles = ['staff', 'manager', 'owner'];

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
  const token = localStorage.getItem('token');
  const [isChangePasswordModalOpen, setIsChangePasswordModalOpen] = useState(false);
  const [passwordData, setPasswordData] = useState({
    currentPassword: '',
    newPassword: '',
  });
  const [isResetPasswordModalOpen, setIsResetPasswordModalOpen] = useState(false);
  const [resetPasswordData, setResetPasswordData] = useState({
    userId: null,
    newPassword: '',
  });

  // ดึงข้อมูลผู้ใช้จาก API
  useEffect(() => {
    if (!token) {
      // ถ้าไม่มี token ให้ redirect ไปหน้า login
      window.location.href = '/login';
      return;
    }
    
    const fetchUsers = async () => {
      setIsLoading(true);
      try {
        const response = await axios.get(`${API_BASE_URL}`, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });
        setUsers(response.data);
      } catch (error) {
        console.error('Error fetching users:', error);
        alert('ไม่สามารถโหลดข้อมูลผู้ใช้ได้');
      } finally {
        setIsLoading(false);
      }
    };

    fetchUsers();
  }, [token]);

  const updateUsers = (newUserList) => {
    setUsers(newUserList);
  };

  const handleAddUser = async () => {
    if (!newUser.name || !newUser.username || !newUser.password) {
      alert('กรุณากรอกข้อมูลให้ครบ');
      return;
    }

    try {
      const userData = {
        username: newUser.username,
        password: newUser.password,
        name: newUser.name,
        role: newUser.role,
      };

      const response = await axios.post(`${API_BASE_URL}`, userData, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
          'accept': 'application/json'
        }
      });
      
      const addedUser = response.data;
      setUsers([...users, addedUser]);
      setIsAddModalOpen(false);
      setNewUser({ name: '', username: '', password: '', role: 'staff' });
      alert('เพิ่มผู้ใช้สำเร็จ');
    } catch (error) {
      console.error('Error adding user:', error);
      if (error.response?.data?.error) {
        alert(error.response.data.error);
      } else if (error.response?.status === 409) {
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
      const response = await axios.delete(`${API_BASE_URL}/${userId}/delete-member`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
          'accept': 'application/json'
        }
      });
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

  const handleChangePassword = async () => {
    if (!passwordData.currentPassword || !passwordData.newPassword) {
      alert('กรุณากรอกข้อมูลให้ครบ');
      return;
    }

    try {
      const response = await axios.put(
        `${API_BASE_URL}/change-password`,
        {
          current_password: passwordData.currentPassword,
          new_password: passwordData.newPassword,
        },
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      alert('เปลี่ยนรหัสผ่านสำเร็จ');
      setIsChangePasswordModalOpen(false);
      setPasswordData({ currentPassword: '', newPassword: '' });
    } catch (error) {
      console.error('Error changing password:', error);
      if (error.response?.data?.error) {
        alert(error.response.data.error);
      } else {
        alert('เกิดข้อผิดพลาดในการเปลี่ยนรหัสผ่าน');
      }
    }
  };

  const handleResetUserPassword = async () => {
    if (!resetPasswordData.newPassword) {
      alert('กรุณากรอกรหัสผ่านใหม่');
      return;
    }

    try {
      const response = await axios.put(
        `${API_BASE_URL}/${resetPasswordData.userId}/reset-password`,
        {
          new_password: resetPasswordData.newPassword,
        },
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      alert('รีเซ็ตรหัสผ่านสำเร็จ');
      setIsResetPasswordModalOpen(false);
      setResetPasswordData({ userId: null, newPassword: '' });
    } catch (error) {
      console.error('Error resetting password:', error);
      if (error.response?.data?.error) {
        alert(error.response.data.error);
      } else {
        alert('เกิดข้อผิดพลาดในการรีเซ็ตรหัสผ่าน');
      }
    }
  };

  return (
    <div className="w-min-screen p-5  bg-gray-50  ">
      <div className="bg-white shadow-md rounded-lg overflow-hidden">
        <div className="px-6 py-4 bg-blue-600 text-white flex items-center justify-between">
          <div className="flex items-center">
            <ShieldCheck className="mr-3" />
            <h1 className="text-2xl font-bold">การจัดการผู้ใช้</h1>
          </div>
          <div className="flex space-x-2">
            <button
              onClick={() => setIsChangePasswordModalOpen(true)}
              className="bg-yellow-500 text-white px-4 py-2 rounded-md hover:bg-yellow-600 flex items-center"
            >
              <Edit className="mr-2" size={20} /> เปลี่ยนรหัสผ่าน
            </button>
            <button
              onClick={() => setIsAddModalOpen(true)}
              className="bg-green-500 text-white px-4 py-2 rounded-md hover:bg-green-600 flex items-center"
            >
              <Plus className="mr-2" size={20} /> เพิ่มผู้ใช้
            </button>
          </div>
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
                        user.role === 'owner'
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
                    <div className="flex justify-end space-x-2">
                      <button
                        onClick={() => {
                          setResetPasswordData({ userId: user.id, newPassword: '' });
                          setIsResetPasswordModalOpen(true);
                        }}
                        className="text-yellow-600 hover:bg-yellow-100 p-2 rounded"
                      >
                        <Edit size={20} />
                      </button>
                      <button
                        onClick={() => handleDeleteUser(user.id)}
                        className="text-red-600 hover:bg-red-100 p-2 rounded"
                      >
                        <Trash2 size={20} />
                      </button>
                    </div>
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

      {/* Modal สำหรับเปลี่ยนรหัสผ่าน */}
      {isChangePasswordModalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center">
          <div className="bg-white p-6 rounded-lg w-96">
            <h2 className="text-xl font-bold mb-4">เปลี่ยนรหัสผ่าน</h2>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700">รหัสผ่านปัจจุบัน</label>
              <input
                type="password"
                value={passwordData.currentPassword}
                onChange={(e) => setPasswordData({ ...passwordData, currentPassword: e.target.value })}
                className="w-full p-2 border rounded-md"
              />
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700">รหัสผ่านใหม่</label>
              <input
                type="password"
                value={passwordData.newPassword}
                onChange={(e) => setPasswordData({ ...passwordData, newPassword: e.target.value })}
                className="w-full p-2 border rounded-md"
              />
            </div>
            <div className="flex justify-end space-x-2">
              <button
                onClick={() => setIsChangePasswordModalOpen(false)}
                className="bg-gray-500 text-white px-4 py-2 rounded-md hover:bg-gray-600"
              >
                ยกเลิก
              </button>
              <button
                onClick={handleChangePassword}
                className="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600"
              >
                ยืนยัน
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal สำหรับรีเซ็ตรหัสผ่านของผู้ใช้อื่น */}
      {isResetPasswordModalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center">
          <div className="bg-white p-6 rounded-lg w-96">
            <h2 className="text-xl font-bold mb-4">รีเซ็ตรหัสผ่านผู้ใช้</h2>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700">รหัสผ่านใหม่</label>
              <input
                type="password"
                value={resetPasswordData.newPassword}
                onChange={(e) => setResetPasswordData({ ...resetPasswordData, newPassword: e.target.value })}
                className="w-full p-2 border rounded-md"
              />
            </div>
            <div className="flex justify-end space-x-2">
              <button
                onClick={() => setIsResetPasswordModalOpen(false)}
                className="bg-gray-500 text-white px-4 py-2 rounded-md hover:bg-gray-600"
              >
                ยกเลิก
              </button>
              <button
                onClick={handleResetUserPassword}
                className="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600"
              >
                ยืนยัน
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default UserManagement;
