import React, { useState, useMemo } from 'react';
import { User, ShieldCheck, Edit, Trash2, Plus, Search, Filter, X } from 'lucide-react';

// Function to simulate API call - in real app, this would be an actual API
const fetchUsers = () => {
  const storedUsers = localStorage.getItem('users');
  return storedUsers 
    ? JSON.parse(storedUsers) 
    : [
      { id: 1, name: 'John Doe', password: '1234', role: 'ผู้จัดการ' },
      { id: 2, name: 'Jane Smith', password: '2345', role: 'เจ้าของธุรกิจ' },
      { id: 3, name: 'Mike Johnson', password: '7412', role: 'พนักงาน' }
    ];
};

const userRoles = ['ผู้จัดการ', 'เจ้าของธุรกิจ', 'พนักงาน'];

const UserRoleManagement = () => {
  const [users, setUsers] = useState(fetchUsers());
  const [selectedUser, setSelectedUser] = useState(null);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState('');
  const [newUser, setNewUser] = useState({
    name: '',
    password: '',
    role: 'พนักงาน'
  });

  // Update localStorage whenever users change
  const updateUsers = (newUserList) => {
    setUsers(newUserList);
    localStorage.setItem('users', JSON.stringify(newUserList));
  };

  // Function to handle role change
  const handleRoleChange = (userId, newRole) => {
    const updatedUsers = users.map(user => 
      user.id === userId ? { ...user, role: newRole } : user
    );
    updateUsers(updatedUsers);
    setIsEditModalOpen(false);
  };

  // Function to open edit modal
  const openEditModal = (user) => {
    setSelectedUser(user);
    setIsEditModalOpen(true);
  };

  // Function to delete user
const handleDeleteUser = (userId) => {
    // ค้นหาผู้ใช้ที่ต้องการลบ
    const user = users.find(user => user.id === userId);
  
    if (user) {
      // แสดงข้อความยืนยันที่มีชื่อผู้ใช้
      const confirmDelete = window.confirm(`คุณแน่ใจหรือไม่ว่าต้องการลบผู้ใช้: ${user.name}?`);
      if (confirmDelete) {
        // อัปเดตรายการผู้ใช้หลังจากลบผู้ใช้ที่เลือก
        const updatedUsers = users.filter(user => user.id !== userId);
        updateUsers(updatedUsers);
      }
    }
  };

  // Function to add new user
  const handleAddUser = () => {
    // Validate input
    if (!newUser.name || !newUser.password) {
      alert('กรุณากรอกชื่อและอีเมลให้ครบ');
      return;
    }

    // Check for duplicate email
    const isDuplicate = users.some(user => user.em === newUser.email);
    if (isDuplicate) {
      alert('อีเมลนี้มีอยู่ในระบบแล้ว');
      return;
    }

    const newUserId = users.length > 0 
      ? Math.max(...users.map(u => u.id)) + 1 
      : 1;

    const userToAdd = {
      id: newUserId,
      ...newUser
    };

    const updatedUsers = [...users, userToAdd];
    updateUsers(updatedUsers);
    setIsAddModalOpen(false);
    
    // Reset new user form
    setNewUser({
      name: '',
      email: '',
      role: 'พนักงาน'
    });
  };

  // Filtered and searched users
  const filteredUsers = useMemo(() => {
    return users.filter(user => {
      const matchesSearch = user.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                            user.password.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesRole = !roleFilter || user.role === roleFilter;
      
      return matchesSearch && matchesRole;
    });
  }, [users, searchTerm, roleFilter]);

  // Reset all filters
  const resetFilters = () => {
    setSearchTerm('');
    setRoleFilter('');
  };

  return (
    <div className="container mx-auto p-6 bg-gray-50 ">
      <div className="bg-white shadow-md rounded-lg overflow-hidden ">
        {/* Header with Search and Add User */}
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

        {/* Search and Filter Section */}
        <div className="p-4 bg-gray-100 flex space-x-4">
          <div className="relative flex-grow">
            <input 
              type="text"
              placeholder="ค้นหาชื่อหรืออีเมล"
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
            {userRoles.map(role => (
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

        {/* User Table */}
        <table className="w-full ">
          <thead className="bg-gray-100 border-b ">
            <tr>
              <th className="p-4 text-left">ชื่อ</th>
              <th className="p-4 text-left">รหัสผ่าน</th>
              <th className="p-4 text-left">บทบาท</th>
              <th className="p-4 text-right">การดำเนินการ</th>
            </tr>
          </thead>
          <tbody>
            {filteredUsers.map(user => (
              <tr key={user.id} className="border-b hover:bg-gray-50">
                <td className="p-4 flex items-center">
                  <User className="mr-3 text-gray-500" />
                  {user.name}
                </td>
                <td className="p-4">{user.password}</td>
                <td className="p-4">
                  <span className={`
                    px-3 py-1 rounded-full text-sm font-semibold
                    ${user.role === 'เจ้าของธุรกิจ' ? 'bg-red-100 text-red-800' : 
                      user.role === 'ผู้จัดการ' ? 'bg-yellow-100 text-yellow-800' : 
                      'bg-green-100 text-green-800'}
                  `}>
                    {user.role}
                  </span>
                </td>
                <td className="p-4 text-right">
                  <button 
                    onClick={() => openEditModal(user)}
                    className="mr-2 text-blue-600 hover:bg-blue-100 p-2 rounded"
                  >
                    <Edit size={20} />
                  </button>
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

        {/* Edit Role Modal */}
        {isEditModalOpen && selectedUser && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
            <div className="bg-white rounded-lg shadow-xl p-6 w-96">
              <h2 className="text-xl font-bold mb-4">แก้ไขบทบาทผู้ใช้</h2>
              <div className="space-y-3">
                <div className="flex items-center">
                  <User className="mr-3 text-gray-500" />
                  <div>
                    <p className="font-semibold">{selectedUser.name}</p>
                    <p className="text-sm text-gray-500">{selectedUser.email}</p>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    เลือกบทบาท
                  </label>
                  <select 
                    value={selectedUser.role}
                    onChange={(e) => handleRoleChange(selectedUser.id, e.target.value)}
                    className="w-full p-2 border rounded-md"
                  >
                    {userRoles.map(role => (
                      <option key={role} value={role}>{role}</option>
                    ))}
                  </select>
                </div>
              </div>
              <div className="flex justify-end mt-4 space-x-2">
                <button 
                  onClick={() => setIsEditModalOpen(false)}
                  className="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300"
                >
                  ยกเลิก
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Add User Modal */}
        {isAddModalOpen && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
            <div className="bg-white rounded-lg shadow-xl p-6 w-96">
              <h2 className="text-xl font-bold mb-4">เพิ่มผู้ใช้ใหม่</h2>
              <div className="space-y-3">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    ชื่อ
                  </label>
                  <input 
                    type="text"
                    value={newUser.name}
                    onChange={(e) => setNewUser({...newUser, name: e.target.value})}
                    className="w-full p-2 border rounded-md"
                    placeholder="กรอกชื่อ"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    รหัสผ่าน
                  </label>
                  <input 
                    type="password"
                    value={newUser.password}
                    onChange={(e) => setNewUser({...newUser, password: e.target.value})}
                    className="w-full p-2 border rounded-md"
                    placeholder="กรอกรหัสผ่าน"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    บทบาท
                  </label>
                  <select 
                    value={newUser.role}
                    onChange={(e) => setNewUser({...newUser, role: e.target.value})}
                    className="w-full p-2 border rounded-md"
                  >
                    {userRoles.map(role => (
                      <option key={role} value={role}>{role}</option>
                    ))}
                  </select>
                </div>
              </div>
              <div className="flex justify-end mt-4 space-x-2">
                <button 
                  onClick={() => setIsAddModalOpen(false)}
                  className="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300 mr-2"
                >
                  ยกเลิก
                </button>
                <button 
                  onClick={handleAddUser}
                  className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600"
                >
                  เพิ่มผู้ใช้
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default UserRoleManagement;