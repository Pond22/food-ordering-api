import React, { useState, useEffect } from 'react';
import { Printer, Search, ChevronDown, ChevronUp, Info } from 'lucide-react';

// Subcomponents
const Table = ({ children, className = "" }) => (
  <table className={`w-full text-sm text-left ${className}`}>{children}</table>
);

const TableBody = ({ children }) => <tbody>{children}</tbody>;

const TableCell = ({ children, className = "", colSpan }) => (
  <td className={`p-4 ${className}`} colSpan={colSpan}>{children}</td>
);

const TableHead = ({ children, className = "" }) => (
  <th className={`p-4 font-medium ${className}`}>{children}</th>
);

const TableHeader = ({ children }) => <thead className="bg-gray-50">{children}</thead>;

const TableRow = ({ children, className = "" }) => (
  <tr className={`border-b ${className}`}>{children}</tr>
);

const Alert = ({ children, variant = "default" }) => (
  <div className={`p-4 rounded-lg ${variant === "destructive" ? "bg-red-50 text-red-600" : "bg-gray-50"}`}>
    <div className="flex gap-2 items-start">{children}</div>
  </div>
);

const preparePrintContent = (job) => {
  if (!job || !job.content) return "ไม่พบข้อมูลสำหรับการพิมพ์";

  try {
    if (job.job_type === "qr_code") {
      return (
        <img
          src={`data:image/png;base64,${job.content}`}
          alt="QR Code"
          className="mx-auto border rounded shadow-md"
          style={{ maxWidth: "200px", height: "auto" }}
        />
      );
    }

    const binaryString = atob(job.content);
    const bytes = new Uint8Array(binaryString.length);
    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }

    const decoder = new TextDecoder("utf-8");
    const decodedContent = decoder.decode(bytes);

    return (
      <pre className="font-mono bg-white p-4 border rounded shadow-sm max-w-md mx-auto">
        {decodedContent}
      </pre>  
    );
  } catch (err) {
    console.error("Error decoding content:", err);
    return "ไม่สามารถแสดงเนื้อหาได้";
  }
};

const renderJobDetails = (job) => (
  <div className="px-4 py-2 bg-gray-50">
    {preparePrintContent(job)} 
  </div>
);

