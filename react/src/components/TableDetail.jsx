import React, { useState } from "react";
import styles from "../styles/TableDetail.module.css";

const TableDetail = ({ tables, onTableAction, onClearTable, onTableCombine }) => (
  <div className="p-4">
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      {tables.map((table) => (
        <div
          key={table.id}
          className={`
            ${table.status === 'ว่าง' ? 'bg-green-100' :
              table.status === 'จอง' ? 'bg-yellow-100' : 
               table.status === 'ไม่ว่าง' ? 'bg-red-100' :'bg-blue-100'}
            border border-gray-300 rounded-lg p-4 shadow-md
          `}
        >
          <div className="flex justify-between items-center mb-2">
            <h3 className="m-0">โต๊ะ {table.number}</h3>
            <div className="flex gap-2">
              {/* ปุ่มรวมโต๊ะ */}
              <button
                onClick={() => onTableCombine(table)}
                className="px-2 py-2 border border-green-500 rounded bg-white text-green-500 hover:bg-green-100"
              >
                รวมโต๊ะ
              </button>
              {/* สิ้นสุดปุ่มรวมโต๊ะ */}

              {/* ปุ่มจัดการโต๊ะ */}
              <button
                onClick={() => onTableAction(table)}
                className="px-2 py-2 border border-blue-500 rounded bg-white text-blue-500 hover:bg-blue-100"
              >
                จัดการ
              </button>
              {/* สิ้นสุดปุ่มจัดการโต๊ะ */}

              {/* ปุ่มเคลียร์โต๊ะ */}
              {table.status !== 'ว่าง' && (
                <button
                  onClick={() => onClearTable(table.id)}
                  className="px-2 py-2 border border-red-500 rounded bg-white text-red-500 hover:bg-red-100"
                >
                  ดูรายละเอียด
                </button>
              )}
              {/* สิ้นสุดปุ่มเคลียร์โต๊ะ */}
            </div>
          </div>
          <div>
            <div className="flex justify-between mb-2">
              <span>สถานะ:</span>
              <span className={`
                ${table.status === 'ว่าง' ? 'text-green-500' :
                  table.status === 'จอง' ? 'text-yellow-500' : 'text-red-500'}
              `}>
                {table.status}
              </span>
            </div>
            <div className="flex justify-between mb-2">
              <span>ที่นั่ง:</span>
              <span>{table.seats} ที่นั่ง</span>
            </div>
            {table.customerName && (
              <div className="flex justify-between mb-2">
                <span>ลูกค้า:</span>
                <span>{table.customerName}</span>
              </div>
            )}
            {table.customerCount && (
              <div className="flex justify-between mb-2">
                <span>จำนวน:</span>
                <span>{table.customerCount} คน</span>
              </div>
            )}
            {table.time && (
              <div className="flex justify-between mb-2">
                <span>เวลา:</span>
                <span>{table.time}</span>
              </div>
            )}
          </div>
        </div>
      ))}
    </div>
  </div>
);

export default TableDetail;

