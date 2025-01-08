import React, { Component } from 'react';

// Error Boundary Component
class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, errorInfo: null };
  }

  static getDerivedStateFromError(error) {
    // อัปเดต state เพื่อให้แสดง UI สำรองเมื่อเกิดข้อผิดพลาด
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    // คุณสามารถบันทึกข้อผิดพลาดหรือทำการส่งไปยังเซิร์ฟเวอร์ได้
    this.setState({ errorInfo });
    console.error("Error Boundary caught an error", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      // แสดง UI สำรองเมื่อเกิดข้อผิดพลาด
      return (
        <div>
          <h1>Something went wrong.</h1>
          <details style={{ whiteSpace: 'pre-wrap' }}>
            {this.state.errorInfo && this.state.errorInfo.componentStack}
          </details>
        </div>
      );
    }

    // ปกติแล้วให้เรนเดอร์คอมโพเนนต์ที่ถูกห่อหุ้ม
    return this.props.children;
  }
}

export default ErrorBoundary;
