import React, { useEffect } from 'react'
import MenuBar from '../components/MenuBar'
import MenuList from '../components/MenuList2'
import useCartStore from '../hooks/cart-store'
import { useSearchParams } from 'react-router-dom'

export default function Menu() {
  const { checkCartExpiry, setTableData } = useCartStore()
  const [searchParams] = useSearchParams()
  
  const tableId = parseInt(searchParams.get('tableID'), 10)
  const uuid = searchParams.get('uuid')

  console.log(tableId, '+', uuid)
  useEffect(() => {
    checkCartExpiry()
  }, [])

  useEffect(() => {
    if (tableId && uuid) {
      setTableData(tableId, uuid)
    }
  }, [tableId, uuid, setTableData])

  return (
    <div className="bg-gray-100 min-h-screen">
      <MenuBar tableID={tableId} uuid={uuid} />
      <MenuList />
    </div>
  )
}
