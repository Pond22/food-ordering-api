import { useEffect, useState } from 'react'

const ActivePromotionsTable = () => {
  const [promotions, setPromotions] = useState([])
  const [error, setError] = useState(null)

  useEffect(() => {
    const fetchPromotions = async () => {
      try {
        const response = await fetch(
          `${import.meta.env.VITE_APP_API_URL}/api/promotions/Active`,
          {
            method: 'GET',
            headers: {
              Accept: 'application/json',
            },
          }
        )

        if (!response.ok) {
          throw new Error('Failed to fetch promotions')
        }

        const data = await response.json()
        setPromotions(data)
      } catch (error) {
        setError(error.message)
      }
    }

    fetchPromotions()
  }, [])

  return (
    <div className="p-6">
      {error && <p className="text-red-500">Error: {error}</p>}
      <h2 className="text-2xl font-semibold mb-4">Active Promotions</h2>

      {promotions.length > 0 ? (
        <table className="min-w-full table-auto border-collapse border border-gray-300">
          <thead className="bg-gray-100">
            <tr>
              <th className="px-4 py-2 border border-gray-300">
                Promotion Name
              </th>
              <th className="px-4 py-2 border border-gray-300">Description</th>
              <th className="px-4 py-2 border border-gray-300">Start Date</th>
              <th className="px-4 py-2 border border-gray-300">End Date</th>
              <th className="px-4 py-2 border border-gray-300">Price (THB)</th>
              <th className="px-4 py-2 border border-gray-300">Items</th>
            </tr>
          </thead>
          <tbody>
            {promotions.map((promotion) => (
              <tr key={promotion.ID} className="border-b">
                <td className="px-4 py-2 border border-gray-300">
                  {promotion.Name}
                </td>
                <td className="px-4 py-2 border border-gray-300">
                  {promotion.Description || 'No description available'}
                </td>
                <td className="px-4 py-2 border border-gray-300">
                  {new Date(promotion.StartDate).toLocaleString()}
                </td>
                <td className="px-4 py-2 border border-gray-300">
                  {new Date(promotion.EndDate).toLocaleString()}
                </td>
                <td className="px-4 py-2 border border-gray-300">
                  {promotion.Price} THB
                </td>
                <td className="px-4 py-2 border border-gray-300">
                  <ul>
                    {promotion.Items.map((item, index) => (
                      <li key={index}>
                        {item.MenuItem.Name} (Quantity: {item.Quantity})
                      </li>
                    ))}
                  </ul>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p>No active promotions available</p>
      )}
    </div>
  )
}

export default ActivePromotionsTable
