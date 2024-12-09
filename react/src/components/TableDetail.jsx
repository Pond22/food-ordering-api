import React, { useState } from "react";
import styles from "../styles/TableDetail.module.css";

const TableDetail = ({ tables, onTableAction, onClearTable }) => (
  <div style={{ padding: '16px' }}>
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '16px' }}>
      {tables.map((table) => (
        <div 
          key={table.id} 
          style={{
            backgroundColor: table.status === 'ว่าง' ? '#e6ffe6' : 
                             table.status === 'จอง' ? '#fff8cc' : '#ffe6e6',
            border: '1px solid #ccc',
            borderRadius: '8px',
            padding: '16px',
            boxShadow: '0 2px 4px rgba(0, 0, 0, 0.1)',
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '8px' }}>
            <h3 style={{ margin: 0 }}>โต๊ะ {table.number}</h3>
            <div style={{ display: 'flex', gap: '8px' }}>
              <button 
                onClick={() => onTableAction(table)}
                style={{
                  padding: '8px 16px',
                  border: '1px solid #007bff',
                  borderRadius: '4px',
                  backgroundColor: '#fff',
                  color: '#007bff',
                  cursor: 'pointer'
                }}
              >
                จัดการ
              </button>
              {table.status !== 'ว่าง' && (
                <button className="px-2 py-4 rounded-md bg-white "
                  onClick={() => onClearTable(table.id)}
                  style={{
                    border: '1px solid #dc3545',
                    color: '#dc3545',
                    cursor: 'pointer'
                  }}
                >
                  ดูรายละเอียด
                </button>
              )}
            </div>
          </div>
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
              <span>สถานะ:</span>
              <span style={{ 
                color: table.status === 'ว่าง' ? '#28a745' : 
                       table.status === 'จอง' ? '#ffc107' : '#dc3545' 
              }}>
                {table.status}
              </span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
              <span>ที่นั่ง:</span>
              <span>{table.seats} ที่นั่ง</span>
            </div>
            {table.customerName && (
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <span>ลูกค้า:</span>
                <span>{table.customerName}</span>
              </div>
            )}
            {table.customerCount && (
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <span>จำนวน:</span>
                <span>{table.customerCount} คน</span>
              </div>
            )}
            {table.time && (
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
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
