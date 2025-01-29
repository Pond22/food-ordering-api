import React from 'react'
import MenuBar from '../components/MenuBar'
import MenuList from '../components/MenuList2'
import { useSearchParams } from 'react-router-dom'

export default function Menu() {
  const [searchParams] = useSearchParams()

  // ดึงค่า table และ uuid จาก URL
  const tableID = parseInt(searchParams.get('tableID'),10)
  const uuid = searchParams.get('uuid')
  console.log(tableID,'+', uuid)

  return (
    <div className="bg-gray-100 min-h-screen">
      <MenuBar tableID={tableID} uuid={uuid}/>
      <MenuList />
    </div>
  )
}
