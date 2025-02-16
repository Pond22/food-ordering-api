import React, { useState, useEffect } from 'react';
import { Printer, Search, ChevronDown, ChevronUp, Info } from 'lucide-react';
import InfiniteScroll from 'react-infinite-scroll-component';

const API_BASE_URL = 'http://127.0.0.1:8080/api/printers'

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
  const [jobType, setJobType] = useState('reprintable');
  
  // เพิ่ม state สำหรับ pagination
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [totalItems, setTotalItems] = useState(0);

  const fetchJobs = async (pageNum = page, isLoadMore = false) => {
    setLoading(true);
    setError(null);
    try {
      const endpoint = jobType === 'failed' 
        ? `${API_BASE_URL}/failed-jobs`
        : `${API_BASE_URL}/reprintable-jobs`;

      const params = new URLSearchParams({
        page: pageNum,
        limit: 10,
        ...(filters.type && { type: filters.type }),
        ...(filters.startDate && { start_date: filters.startDate }),
        ...(filters.endDate && { end_date: filters.endDate })
      });

      const response = await fetch(`${endpoint}?${params}`);
      
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }
      
      const result = await response.json();
      
      if (isLoadMore) {
        if (result.data && Array.isArray(result.data)) {
          setJobs(prev => [...prev, ...result.data]);
        }
      } else {
        if (result.data && Array.isArray(result.data)) {
          setJobs(result.data);
        } else {
          setJobs([]);
        }
      }
      
      setTotalItems(result.pagination?.total_items || 0);
      setHasMore(result.pagination?.total_pages ? pageNum < result.pagination.total_pages : false);
    } catch (err) {
      console.error('Fetch error:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadMore = () => {
    if (!loading && hasMore) {
      setPage(prev => prev + 1);
      fetchJobs(page + 1, true);
    }
  };

  // Reset pagination when filters change
  useEffect(() => {
    setPage(1);
    setHasMore(true);
    fetchJobs(1, false);
  }, [jobType, filters.type, filters.startDate, filters.endDate]);

  const handleReprint = async (jobId) => {
    try {
      setLoading(true);
      const response = await fetch(`${API_BASE_URL}/reprint/${jobId}`, {
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

  // แยก JobRow เป็น Memoized Component
  const JobRow = React.memo(({ job }) => {
    const isExpanded = expandedRows.has(job.id);
    
    return (
      <React.Fragment>
        <TableRow className="cursor-pointer hover:bg-gray-50">
          <TableCell>
            <button
              onClick={() => toggleRowExpand(job.id)}
              className="h-4 w-4"
            >
              {isExpanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
            </button>
          </TableCell>
          <TableCell>{new Date(job.created_at).toLocaleString('th-TH')}</TableCell>
          <TableCell>
            {job.job_type === 'order' ? 'ออเดอร์' : 
             job.job_type === 'receipt' ? 'ใบเสร็จ' :
             job.job_type === 'cancelation' ? 'ยกเลิกรายการ' :
             job.job_type === 'qr_code' ? 'QR Code' : '-'}
          </TableCell>
          <TableCell>{job.order?.table_id || job.receipt?.TableID || '-'}</TableCell>
          <TableCell>{job.printer?.name || `เครื่องพิมพ์ ${job.printer_id}`}</TableCell>
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
        {isExpanded && (
          <TableRow>
            <TableCell colSpan={6} className="p-0 border-b">
              {renderJobDetails(job)}
            </TableCell>
          </TableRow>
        )}
      </React.Fragment>
    );
  });

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">ระบบจัดการรีปริ้น</h1>
      
      <div className="space-y-6">
        <div className="flex flex-wrap gap-4">
          <div className="flex gap-2">
            <button
              onClick={() => {
                setJobType('reprintable')
                setFilters(prev => ({...prev, type: '', searchTerm: ''}))
              }}
              className={`px-4 py-2 rounded-lg transition-colors ${
                jobType === 'reprintable' 
                  ? 'bg-blue-500 text-white' 
                  : 'bg-gray-100 text-gray-700'
              }`}
            >
              งานปกติ
            </button>
            <button
              onClick={() => {
                setJobType('failed')
                setFilters(prev => ({...prev, type: '', searchTerm: ''}))
              }}
              className={`px-4 py-2 rounded-lg transition-colors ${
                jobType === 'failed'
                  ? 'bg-red-500 text-white'
                  : 'bg-gray-100 text-gray-700'
              }`}
            >
              งานที่ล้มเหลว
            </button>
          </div>

          {jobType === 'reprintable' && (
            <>
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
            </>
          )}

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
          <InfiniteScroll
            dataLength={jobs.length}
            next={loadMore}
            hasMore={hasMore}
            loader={<div className="text-center py-4">กำลังโหลด...</div>}
            endMessage={<div className="text-center py-4">ไม่มีรายการเพิ่มเติม</div>}
            scrollableTarget="scrollableDiv"
            className="overflow-auto"
            height={600}
          >
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
                {loading && jobs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                      กำลังโหลด...
                    </TableCell>
                  </TableRow>
                ) : jobs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                      {error ? 'เกิดข้อผิดพลาดในการโหลดข้อมูล' : 'ไม่พบรายการ'}
                    </TableCell>
                  </TableRow>
                ) : (
                  jobs.map(job => <JobRow key={job.id} job={job} />)
                )}
              </TableBody>
            </Table>
          </InfiniteScroll>
        </div>
      </div>
    </div>
  );
};

export default React.memo(Reprint);