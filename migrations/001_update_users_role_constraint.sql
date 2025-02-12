-- ลบ constraint เก่าออกก่อน
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;

-- สร้าง type enum สำหรับ role
DROP TYPE IF EXISTS user_role;
CREATE TYPE user_role AS ENUM ('staff', 'manager', 'owner');

-- แก้ไข column role ให้ใช้ type enum
ALTER TABLE users 
    ALTER COLUMN role TYPE user_role 
    USING role::user_role;

-- เพิ่ม constraint ใหม่ (ถ้าต้องการ)
ALTER TABLE users 
    ADD CONSTRAINT chk_users_role 
    CHECK (role IN ('staff', 'manager', 'owner'));