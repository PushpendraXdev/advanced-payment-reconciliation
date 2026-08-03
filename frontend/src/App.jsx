import './App.css'
import { useEffect, useState } from "react";

function App() {
  const [transaction, setTransaction] = useState([])
  const [pendingApproval,takePendingApproval]=useState([])


   async function fetchPendingApproval() {
      const respone=await fetch("http://localhost:8080/matches/pending-approval")
      const data =await respone.json()
      takePendingApproval(data || [])
    }
  async function HandleApprove(matchId){
    await fetch(`http://localhost:8080/matches/${matchId}/approve`,{
      method:"POST"
    })
    fetchPendingApproval()
   }
   async function HandleReject(matchId){
    await fetch(`http://localhost:8080/matches/${matchId}/reject`,{
      method:"POST"
    })
    fetchPendingApproval()
   }

  useEffect(() => {
    async function fetchData() {
      const response = await fetch("http://localhost:8080/transactions")
      const data = await response.json()
      setTransaction(data || [])
    }
        fetchData()
fetchPendingApproval()
  },[])


  return (

   <div>
  <div className="dashboard">
    <h1>Reconciliation Dashboard</h1>
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Order ID</th>
          <th>Amount</th>
          <th>Mode</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        {transaction.map(t => (
          <tr key={t.id}>
            <td>{t.id}</td>
            <td>{t.order_id}</td>
            <td>₹{t.amount}</td>
            <td>{t.mode_of_payment}</td>
<td>
  <span className={`badge ${t.status_of_payment}`}>
    {t.status_of_payment}
  </span>
</td>          </tr>
        ))}
      </tbody>
    </table>
  </div>

    <div className='pendingapproval'>
       <h1>Pending Approval</h1>
       <table>
        <thead>
          <tr>
            <th>
              MathcId
            </th>
            <th>Internal Txn</th>
            <th>Gateway Txn</th>
            <th>Actions</th>
          </tr>
          
        </thead>

        <tbody>
       { pendingApproval.map(m=>(
         <tr key={m.id}>
    <td>{m.id}</td>
    <td>{m.internal_transaction_id}</td>
    <td>{m.gateway_transaction_id}</td>
    <td>
<button onClick={() => HandleApprove(m.id)}>Approve</button>
<button onClick={() => HandleReject(m.id)}>Reject</button>            </td>
    </tr>
        ))}
        
        </tbody>
       </table>
    </div>

    </div>
)



}

export default App