const Reprint = () => {
  const today = new Date();
  const startOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
  const endOfMonth = new Date(today.getFullYear(), today.getMonth() + 1, 0);

  const [jobs, setJobs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [expandedRows, setExpandedRows] = useState(new Set());
  const [filters, setFilters] = useState({
    type: '',
    startDate: startOfMonth.toISOString().split('T')[0], 
    endDate: endOfMonth.toISOString().split('T')[0],
    searchTerm: ''
  });

  const fetchJobs = async () => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams();
      if (filters.type) params.append('type', filters.type);
      if (filters.startDate) params.append('start_date', filters.startDate);
      if (filters.endDate) params.append('end_date', filters.endDate);
      
      const response = await fetch(
        `http://localhost:8080/api/printers/reprintable-jobs?${params.toString()}`
      );
      
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }
      
      const data = await response.json();
      setJobs(Array.isArray(data) ? data : []);  
    } catch (err) {
      console.error('Fetch error:', err);
      setError(err.message);
      setJobs([]);
    } finally {
      setLoading(false);
    }
  };

  const handleReprint = async (jobId) => {
    try {
      setLoading(true);
      const response = await fetch(`http://localhost:8080/api/printers/reprint/${jobId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        throw new Error(`Reprint failed with status ${response.status}`);
      }
      
      alert('รีปริ้นสำเร็จ');
      await fetchJobs();
    } catch (err) {
      console.error('Reprint error:', err);
      alert('เกิดข้อผิดพลาดในการรีปริ้น: ' + err.message);  
    } finally {
      setLoading(false);
    }
  };

  const toggleRowExpand = (id) => {
    setExpandedRows(prev => {
      const newSet = new Set(prev);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      return newSet;
    });
  };

  const filteredJobs = React.useMemo(() => {
    if (!filters.searchTerm) return jobs;
    
    const searchLower = filters.searchTerm.toLowerCase();
    return jobs.filter(job => {
      const orderId = job?.order?.id?.toString() || '';
      const receiptId = job?.receipt?.id?.toString() || '';
      const printerName = job?.printer?.name || '';  
      const tableId = job?.order?.table_id?.toString() || job?.receipt?.table_id || '';
      
      return (
        orderId.includes(searchLower) ||
        receiptId.includes(searchLower) ||
        printerName.toLowerCase().includes(searchLower) ||
        tableId.includes(searchLower)  
      );
    });
  }, [jobs, filters.searchTerm]);

  useEffect(() => {
    fetchJobs();
  }, [filters.type, filters.startDate, filters.endDate]);
  
  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">ระบบจัดการรีปริ้น</h1>
      
      <div className="space-y-6">
        <div className="flex flex-wrap gap-4">
          <select
            value={filters.type}
            onChange={(e) => setFilters(prev => ({ ...prev, type: e.target.value }))}
            className="border rounded-md px-3 py-2 min-w-[200px]"
          >
            <option value="">ทั้งหมด</option>
            <option value="order">ออเดอร์</option>
            <option value="receipt">ใบเสร็จ</option>
            <option value="cancelation">ยกเลิกรายการ</option>
            <option value="qr_code">QR Code</option>
          </select>
          
          <input
            type="date"
            value={filters.startDate}
            onChange={(e) => setFilters(prev => ({ ...prev, startDate: e.target.value }))}
            className="border rounded-md px-3 py-2"  
          />
          
          <input
            type="date"
            value={filters.endDate}
            onChange={(e) => setFilters(prev => ({ ...prev, endDate: e.target.value }))}
            className="border rounded-md px-3 py-2"
          />  
          
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-gray-400" />
            <input
              type="text"
              placeholder="ค้นหา..."
              value={filters.searchTerm}
              onChange={(e) => setFilters(prev => ({ ...prev, searchTerm: e.target.value }))}
              className="w-full border rounded-md pl-8 pr-3 py-2"
            />
          </div>
        </div>

        {error && (
          <Alert variant="destructive">
            <div>{error}</div>  
          </Alert>  
        )}
        
        <div className="bg-white rounded-lg overflow-hidden border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-8"></TableHead>
                <TableHead>วันที่-เวลา</TableHead>
                <TableHead>ประเภท</TableHead>
                <TableHead>โต๊ะ</TableHead>
                <TableHead>เครื่องพิมพ์</TableHead>
                <TableHead className="text-right">การดำเนินการ</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                    กำลังโหลด...
                  </TableCell>
                </TableRow>
              ) : filteredJobs.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                    {error ? 'เกิดข้อผิดพลาดในการโหลดข้อมูล' : 'ไม่พบรายการ'}
                  </TableCell>
                </TableRow>
              ) : (
                filteredJobs.map((job) => (
                  <React.Fragment key={job.id}>
                    <TableRow className="cursor-pointer hover:bg-gray-50">
                      <TableCell>
                        <button
                          onClick={() => toggleRowExpand(job.id)}
                          className="h-4 w-4"
                        >
                          {expandedRows.has(job.id) ? (
                            <ChevronUp className="h-4 w-4" />
                          ) : (
                            <ChevronDown className="h-4 w-4" />
                          )}
                        </button>
                      </TableCell>
                      <TableCell>
                        {new Date(job.created_at).toLocaleString('th-TH')}
                      </TableCell>
                      <TableCell>
                        {job.job_type === 'order' ? 'ออเดอร์' : 
                         job.job_type === 'receipt' ? 'ใบเสร็จ' :
                         job.job_type === 'cancelation' ? 'ยกเลิกรายการ' :
                         job.job_type === 'qr_code' ? 'QR Code' : '-'}
                      </TableCell>
                      <TableCell>
                        {job.order?.table_id || job.receipt?.TableID || '-'}
                      </TableCell>
                      <TableCell>
                        {job.printer?.name || `เครื่องพิมพ์ ${job.printer_id}`}  
                      </TableCell>
                      <TableCell className="text-right">
                        <button 
                          onClick={() => handleReprint(job.id)}
                          disabled={loading}
                          className="inline-flex items-center gap-2 px-4 py-2 bg-gray-100 hover:bg-gray-200 rounded disabled:opacity-50"
                        >
                          <Printer className="h-4 w-4" />
                          รีปริ้น
                        </button>
                      </TableCell>
                    </TableRow>
                    <TableRow>
                      <TableCell colSpan={6} className="p-0 border-b">
                        {expandedRows.has(job.id) && renderJobDetails(job)}
                      </TableCell>  
                    </TableRow>
                  </React.Fragment>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
};

export default Reprint;