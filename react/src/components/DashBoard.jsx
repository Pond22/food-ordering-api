import React, { useState } from 'react';
import { 
  BarChart, Bar, LineChart, Line, 
  XAxis, YAxis, CartesianGrid, 
  Tooltip, Legend, ResponsiveContainer, 
  PieChart, Pie, Cell 
} from 'recharts';
import { 
  TrendingUp, 
  Calendar, 
  RefreshCcw, 
  Clock 
} from 'lucide-react';

const generateMockSalesData = (period) => {
  const data = [];
  const labels = {
    'daily': ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    'weekly': ['Week 1', 'Week 2', 'Week 3', 'Week 4'],
    'monthly': ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
    'yearly': ['2022', '2023']
  };

  const baseSales = period === 'daily' ? 1000 : 
                    period === 'weekly' ? 5000 : 
                    period === 'monthly' ? 20000 : 100000;

  labels[period].forEach((label, index) => {
    const variation = Math.random() * 0.3 - 0.15; // -15% to +15%
    data.push({
      name: label,
      sales: Math.round(baseSales * (1 + variation)),
      target: baseSales,
      profit: Math.round(baseSales * (0.3 + variation)),
      expenses: Math.round(baseSales * (0.2 - variation))
    });
  });

  return data;
};

const generateProductSalesData = () => {
  return [
    { name: 'Product A', value: 400 },
    { name: 'Product B', value: 300 },
    { name: 'Product C', value: 200 },
    { name: 'Product D', value: 100 }
  ];
};

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

const DashBoard = () => {
  const [period, setPeriod] = useState('daily');
  const salesData = generateMockSalesData(period);
  const productSalesData = generateProductSalesData();

  const totalSales = salesData.reduce((sum, item) => sum + item.sales, 0);
  const averageSales = Math.round(totalSales / salesData.length);

  return (
    <div className="bg-white p-6 rounded-xl shadow-lg mx-auto ">
      <div className="flex justify-between items-center mb-6 ml-12">
        <h1 className="text-2xl font-bold text-gray-800 flex items-center">
          <TrendingUp className="mr-2 text-blue-600" /> Sales Dashboard
        </h1>
        <div className="flex space-x-2">
          {['daily', 'weekly', 'monthly', 'yearly'].map(p => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`px-3 py-1 rounded-full text-sm uppercase tracking-wider transition-all ${
                period === p 
                  ? 'bg-blue-600 text-white' 
                  : 'bg-gray-100 text-gray-600 hover:bg-blue-100'
              }`}
            >
              {p}
            </button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="bg-blue-50 p-4 rounded-lg">
          <div className="flex items-center mb-2">
            <Clock className="mr-2 text-blue-600" />
            <span className="font-semibold text-gray-600">Real-time Sales</span>
          </div>
          <p className="text-2xl font-bold text-blue-800">฿{totalSales.toLocaleString()}</p>
        </div>
        
        <div className="bg-green-50 p-4 rounded-lg">
          <div className="flex items-center mb-2">
            <Calendar className="mr-2 text-green-600" />
            <span className="font-semibold text-gray-600">{period.charAt(0).toUpperCase() + period.slice(1)} Average</span>
          </div>
          <p className="text-2xl font-bold text-green-800">฿{averageSales.toLocaleString()}</p>
        </div>
        
        <div className="bg-purple-50 p-4 rounded-lg">
          <div className="flex items-center mb-2">
            <RefreshCcw className="mr-2 text-purple-600" />
            <span className="font-semibold text-gray-600">Performance Period</span>
          </div>
          <p className="text-2xl font-bold text-purple-800">{period.toUpperCase()}</p>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-6">
        <div className="h-[400px] bg-gray-50 p-4 rounded-lg">
          <h2 className="text-xl font-semibold mb-4 text-center">Sales vs Target</h2>
          <ResponsiveContainer width="100%" height="90%">
            <BarChart data={salesData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="sales" fill="#3b82f6" name="Actual Sales" />
              <Bar dataKey="target" fill="#10b981" name="Sales Target" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="h-[400px] bg-gray-50 p-4 rounded-lg">
          <h2 className="text-xl font-semibold mb-4 text-center">Profit and Expenses</h2>
          <ResponsiveContainer width="100%" height="90%">
            <LineChart data={salesData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="profit" stroke="#8884d8" name="Profit" />
              <Line type="monotone" dataKey="expenses" stroke="#82ca9d" name="Expenses" />
            </LineChart>
          </ResponsiveContainer>
        </div>

        <div className="h-[400px] bg-gray-50 p-4 rounded-lg col-span-2">
          <h2 className="text-xl font-semibold mb-4 text-center">Product Sales Distribution</h2>
          <ResponsiveContainer width="100%" height="90%" >
            <PieChart>
              <Pie
                data={productSalesData}
                cx="50%"
                cy="50%"
                labelLine={false}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
              >
                {productSalesData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
};

export default DashBoard;