import React, { useState, useEffect } from 'react';
import { 
  FileSpreadsheet, 
  Upload, 
  CheckCircle, 
  AlertCircle, 
  X,
  Download,
  Info,
  ArrowLeft
} from 'lucide-react';
import * as XLSX from 'xlsx';

// Custom Alert Component
const Alert = ({ variant = 'default', children, className = '' }) => {
  const baseStyles = 'p-4 rounded-lg mb-4 flex items-start gap-3';
  const variantStyles = {
    default: 'bg-blue-50 text-blue-800',
    destructive: 'bg-red-50 text-red-800'
  };
  
  return (
    <div className={`${baseStyles} ${variantStyles[variant]} ${className}`}>
      {children}
    </div>
  );
};

// Custom Card Components
const Card = ({ children, className = '' }) => (
  <div className={`bg-white rounded-lg border shadow-sm ${className}`}>
    {children}
  </div>
);

const CardHeader = ({ children, className = '' }) => (
  <div className={`p-6 ${className}`}>
    {children}
  </div>
);

const CardTitle = ({ children, className = '' }) => (
  <h3 className={`text-lg font-semibold leading-none tracking-tight ${className}`}>
    {children}
  </h3>
);

const CardDescription = ({ children, className = '' }) => (
  <p className={`text-sm text-gray-500 ${className}`}>
    {children}
  </p>
);

const CardContent = ({ children, className = '' }) => (
  <div className={`p-6 pt-0 ${className}`}>
    {children}
  </div>
);

// Fetch categories function
const fetchCategories = async () => {
  try {
    const token = localStorage.getItem('token');
    const response = await fetch('http://localhost:8080/api/categories', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error('Failed to fetch categories');
    }

    const data = await response.json();
    console.log('Fetched categories:', data); // เพิ่ม log เพื่อดูข้อมูลที่ได้
    return data;
  } catch (error) {
    console.error('Error fetching categories:', error);
    throw error;
  }
};

const MenuImportPage = () => {
  const [file, setFile] = useState(null);
  const [previewData, setPreviewData] = useState([]);
  const [error, setError] = useState(null);
  const [importing, setImporting] = useState(false);
  const [importResult, setImportResult] = useState(null);
  const [step, setStep] = useState(1);
  const [categories, setCategories] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [optionDetail, setOptionDetail] = useState({ show: false, content: '' });
  const [loadingCategories, setLoadingCategories] = useState(true);

  // Load categories on component mount
  useEffect(() => {
    const loadCategories = async () => {
      try {
        setLoadingCategories(true);
        setError(null);
        const data = await fetchCategories();
        if (Array.isArray(data) && data.length > 0) {
          setCategories(data);
          console.log('Categories loaded:', data); // เพิ่ม log เพื่อตรวจสอบ
        } else {
          setError('ไม่พบข้อมูลหมวดหมู่ในระบบ');
        }
      } catch (error) {
        console.error('Error loading categories:', error);
        setError('ไม่สามารถโหลดข้อมูลหมวดหมู่ได้');
        setCategories([]); // เคลียร์ categories เมื่อเกิดข้อผิดพลาด
      } finally {
        setLoadingCategories(false);
      }
    };

    loadCategories();
  }, []);

  const handleDragOver = (e) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e) => {
    e.preventDefault();
    e.stopPropagation();
    const droppedFile = e.dataTransfer.files[0];
    if (droppedFile?.type.includes('sheet') || 
        droppedFile?.type.includes('csv') ||
        droppedFile?.name.match(/\.(xlsx|xls|csv)$/)) {
      handleFile(droppedFile);
    } else {
      setError('กรุณาอัปโหลดไฟล์ Excel (.xlsx, .xls) หรือ CSV (.csv) เท่านั้น');
    }
  };

  const handleFile = async (file) => {
    // เพิ่มการตรวจสอบขนาดไฟล์
    const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5MB
    if (file.size > MAX_FILE_SIZE) {
      setError('ขนาดไฟล์ต้องไม่เกิน 5MB');
      return;
    }

    setFile(file);
    setError(null);
    setImportResult(null);

    try {
      const reader = new FileReader();
      reader.onload = (e) => {
        try {
          let data;
          if (file.name.endsWith('.csv')) {
            // อ่านไฟล์ CSV
            const csvText = e.target.result;
            data = csvText.split('\n').map(row => row.split(','));
          } else {
            // อ่านไฟล์ Excel
            const workbook = XLSX.read(e.target.result, { type: 'array' });
            const firstSheet = workbook.Sheets[workbook.SheetNames[0]];
            data = XLSX.utils.sheet_to_json(firstSheet, { header: 1 });
          }

          const preview = data.slice(1)
            .filter(row => row.length > 0)
            .map((row, index) => ({
              id: index,
              name: row[0]?.trim() || '',
              nameEn: row[1]?.trim() || '',
              nameCh: row[2]?.trim() || '',
              description: row[3]?.trim() || '',
              descriptionEn: row[4]?.trim() || '',
              descriptionCh: row[5]?.trim() || '',
              category: row[6]?.trim() || '',
              price: row[7]?.trim() || '',
              options: row[8]?.trim() || '',
              isValid: Boolean(
                row[0] && // ชื่อเมนู
                row[6] && // หมวดหมู่
                validateCategory(row[6]) && // ตรวจสอบว่าหมวดหมู่มีอยู่ในระบบ
                row[7] && // ราคา
                !isNaN(row[7]) && // ราคาต้องเป็นตัวเลข
                validateOptions(row[8]?.trim()) // ตรวจสอบรูปแบบตัวเลือก
              )
            }));

          setPreviewData(preview);
          setStep(2);
        } catch (error) {
          setError('เกิดข้อผิดพลาดในการอ่านไฟล์');
        }
      };
      
      if (file.name.endsWith('.csv')) {
        reader.readAsText(file);
      } else {
        reader.readAsArrayBuffer(file);
      }
    } catch (error) {
      setError('เกิดข้อผิดพลาดในการอ่านไฟล์');
    }
  };

  const handleFileSelect = (e) => {
    const file = e.target.files[0];
    if (file) handleFile(file);
  };

  const handleImport = async () => {
    if (!file || !previewData.length) return;

    setImporting(true);
    setError(null);

    try {
        const formData = new FormData();
        formData.append('file', file);
        const token = localStorage.getItem('token');
    
        const response = await fetch('http://localhost:8080/api/menu/import', {
            method: 'POST',
            body: formData,
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const result = await response.json();
        console.log('Import result:', result); // เพิ่ม log ตรงนี้

        if (!response.ok) {
            throw new Error(result.error || 'นำเข้าข้อมูลไม่สำเร็จ');
        }

        // แก้ไขการตั้งค่า importResult
        setImportResult({
            Success: result.success,
            Failed: result.failed,
            FailedItems: result.failed_items,
            success_items: result.success_items // ตรวจสอบให้แน่ใจว่าใช้ key นี้
        });
        
        setStep(3);
    } catch (error) {
        console.error('Import error:', error);
        setError(error.message);
    } finally {
        setImporting(false);
    }
  };

  const downloadTemplate = () => {
    const template = [
      // Header
      [
        'ชื่อเมนู(ไทย)*', 
        'ชื่อเมนู(EN)', 
        'ชื่อเมนู(CN)', 
        'คำอธิบาย(ไทย)', 
        'คำอธิบาย(EN)', 
        'คำอธิบาย(CN)', 
        'หมวดหมู่*', 
        'ราคา*', 
        'ตัวเลือกเพิ่มเติม'
      ],
      // เมนูพื้นฐาน ไม่มีตัวเลือก
      [
        'ข้าวผัดหมู',
        'Pork Fried Rice',
        '猪肉炒饭',
        'ข้าวผัดหมูสูตรพิเศษ',
        'Special pork fried rice',
        '特制猪肉炒饭',
        'อาหารจานเดียว',
        '60',
        ''
      ],
      // เมนูที่มีตัวเลือกเดียว
      [
        'ผัดกะเพราหมูสับ',
        'Stir-fried Pork with Holy Basil',
        '炒猪肉罗勒',
        'ผัดกะเพราหมูสับ ใส่ไข่ดาว',
        'Stir-fried minced pork with holy basil, served with fried egg',
        '炒猪肉罗勒配煎蛋',
        'อาหารจานเดียว',
        '65',
        'ความเผ็ด|Spiciness|辣度|1|true|ไม่เผ็ด:Not Spicy:不辣:0,เผ็ดน้อย:Less Spicy:小辣:0,เผ็ดมาก:Very Spicy:大辣:0'
      ],
      // เมนูที่มีตัวเลือกเพิ่มเติมและราคา
      [
        'ก๋วยเตี๋ยวต้มยำ',
        'Tom Yum Noodle',
        '冬荫面',
        'ก๋วยเตี๋ยวต้มยำรสเด็ด',
        'Spicy and sour noodle soup',
        '酸辣面',
        'ก๋วยเตี๋ยว',
        '55',
        'เส้น|Noodle Type|面条|1|true|เส้นเล็ก:Rice Noodle:米粉:0,เส้นใหญ่:Wide Noodle:河粉:0,บะหมี่:Egg Noodle:蛋面:0||เนื้อสัตว์|Meat|肉类|1|true|หมู:Pork:猪肉:0,ไก่:Chicken:鸡肉:0,ทะเล:Seafood:海鲜:+20||ไข่|Egg|蛋|1|false|ไข่ไก่:Chicken Egg:鸡蛋:+10,ไข่เป็ด:Duck Egg:鸭蛋:+15'
      ],
      // เมนูที่มีตัวเลือกหลายรายการ
      [
        'สุกี้',
        'Suki',
        '泰式火锅',
        'สุกี้น้ำหรือแห้ง',
        'Thai-style hot pot, dry or soup',
        '泰式火锅',
        'อาหารจานเดียว',
        '70',
        'ประเภท|Type|类型|1|true|น้ำ:Soup:汤:0,แห้ง:Dry:干:0||เนื้อสัตว์|Meat|肉类|2|true|หมู:Pork:猪肉:0,ไก่:Chicken:鸡肉:0,ทะเล:Seafood:海鲜:+20||ไข่|Egg|蛋|1|false|ไข่ไก่:Chicken Egg:鸡蛋:+10,ไข่เป็ด:Duck Egg:鸭蛋:+15'
      ],
      // เมนูที่มีตัวเลือกพิเศษ
      [
        'ชาไทย',
        'Thai Tea',
        '泰式奶茶',
        'ชาไทยรสชาติดั้งเดิม',
        'Traditional Thai tea',
        '传统泰式奶茶',
        'เครื่องดื่ม',
        '45',
        'ความหวาน|Sweetness|糖度|1|true|ปกติ:Normal:正常:0,น้อย:Less:少糖:0,ไม่หวาน:No Sugar:无糖:0||เพิ่มเติม|Extra|额外|3|false|ไข่มุก:Boba:珍珠:+10,วิปครีม:Whipped Cream:奶油:+15,พุดดิ้ง:Pudding:布丁:+10'
      ]
    ];

    // คำอธิบายการใช้งาน
    const instructions = [
      ['คำแนะนำการใช้งาน:'],
      ['1. ช่องที่มีเครื่องหมาย * จำเป็นต้องกรอก'],
      ['2. หมวดหมู่ต้องตรงกับที่มีในระบบ'],
      ['3. รูปแบบตัวเลือกเพิ่มเติม:'],
      ['   - แต่ละกลุ่มตัวเลือกคั่นด้วย ||'],
      ['   - แต่ละส่วนในกลุ่มคั่นด้วย |'],
      ['   - รูปแบบ: ชื่อกลุ่ม|ชื่อกลุ่มEN|ชื่อกลุ่มCN|จำนวนที่เลือกได้|บังคับ?|ตัวเลือก'],
      ['   - รูปแบบตัวเลือก: ชื่อ:ชื่อEN:ชื่อCN:ราคาเพิ่ม'],
      ['4. ตัวอย่างตัวเลือก:'],
      ['   ความเผ็ด|Spiciness|辣度|1|true|ไม่เผ็ด:Not Spicy:不辣:0,เผ็ดน้อย:Less Spicy:小辣:0'],
      ['5. คำอธิบายพารามิเตอร์:'],
      ['   - จำนวนที่เลือกได้: ระบุเป็นตัวเลข เช่น 1 คือเลือกได้ 1 รายการ'],
      ['   - บังคับ?: ใส่ true ถ้าต้องเลือก, false ถ้าไม่บังคับ'],
      ['   - ราคาเพิ่ม: ใส่ 0 ถ้าไม่มีราคาเพิ่ม, ใส่ +10 ถ้าเพิ่ม 10 บาท'],
      ['6. หมายเหตุ:'],
      ['   - สามารถมีได้หลายกลุ่มตัวเลือก คั่นด้วย ||'],
      ['   - ราคาต้องเป็นตัวเลขเท่านั้น'],
      ['   - ชื่อหมวดหมู่ต้องตรงกับที่มีในระบบ']
    ];

    const wb = XLSX.utils.book_new();
    
    // สร้าง Sheet ตัวอย่าง
    const wsData = XLSX.utils.aoa_to_sheet(template);
    XLSX.utils.book_append_sheet(wb, wsData, 'ตัวอย่างข้อมูล');
    
    // สร้าง Sheet คำอธิบาย
    const wsInstructions = XLSX.utils.aoa_to_sheet(instructions);
    XLSX.utils.book_append_sheet(wb, wsInstructions, 'คำอธิบายการใช้งาน');

    // ปรับความกว้างคอลัมน์
    const wsDataCols = [
      {wch: 20}, // ชื่อเมนู
      {wch: 25}, // ชื่อเมนู EN
      {wch: 20}, // ชื่อเมนู CN
      {wch: 35}, // คำอธิบาย
      {wch: 35}, // คำอธิบาย EN
      {wch: 35}, // คำอธิบาย CN
      {wch: 15}, // หมวดหมู่
      {wch: 10}, // ราคา
      {wch: 100} // ตัวเลือกเพิ่มเติม
    ];
    wsData['!cols'] = wsDataCols;

    // ปรับความกว้างคอลัมน์คำอธิบาย
    wsInstructions['!cols'] = [{wch: 100}];

    XLSX.writeFile(wb, 'menu_import_template.xlsx');
  };

  const getValidationMessage = (item) => {
    const errors = [];
    
    if (!item.name) errors.push('ชื่อเมนูจำเป็น');
    if (!item.category) errors.push('หมวดหมู่จำเป็น');
    if (!validateCategory(item.category)) {
      errors.push(`หมวดหมู่ "${item.category}" ไม่มีในระบบ`);
    }
    if (!item.price || isNaN(item.price)) errors.push('ราคาไม่ถูกต้อง');
    if (item.options && item.options.trim() !== '' && !validateOptions(item.options)) {
      errors.push('รูปแบบตัวเลือกไม่ถูกต้อง (ตรวจสอบรูปแบบในเทมเพลต)');
    }
    
    return errors.join(', ');
  };

  // เพิ่มฟังก์ชันสำหรับแสดงตัวอย่างข้อมูลตัวเลือก
  const formatOptionsPreview = (optionsStr) => {
    if (!optionsStr) return '-';
    const groups = optionsStr.split('||');
    return groups.map(group => {
      const [name] = group.split('|');
      return name;
    }).join(', ');
  };

  // แก้ไขฟังก์ชัน validateOptions
  const validateOptions = (optionsStr) => {
    if (!optionsStr) return true; // ถ้าไม่มีตัวเลือก ถือว่าผ่าน
    if (optionsStr.trim() === '') return true; // ถ้าเป็นช่องว่าง ถือว่าผ่าน
    
    try {
      // ทำความสะอาดข้อมูลก่อนตรวจสอบ
      const cleanedStr = optionsStr.trim().replace(/\s+/g, ' ');
      const groups = cleanedStr.split('||').filter(group => group.trim() !== '');

      for (const group of groups) {
        const parts = group.trim().split('|').map(part => part.trim());
        
        // ตรวจสอบจำนวนส่วนประกอบ
        if (parts.length !== 6) {
          console.log('Invalid parts length:', parts);
          return false;
        }

        // ตรวจสอบจำนวนที่เลือกได้
        const maxSelections = parseInt(parts[3]);
        if (isNaN(maxSelections) || maxSelections < 1) {
          console.log('Invalid max selections:', parts[3]);
          return false;
        }

        // ตรวจสอบค่าบังคับเลือก
        if (!['true', 'false'].includes(parts[4].toLowerCase())) {
          console.log('Invalid required flag:', parts[4]);
          return false;
        }

        // ตรวจสอบตัวเลือก
        const options = parts[5].split(',').filter(opt => opt.trim() !== '');
        for (const opt of options) {
          const optParts = opt.trim().split(':').map(part => part.trim());
          if (optParts.length !== 4) {
            console.log('Invalid option parts:', optParts);
            return false;
          }
          // ตรวจสอบราคา (ยอมให้มี + ได้)
          const price = optParts[3].replace('+', '');
          if (isNaN(parseFloat(price))) {
            console.log('Invalid price:', optParts[3]);
            return false;
          }
        }
      }
      return true;
    } catch (error) {
      console.error('Error validating options:', error);
      return false;
    }
  };

  // แก้ไขฟังก์ชัน formatOptionsDetail
  const formatOptionsDetail = (optionsStr) => {
    if (!optionsStr) return 'ไม่มีตัวเลือกเพิ่มเติม';
    
    try {
        const groups = optionsStr.split('||').filter(group => group.trim() !== '');
        return groups.map(group => {
            const parts = group.split('|');
            if (parts.length < 6) return null;

            const [name, nameEn, nameCh, maxSel, required, options] = parts;
            const optList = options.split(',')
                .filter(opt => opt.trim() !== '')
                .map(opt => {
                    const [optName, optNameEn, optNameCh, price] = opt.split(':');
                    return `${optName}${price !== '0' ? ` (${price} บาท)` : ''}`;
                })
                .join(', ');
            
            return `
                กลุ่ม: ${name}
                ชื่อ EN: ${nameEn}
                ชื่อ CN: ${nameCh}
                เลือกได้: ${maxSel} รายการ
                บังคับเลือก: ${required === 'true' ? 'ใช่' : 'ไม่'}
                ตัวเลือก: ${optList}
            `;
        }).filter(Boolean).join('\n\n');
    } catch (error) {
        console.error('Error formatting options:', error);
        return 'เกิดข้อผิดพลาดในการแสดงรายละเอียดตัวเลือก';
    }
  };

  // เพิ่มฟังก์ชัน showOptionDetail
  const showOptionDetail = (optionsStr) => {
    const detail = formatOptionsDetail(optionsStr);
    setOptionDetail({ 
        show: true, 
        content: detail 
    });
  };

  // เพิ่ม Modal component สำหรับแสดงรายละเอียด
  const OptionDetailModal = ({ isOpen, onClose, content }) => {
    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 max-h-[80vh] overflow-y-auto">
                <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-medium">รายละเอียดตัวเลือกเพิ่มเติม</h3>
                    <button
                        onClick={onClose}
                        className="text-gray-500 hover:text-gray-700"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>
                <div className="whitespace-pre-line">
                    {content}
                </div>
            </div>
        </div>
    );
  };

  // แก้ไขฟังก์ชัน validateCategory
  const validateCategory = (categoryName) => {
    if (!categoryName) return false;
    
    // ตรวจสอบว่ามีหมวดหมู่นี้ในระบบหรือไม่
    return categories.some(cat => 
      cat.Name.trim().toLowerCase() === categoryName.trim().toLowerCase() ||
      cat.NameEn.trim().toLowerCase() === categoryName.trim().toLowerCase()
    );
  };

  // เพิ่มฟังก์ชันสำหรับจัดรูปแบบการแสดงผลตัวเลือก
  const formatOptionGroupsDisplay = (optionString) => {
    if (!optionString) return '';
    
    try {
        const groups = optionString.split(';').map(group => {
            const [name, nameEn, nameCh, maxSelect, required, options] = group.split('|');
            const optionsList = options.split(',').map(opt => {
                const [optName, optNameEn, optNameCh, price] = opt.split(':');
                return `${optName} (${price} บาท)`;
            }).join(', ');
            
            return `
                <div class="mb-2">
                    <div class="font-medium">${name} / ${nameEn}</div>
                    <div class="text-sm text-gray-600">
                        • เลือกได้ ${maxSelect} รายการ
                        • ${required === 'true' ? 'ต้องเลือก' : 'เลือกหรือไม่ก็ได้'}
                    </div>
                    <div class="text-sm">ตัวเลือก: ${optionsList}</div>
                </div>
            `;
        });
        
        return groups.join('');
    } catch (e) {
        return optionString;
    }
  };

  return (
    <div className="max-w-7xl mx-auto p-6">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">นำเข้าเมนู</h1>
        <p className="text-gray-500">อัปโหลดไฟล์ Excel เพื่อนำเข้าเมนูและตัวเลือกเพิ่มเติม</p>
      </div>

      {/* Steps */}
      <div className="flex items-center justify-center mb-8">
        <div className="flex items-center">
          <div className={`flex items-center justify-center w-8 h-8 rounded-full ${
            step >= 1 ? 'bg-blue-600 text-white' : 'bg-gray-200'
          }`}>
            1
          </div>
          <div className="mx-2 text-sm">อัปโหลด</div>
          <div className="w-16 h-1 mx-2 bg-gray-200">
            <div className={`h-full ${step >= 2 ? 'bg-blue-600' : ''}`}></div>
          </div>
        </div>
        <div className="flex items-center">
          <div className={`flex items-center justify-center w-8 h-8 rounded-full ${
            step >= 2 ? 'bg-blue-600 text-white' : 'bg-gray-200'
          }`}>
            2
          </div>
          <div className="mx-2 text-sm">ตรวจสอบ</div>
          <div className="w-16 h-1 mx-2 bg-gray-200">
            <div className={`h-full ${step >= 3 ? 'bg-blue-600' : ''}`}></div>
          </div>
        </div>
        <div className="flex items-center">
          <div className={`flex items-center justify-center w-8 h-8 rounded-full ${
            step >= 3 ? 'bg-blue-600 text-white' : 'bg-gray-200'
          }`}>
            3
          </div>
          <div className="mx-2 text-sm">เสร็จสิ้น</div>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <div>
            <div className="font-medium">เกิดข้อผิดพลาด</div>
            <div className="text-sm">{error}</div>
          </div>
        </Alert>
      )}

      {step === 1 && (
        <Card>
          <CardHeader>
            <CardTitle>อัปโหลดไฟล์</CardTitle>
            <CardDescription>
              ลากไฟล์มาวางหรือคลิกเพื่อเลือกไฟล์ Excel (.xlsx, .xls)
            </CardDescription>
          </CardHeader>
          <CardContent>
            {loadingCategories ? (
              <div className="text-center py-4">
                <div className="w-6 h-6 border-2 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
                <p className="text-gray-500">กำลังโหลดข้อมูลหมวดหมู่...</p>
              </div>
            ) : categories.length === 0 ? (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <div>
                  <div className="font-medium">ไม่พบข้อมูลหมวดหมู่</div>
                  <div className="text-sm">กรุณาเพิ่มหมวดหมู่ก่อนนำเข้าเมนู</div>
                </div>
              </Alert>
            ) : (
              <div 
                className="border-2 border-dashed rounded-lg p-8 text-center hover:border-blue-500 transition-colors"
                onDragOver={handleDragOver}
                onDrop={handleDrop}
              >
                <div className="mb-4">
                  <FileSpreadsheet className="w-12 h-12 mx-auto text-gray-400" />
                </div>
                <label className="cursor-pointer">
                  <span className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 inline-block mb-2">
                    เลือกไฟล์
                  </span>
                  <input
                    type="file"
                    accept=".xlsx,.xls"
                    className="hidden"
                    onChange={handleFileSelect}
                  />
                </label>
                <p className="text-sm text-gray-500">หรือลากไฟล์มาวางที่นี่</p>
              </div>
            )}
            <div className="mt-4 flex justify-between items-center">
              <button
                onClick={downloadTemplate}
                className="flex items-center gap-2 text-blue-600 hover:text-blue-800"
              >
                <Download className="w-4 h-4" />
                ดาวน์โหลดเทมเพลต
              </button>
              <div className="flex items-center gap-1 text-sm text-gray-500">
                <Info className="w-4 h-4" />
                รองรับไฟล์ .xlsx, .xls และ .csv เท่านั้น
              </div>
            </div>
            {!loadingCategories && categories.length > 0 && (
              <div className="mt-4">
                <h4 className="font-medium mb-2">หมวดหมู่ที่มีในระบบ:</h4>
                <div className="flex flex-wrap gap-2">
                  {categories
                    .filter(cat => cat.Name.trim() !== '') // กรองหมวดหมู่ที่มีชื่อเป็นค่าว่าง
                    .map(cat => (
                      <span 
                        key={cat.ID} 
                        className="px-2 py-1 bg-gray-100 rounded-lg text-sm flex items-center gap-1"
                      >
                        <span className="font-medium">{cat.Name}</span>
                        {cat.NameEn && (
                          <span className="text-gray-500">({cat.NameEn})</span>
                        )}
                      </span>
                    ))}
                </div>
                <div className="mt-2 text-sm text-red-500">
                  * กรุณาใช้ชื่อหมวดหมู่ภาษาไทยหรือภาษาอังกฤษตามที่แสดงด้านบนเท่านั้น
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {step === 2 && previewData.length > 0 && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle>ตรวจสอบข้อมูล</CardTitle>
              <CardDescription>
                ตรวจสอบข้อมูลก่อนนำเข้า ({previewData.length} รายการ)
              </CardDescription>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setStep(1)}
                className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:text-gray-800"
              >
                <ArrowLeft className="w-4 h-4" />
                ย้อนกลับ
              </button>
              <button
                onClick={handleImport}
                disabled={importing || !previewData.some(item => item.isValid) || isLoading}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400"
              >
                {isLoading ? (
                  <>
                    <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                    กำลังประมวลผล...
                  </>
                ) : (
                  <>
                    <Upload className="w-4 h-4" />
                    {importing ? 'กำลังนำเข้า...' : 'นำเข้าข้อมูล'}
                  </>
                )}
              </button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">สถานะ</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">ชื่อเมนู</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">หมวดหมู่</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">ราคา</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">ตัวเลือกเพิ่มเติม</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {previewData.map((item) => (
                    <tr key={item.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3">
                        {item.isValid ? (
                          <div className="flex items-center gap-2">
                            <CheckCircle className="w-5 h-5 text-green-500" />
                            <span className="text-sm text-green-600">ผ่าน</span>
                          </div>
                        ) : (
                          <div className="flex items-center gap-2">
                            <AlertCircle className="w-5 h-5 text-red-500" />
                            <span className="text-sm text-red-600">{getValidationMessage(item)}</span>
                          </div>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <div>
                          <div className="font-medium text-gray-900">{item.name}</div>
                          <div className="text-sm text-gray-500">{item.nameEn}</div>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-gray-900">{item.category}</td>
                      <td className="px-4 py-3 text-gray-900">{item.price}</td>
                      <td className="px-4 py-3">
                        <div className="max-w-xs">
                          <div className="text-gray-900">{formatOptionsPreview(item.options)}</div>
                          {item.options && (
                            <button
                              onClick={() => showOptionDetail(item.options)}
                              className="text-sm text-blue-600 hover:text-blue-800"
                            >
                              ดูรายละเอียด
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {step === 3 && importResult && (
        <Card>
          <CardHeader>
            <CardTitle>นำเข้าข้อมูลเสร็จสิ้น</CardTitle>
            <CardDescription>สรุปผลการนำเข้าข้อมูล</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-6">
                {/* แสดงสรุปจำนวน */}
                <div className="flex items-center gap-4">
                    {importResult.Success > 0 && (
                        <div className="bg-green-100 text-green-800 rounded-lg p-4 flex-1">
                            <div className="font-medium">นำเข้าสำเร็จ</div>
                            <div className="text-2xl font-bold">
                                {importResult.Success} รายการ
                            </div>
                        </div>
                    )}
                    {importResult.Failed > 0 && (
                        <div className="bg-red-100 text-red-800 rounded-lg p-4 flex-1">
                            <div className="font-medium">ไม่สำเร็จ</div>
                            <div className="text-2xl font-bold">
                                {importResult.Failed} รายการ
                            </div>
                        </div>
                    )}
                </div>

                {/* แสดงรายการที่สำเร็จ */}
                {importResult.Success > 0 && importResult.success_items && (
                    <div className="mt-6">
                        <h4 className="font-medium text-lg mb-3">รายการที่นำเข้าสำเร็จ ({importResult.Success} รายการ)</h4>
                        <div className="grid gap-4">
                            {importResult.success_items.map((item, index) => (
                                <div key={index} className="bg-green-50 p-4 rounded-lg">
                                    <div className="flex items-center gap-2 mb-2">
                                        <CheckCircle className="w-5 h-5 text-green-500" />
                                        <div className="font-medium">{item.name}</div>
                                    </div>
                                    <div className="grid grid-cols-3 gap-4 text-sm">
                                        <div>
                                            <div className="text-gray-600">ชื่อภาษาอังกฤษ:</div>
                                            <div>{item.name_en || '-'}</div>
                                        </div>
                                        <div>
                                            <div className="text-gray-600">หมวดหมู่:</div>
                                            <div>{item.category_name}</div>
                                        </div>
                                        <div>
                                            <div className="text-gray-600">ราคา:</div>
                                            <div>{item.price} บาท</div>
                                        </div>
                                    </div>
                                    {item.option_groups && item.option_groups.length > 0 && (
                                        <div className="mt-3">
                                            <div className="text-gray-600 mb-1">ตัวเลือกเพิ่มเติม:</div>
                                            <div className="text-sm">
                                                {item.option_groups.map((group, gIndex) => (
                                                    <div key={gIndex} className="mb-2">
                                                        <div className="font-medium">{group.name}</div>
                                                        <div className="text-gray-600">
                                                            {group.options.map(opt => opt.name).join(', ')}
                                                        </div>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {/* แสดงรายการที่ไม่สำเร็จ */}
                {importResult.FailedItems && importResult.FailedItems.length > 0 && (
                    <div className="mt-6">
                        <h4 className="font-medium text-lg mb-3">รายการที่ไม่สำเร็จ</h4>
                        <div className="space-y-4">
                            {importResult.FailedItems.map((item, index) => (
                                <div key={index} className="bg-red-50 p-4 rounded-lg">
                                    <div className="flex items-start gap-2">
                                        <AlertCircle className="w-5 h-5 text-red-500 mt-1" />
                                        <div className="flex-1">
                                            <div className="font-medium text-red-800">
                                                แถวที่ {item.row}
                                            </div>
                                            <div className="text-red-600 mt-1">
                                                {item.error}
                                            </div>
                                            {item.input_data && (
                                                <div className="mt-2 bg-white p-3 rounded border border-red-100">
                                                    <div className="font-medium text-gray-700 mb-1">ข้อมูลที่ระบุ:</div>
                                                    <div className="text-sm text-gray-600"
                                                         dangerouslySetInnerHTML={{
                                                             __html: formatOptionGroupsDisplay(item.input_data)
                                                         }}
                                                    />
                                                </div>
                                            )}
                                            {item.suggestion && (
                                                <div className="mt-2 text-blue-600">
                                                    <span className="font-medium">คำแนะนำ:</span> {item.suggestion}
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                <div className="flex justify-end gap-2 mt-6">
                    <button
                        onClick={() => {
                            setStep(1);
                            setFile(null);
                            setPreviewData([]);
                            setImportResult(null);
                            setError(null);
                        }}
                        className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200"
                    >
                        นำเข้าไฟล์ใหม่
                    </button>
                    <button
                        onClick={() => window.location.href = '/menu'}
                        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                    >
                        ไปที่หน้าจัดการเมนู
                    </button>
                </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* เพิ่ม Modal component */}
      <OptionDetailModal
        isOpen={optionDetail.show}
        onClose={() => setOptionDetail({ show: false, content: '' })}
        content={optionDetail.content}
      />
    </div>
  );
};

export default MenuImportPage;